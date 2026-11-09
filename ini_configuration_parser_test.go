package ini_configuration_parser

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestParseBasic(t *testing.T) {
	input := `
; global comment
[section1]
key1 = value1
key2=value2
# another comment
key3 = value with spaces

[section2]
foo = bar
`
	doc, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(doc.GlobalComments) != 1 || doc.GlobalComments[0] != "global comment" {
		t.Errorf("GlobalComments = %v, want [global comment]", doc.GlobalComments)
	}

	if len(doc.Sections) != 2 {
		t.Fatalf("len(Sections) = %d, want 2", len(doc.Sections))
	}

	sec1 := doc.Sections[0]
	if sec1.Name != "section1" {
		t.Errorf("sec1.Name = %q, want %q", sec1.Name, "section1")
	}
	if len(sec1.Comments) != 1 || sec1.Comments[0] != "another comment" {
		t.Errorf("sec1.Comments = %v, want [another comment]", sec1.Comments)
	}
	wantKeys := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value with spaces",
	}
	if !reflect.DeepEqual(sec1.Keys, wantKeys) {
		t.Errorf("sec1.Keys = %v, want %v", sec1.Keys, wantKeys)
	}

	sec2 := doc.Sections[1]
	if sec2.Name != "section2" {
		t.Errorf("sec2.Name = %q, want %q", sec2.Name, "section2")
	}
	if len(sec2.Keys) != 1 || sec2.Keys["foo"] != "bar" {
		t.Errorf("sec2.Keys = %v, want map[foo:bar]", sec2.Keys)
	}
}

func TestParseEmpty(t *testing.T) {
	doc, err := Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(doc.Sections) != 0 {
		t.Errorf("len(Sections) = %d, want 0", len(doc.Sections))
	}
	if len(doc.GlobalComments) != 0 {
		t.Errorf("len(GlobalComments) = %d, want 0", len(doc.GlobalComments))
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr string
	}{
		{"empty section", "[]\n", "empty section name"},
		{"missing equals", "key\n", "expected key=value"},
		{"empty key", "= value\n", "empty key"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(strings.NewReader(tt.input))
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	doc := &Document{
		GlobalComments: []string{"global comment"},
		Sections: []*Section{
			{
				Name:     "section1",
				Comments: []string{"section comment"},
				Keys: map[string]string{
					"z": "last",
					"a": "first",
				},
			},
			{
				Name: "section2",
				Keys: map[string]string{"key": "value"},
			},
		},
	}

	var buf bytes.Buffer
	err := doc.Write(&buf)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	got := buf.String()
	want := `; global comment

[section1]
; section comment
a = first
z = last

[section2]
key = value
`
	if got != want {
		t.Errorf("Write output mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestRoundTrip(t *testing.T) {
	input := `; global
[section]
key = value
another = thing
`
	doc, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	var buf bytes.Buffer
	if err := doc.Write(&buf); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	// Re-parse the output and compare the Documents.
	doc2, err := Parse(&buf)
	if err != nil {
		t.Fatalf("Re-parse failed: %v", err)
	}

	if len(doc2.Sections) != 1 || doc2.Sections[0].Name != "section" {
		t.Fatalf("unexpected sections after round-trip: %+v", doc2.Sections)
	}
	if !reflect.DeepEqual(doc.Sections[0].Keys, doc2.Sections[0].Keys) {
		t.Errorf("keys mismatch after round-trip: %v vs %v", doc.Sections[0].Keys, doc2.Sections[0].Keys)
	}
}
