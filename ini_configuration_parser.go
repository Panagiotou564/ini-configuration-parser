package ini_configuration_parser

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"
)

// Section represents a single INI section with its keys and comments.
type Section struct {
	Name     string
	Keys     map[string]string
	Comments []string
}

// Document represents a parsed INI file.
type Document struct {
	GlobalComments []string
	Sections       []*Section
}

// Parse reads an INI configuration from r and returns a Document.
//
// The parser is intentionally simple: it recognises section headers of the form
// [name], key-value pairs separated by '=' (with optional surrounding whitespace),
// and comments starting with ';' or '#'. Blank lines are ignored. Inline comments
// are NOT supported because they conflict with values that legitimately contain
// ';' or '#' (e.g. URLs). This is a deliberate trade-off: simpler and more robust
// for configuration files that need to store such values.
func Parse(r io.Reader) (*Document, error) {
	doc := &Document{}
	scanner := bufio.NewScanner(r)
	var currentSection *Section
	inGlobal := true
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Comment line
		if strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			comment := strings.TrimSpace(line[1:])
			if inGlobal {
				doc.GlobalComments = append(doc.GlobalComments, comment)
			} else {
				currentSection.Comments = append(currentSection.Comments, comment)
			}
			continue
		}

		// Section header
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			name := strings.TrimSpace(line[1 : len(line)-1])
			if name == "" {
				return nil, fmt.Errorf("line %d: empty section name", lineNum)
			}
			currentSection = &Section{
				Name: name,
				Keys: make(map[string]string),
			}
			doc.Sections = append(doc.Sections, currentSection)
			inGlobal = false
			continue
		}

		// Key-value pair
		eq := strings.Index(line, "=")
		if eq == -1 {
			return nil, fmt.Errorf("line %d: expected key=value or [section], got %q", lineNum, line)
		}
		key := strings.TrimSpace(line[:eq])
		value := strings.TrimSpace(line[eq+1:])
		if key == "" {
			return nil, fmt.Errorf("line %d: empty key", lineNum)
		}
		if inGlobal {
			// Global keys are not stored; they are ignored. This is a design choice:
			// the Document type only holds named sections. If global keys are needed,
			// they can be represented as a special section, but we keep it simple.
			continue
		}
		currentSection.Keys[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return doc, nil
}

// Write serialises the Document to w in INI format.
//
// The output is deterministic: global comments are written first, then sections
// in the order they appear in the Document. Keys within a section are sorted
// alphabetically to ensure stable output across runs. This makes the serialised
// form suitable for diffing and version control.
func (d *Document) Write(w io.Writer) error {
	for _, c := range d.GlobalComments {
		if _, err := fmt.Fprintf(w, "; %s\n", c); err != nil {
			return err
		}
	}
	if len(d.GlobalComments) > 0 {
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}

	for i, sec := range d.Sections {
		if i > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "[%s]\n", sec.Name); err != nil {
			return err
		}
		for _, c := range sec.Comments {
			if _, err := fmt.Fprintf(w, "; %s\n", c); err != nil {
				return err
			}
		}
		keys := make([]string, 0, len(sec.Keys))
		for k := range sec.Keys {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if _, err := fmt.Fprintf(w, "%s = %s\n", k, sec.Keys[k]); err != nil {
				return err
			}
		}
	}
	return nil
}
