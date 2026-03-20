package pdl

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type parser struct {
	scanner          *bufio.Scanner
	line             int
	proto            *Protocol
	domain           *Domain
	item             interface{}
	section          string
	inlineEnumProp   *Property
	inlineEnumParent string
}

// Parse reads a PDL file and returns the parsed protocol.
func Parse(r io.Reader) (*Protocol, error) {
	p := &parser{
		scanner: bufio.NewScanner(r),
		proto:   &Protocol{},
	}
	return p.parse()
}

// ParseString parses PDL content from a string.
func ParseString(s string) (*Protocol, error) {
	return Parse(strings.NewReader(s))
}

func stripModifiers(line string) (experimental, deprecated bool, rest string) {
	for {
		if strings.HasPrefix(line, "experimental ") {
			experimental = true
			line = strings.TrimPrefix(line, "experimental ")
		} else if strings.HasPrefix(line, "deprecated ") {
			deprecated = true
			line = strings.TrimPrefix(line, "deprecated ")
		} else {
			break
		}
	}
	return experimental, deprecated, line
}

func containsKeyword(line, keyword string) bool {
	words := strings.Fields(line)
	for i, w := range words {
		if w == keyword {
			return i <= 2
		}
		if w != "experimental" && w != "deprecated" {
			return false
		}
	}
	return false
}

func (p *parser) resetInlineEnum() {
	p.inlineEnumProp = nil
	p.inlineEnumParent = ""
}

func (p *parser) errorf(format string, args ...any) error {
	return fmt.Errorf(
		"pdl: line %d: "+format,
		append([]any{p.line}, args...)...)
}

func (p *parser) parse() (*Protocol, error) {
	var desc strings.Builder

	for p.scanner.Scan() {
		p.line++
		raw := p.scanner.Text()
		trimmed := strings.TrimSpace(raw)

		if trimmed == "" {
			if desc.Len() > 0 {
				desc.Reset()
			}
			continue
		}

		if strings.HasPrefix(trimmed, "#") {
			comment := strings.TrimPrefix(trimmed, "#")
			comment = strings.TrimPrefix(comment, " ")
			if desc.Len() > 0 {
				desc.WriteByte('\n')
			}
			desc.WriteString(comment)
			continue
		}

		indent := len(raw) - len(strings.TrimLeft(raw, " "))
		description := desc.String()
		desc.Reset()

		if err := p.parseLine(trimmed, indent, description); err != nil {
			return nil, err
		}
	}

	if err := p.scanner.Err(); err != nil {
		return nil, fmt.Errorf("pdl: scan error: %w", err)
	}

	return p.proto, nil
}

func (p *parser) parseLine(line string, indent int, description string) error {
	switch {
	case strings.HasPrefix(line, "version"):
		return p.parseVersion()

	case strings.HasPrefix(line, "include "):
		p.proto.Includes = append(
			p.proto.Includes,
			strings.TrimPrefix(line, "include "),
		)
		return nil

	case strings.HasPrefix(line, "domain ") || strings.HasPrefix(line, "experimental domain ") || strings.HasPrefix(line, "deprecated domain "):
		return p.parseDomain(line, description)

	case strings.HasPrefix(line, "depends on "):
		if p.domain == nil {
			return p.errorf("depends on outside domain")
		}
		p.domain.Dependencies = append(
			p.domain.Dependencies,
			strings.TrimPrefix(line, "depends on "),
		)
		return nil

	case containsKeyword(line, "type"):
		return p.parseType(line, description)

	case containsKeyword(line, "command"):
		return p.parseCommand(line, description)

	case containsKeyword(line, "event"):
		return p.parseEvent(line, description)

	case line == "properties":
		p.section = "properties"
		p.resetInlineEnum()
		return nil

	case line == "parameters":
		p.section = "parameters"
		p.resetInlineEnum()
		return nil

	case line == "returns":
		p.section = "returns"
		p.resetInlineEnum()
		return nil

	case line == "enum":
		p.section = "enum"
		p.resetInlineEnum()
		return nil

	case strings.HasPrefix(line, "redirect "):
		if cmd, ok := p.item.(*Command); ok {
			cmd.Redirect = strings.TrimPrefix(line, "redirect ")
		}
		return nil

	default:
		return p.parseField(line, indent, description)
	}
}

func (p *parser) parseVersion() error {
	for p.scanner.Scan() {
		p.line++
		trimmed := strings.TrimSpace(p.scanner.Text())

		if trimmed == "" {
			break
		}

		parts := strings.Fields(trimmed)
		if len(parts) != 2 {
			return p.errorf("invalid version line: %q", trimmed)
		}

		switch parts[0] {
		case "major":
			p.proto.Version.Major = parts[1]
		case "minor":
			p.proto.Version.Minor = parts[1]
		}
	}
	return nil
}

func (p *parser) parseDomain(line, description string) error {
	experimental, deprecated, line := stripModifiers(line)
	name := strings.TrimPrefix(line, "domain ")
	p.domain = &Domain{
		Name:         name,
		Description:  description,
		Experimental: experimental,
		Deprecated:   deprecated,
	}
	p.proto.Domains = append(p.proto.Domains, p.domain)
	p.item = nil
	p.section = ""
	p.resetInlineEnum()
	return nil
}

func (p *parser) parseType(line, description string) error {
	if p.domain == nil {
		return p.errorf("type outside domain")
	}

	experimental, deprecated, line := stripModifiers(line)
	line = strings.TrimPrefix(line, "type ")

	parts := strings.Fields(line)
	if len(parts) < 3 || parts[1] != "extends" {
		return p.errorf("invalid type declaration: %q", line)
	}

	t := &Type{
		ID:           parts[0],
		Description:  description,
		Experimental: experimental,
		Deprecated:   deprecated,
		BaseType:     parts[2],
	}

	if t.BaseType == "array" && len(parts) >= 5 && parts[3] == "of" {
		t.Items = parseTypeRef(parts[4])
	}

	p.domain.Types = append(p.domain.Types, t)
	p.item = t
	p.section = ""
	p.resetInlineEnum()
	return nil
}

func (p *parser) parseCommand(line, description string) error {
	if p.domain == nil {
		return p.errorf("command outside domain")
	}

	experimental, deprecated, line := stripModifiers(line)
	name := strings.TrimPrefix(line, "command")
	name = strings.TrimSpace(name)

	cmd := &Command{
		Name:         name,
		Description:  description,
		Experimental: experimental,
		Deprecated:   deprecated,
	}

	p.domain.Commands = append(p.domain.Commands, cmd)
	p.item = cmd
	p.section = ""
	p.resetInlineEnum()
	return nil
}

func (p *parser) parseEvent(line, description string) error {
	if p.domain == nil {
		return p.errorf("event outside domain")
	}

	experimental, deprecated, line := stripModifiers(line)
	name := strings.TrimPrefix(line, "event")
	name = strings.TrimSpace(name)

	ev := &Event{
		Name:         name,
		Description:  description,
		Experimental: experimental,
		Deprecated:   deprecated,
	}

	p.domain.Events = append(p.domain.Events, ev)
	p.item = ev
	p.section = ""
	p.resetInlineEnum()
	return nil
}

func (p *parser) parseField(line string, indent int, description string) error {
	if p.section == "enum" {
		return p.addEnumValue(line)
	}

	if p.section == "inline-enum" {
		if !strings.Contains(line, " ") && !strings.HasPrefix(line, "optional ") &&
			!strings.HasPrefix(line, "experimental ") &&
			!strings.HasPrefix(line, "deprecated ") {
			p.inlineEnumProp.Enum = append(p.inlineEnumProp.Enum, line)
			return nil
		}
		p.section = p.inlineEnumParent
		p.inlineEnumProp = nil
	}

	if p.section == "" {
		return nil
	}

	prop, err := parseProperty(line, description)
	if err != nil {
		return p.errorf("%v", err)
	}

	switch p.section {
	case "properties":
		if t, ok := p.item.(*Type); ok {
			t.Properties = append(t.Properties, prop)
		}
	case "parameters":
		switch item := p.item.(type) {
		case *Command:
			item.Parameters = append(item.Parameters, prop)
		case *Event:
			item.Parameters = append(item.Parameters, prop)
		}
	case "returns":
		if cmd, ok := p.item.(*Command); ok {
			cmd.Returns = append(cmd.Returns, prop)
		}
	}

	if prop.Enum != nil {
		p.inlineEnumParent = p.section
		p.section = "inline-enum"
		p.inlineEnumProp = prop
	}

	return nil
}

func (p *parser) addEnumValue(line string) error {
	if p.inlineEnumProp != nil {
		p.inlineEnumProp.Enum = append(p.inlineEnumProp.Enum, line)
		return nil
	}
	if t, ok := p.item.(*Type); ok {
		t.Enum = append(t.Enum, line)
		return nil
	}
	return p.errorf("enum value outside type or property")
}

func parseProperty(line, description string) (*Property, error) {
	prop := &Property{Description: description}

	for {
		if strings.HasPrefix(line, "deprecated ") {
			prop.Deprecated = true
			line = strings.TrimPrefix(line, "deprecated ")
		} else if strings.HasPrefix(line, "experimental ") {
			prop.Experimental = true
			line = strings.TrimPrefix(line, "experimental ")
		} else if strings.HasPrefix(line, "optional ") {
			prop.Optional = true
			line = strings.TrimPrefix(line, "optional ")
		} else {
			break
		}
	}

	if strings.HasPrefix(line, "enum ") {
		prop.Name = strings.TrimPrefix(line, "enum ")
		prop.Ref = &TypeRef{RawType: "string"}
		prop.Enum = []string{}
		return prop, nil
	}

	if strings.HasPrefix(line, "array of ") {
		rest := strings.TrimPrefix(line, "array of ")
		parts := strings.Fields(rest)
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid array property: %q", line)
		}
		prop.Ref = &TypeRef{Items: parseTypeRef(parts[0])}
		prop.Name = parts[1]
		return prop, nil
	}

	parts := strings.Fields(line)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid property: %q", line)
	}

	prop.Ref = parseTypeRef(parts[0])
	prop.Name = parts[1]
	return prop, nil
}

func parseTypeRef(s string) *TypeRef {
	if i := strings.IndexByte(s, '.'); i >= 0 {
		return &TypeRef{
			Domain: s[:i],
			Name:   s[i+1:],
		}
	}

	switch s {
	case "string", "integer", "number", "boolean", "any", "binary", "object":
		return &TypeRef{RawType: s}
	}

	return &TypeRef{Name: s}
}
