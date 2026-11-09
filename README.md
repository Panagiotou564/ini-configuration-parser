# ini-configuration-parser

A small Go library for parsing and serialising INI configuration files with sections, keys, and comments.

## Usage

```go
package main

import (
	"fmt"
	"strings"

	ini "github.com/yourname/ini-configuration-parser"
)

func main() {
	input := `
; global comment
[server]
host = localhost
port = 8080
`
	doc, err := ini.Parse(strings.NewReader(input))
	if err != nil {
		panic(err)
	}
	fmt.Println(doc.Sections[0].Name)          // server
	fmt.Println(doc.Sections[0].Keys["host"])  // localhost
}
```

## Why

Most INI parsers either support inline comments or don't, and many make it hard to round-trip a file without losing comments. This library keeps a simple model: comments are attached to the global document or to a section, and serialisation writes them back in a deterministic order. The trade-off is that inline comments (e.g. `key = value ; comment`) are not supported, because values may legitimately contain `;` or `#` (URLs, passwords). If you need inline comments, this library is not for you.

## Awkward edge

Global keys (keys before any section header) are ignored during parsing. The `Document` type only holds named sections. This is a deliberate simplification: most INI files use sections, and supporting global keys would require an additional map and complicate the API.

## API

- `Parse(r io.Reader) (*Document, error)`
- `(*Document).Write(w io.Writer) error`
- `type Document struct { GlobalComments []string; Sections []*Section }`
- `type Section struct { Name string; Keys map[string]string; Comments []string }`
