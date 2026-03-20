package gen

import (
	"strings"
	"text/template"

	"github.com/joakimcarlsson/bonk/internal/pdl"
)

// TemplateFuncs returns the template function map used by all templates.
func TemplateFuncs(
	modulePath string,
	currentDomain *pdl.Domain,
) template.FuncMap {
	return template.FuncMap{
		"comment":           comment,
		"fieldComment":      fieldComment,
		"docComment":        docComment,
		"goType":            goType,
		"resolveRef":        makeResolveRef(modulePath, currentDomain),
		"exportedName":      pdl.ExportedName,
		"unexportedName":    safeUnexportedName,
		"enumIdent":         pdl.EnumValueToIdent,
		"domainPkg":         pdl.DomainToPackage,
		"domainPrefix":      domainPrefix,
		"jsonTag":           jsonTag,
		"hasRequiredParams": hasRequiredParams,
		"requiredParams":    requiredParams,
		"optionalParams":    optionalParams,
		"hasReturns":        func(c *pdl.Command) bool { return len(c.Returns) > 0 },
		"commandMethod":     func(d *pdl.Domain, c *pdl.Command) string { return d.Name + "." + c.Name },
		"eventMethod":       func(d *pdl.Domain, e *pdl.Event) string { return d.Name + "." + e.Name },
		"toLower":           strings.ToLower,
		"isRefType":         isRefType,
		"needsPointer":      needsPointer,
		"isEnum":            func(p *pdl.Property) bool { return p.Enum != nil && len(p.Enum) > 0 },
		"isInlineEnum":      func(p *pdl.Property) bool { return p.Enum != nil && len(p.Enum) > 0 },
		"isSelfRef": func(domain string, typeID string, ref *pdl.TypeRef) bool {
			if ref == nil || ref.Items != nil {
				return false
			}
			resolved := ""
			if ref.Domain != "" {
				resolved = ref.Domain + pdl.ExportedName(ref.Name)
			} else if ref.Name != "" {
				resolved = domain + pdl.ExportedName(ref.Name)
			}
			return resolved == domain+pdl.ExportedName(typeID)
		},
	}
}

func goType(ref *pdl.TypeRef) string {
	if ref == nil {
		return "json.RawMessage"
	}

	if ref.Items != nil {
		return "[]" + goType(ref.Items)
	}

	switch ref.RawType {
	case "string":
		return "string"
	case "integer":
		return "int64"
	case "number":
		return "float64"
	case "boolean":
		return "bool"
	case "any", "object":
		return "json.RawMessage"
	case "binary":
		return "[]byte"
	}

	return ""
}

func makeResolveRef(
	modulePath string,
	currentDomain *pdl.Domain,
) func(*pdl.TypeRef) string {
	resolve := func(ref *pdl.TypeRef) string { return "" }
	resolve = func(ref *pdl.TypeRef) string {
		if ref == nil {
			return "json.RawMessage"
		}

		if ref.Items != nil {
			return "[]" + resolve(ref.Items)
		}

		if primitive := goType(ref); primitive != "" {
			return primitive
		}

		if ref.Domain != "" {
			name := pdl.ExportedName(ref.Name)
			if handWrittenTypes[ref.Name] || handWrittenTypes[name] {
				return name
			}
			return ref.Domain + name
		}

		if ref.Name != "" {
			exported := pdl.ExportedName(ref.Name)
			if handWrittenTypes[ref.Name] || handWrittenTypes[exported] {
				return exported
			}
			return currentDomain.Name + exported
		}

		return "json.RawMessage"
	}
	return resolve
}

func domainPrefix(domain, name string) string {
	if handWrittenTypes[name] {
		return name
	}
	return domain + name
}

func jsonTag(name string, optional bool) string {
	if optional {
		return name + ",omitempty"
	}
	return name
}

func hasRequiredParams(params []*pdl.Property) bool {
	for _, p := range params {
		if !p.Optional {
			return true
		}
	}
	return false
}

func requiredParams(params []*pdl.Property) []*pdl.Property {
	var result []*pdl.Property
	for _, p := range params {
		if !p.Optional {
			result = append(result, p)
		}
	}
	return result
}

func optionalParams(params []*pdl.Property) []*pdl.Property {
	var result []*pdl.Property
	for _, p := range params {
		if p.Optional {
			result = append(result, p)
		}
	}
	return result
}

func isRefType(ref *pdl.TypeRef) bool {
	if ref == nil {
		return false
	}
	return ref.Name != "" || ref.Domain != ""
}

func comment(name, desc string) string {
	if desc == "" {
		return ""
	}
	lines := strings.Split(desc, "\n")
	var b strings.Builder
	b.WriteString("// " + name + " " + lines[0])
	for _, line := range lines[1:] {
		b.WriteString("\n// " + line)
	}
	return b.String()
}

func fieldComment(name, desc string) string {
	if desc == "" {
		return ""
	}
	lines := strings.Split(desc, "\n")
	var b strings.Builder
	b.WriteString("\t// " + name + " " + lines[0])
	for _, line := range lines[1:] {
		b.WriteString("\n\t// " + line)
	}
	return b.String()
}

var goKeywords = map[string]string{
	"break":       "brk",
	"case":        "cas",
	"chan":        "ch",
	"const":       "cnst",
	"continue":    "cont",
	"default":     "def",
	"defer":       "dfr",
	"else":        "els",
	"fallthrough": "ft",
	"for":         "fr",
	"func":        "fn",
	"go":          "g",
	"goto":        "gt",
	"if":          "cond",
	"import":      "imp",
	"interface":   "iface",
	"map":         "mp",
	"package":     "pkg",
	"range":       "rng",
	"return":      "ret",
	"select":      "sel",
	"struct":      "st",
	"switch":      "sw",
	"type":        "typ",
	"var":         "vr",
}

func safeUnexportedName(name string) string {
	result := pdl.UnexportedName(name)
	if replacement, ok := goKeywords[result]; ok {
		return replacement
	}
	return result
}

func docComment(desc string) string {
	if desc == "" {
		return ""
	}
	lines := strings.Split(desc, "\n")
	var b strings.Builder
	for _, line := range lines {
		b.WriteString("// " + line + "\n")
	}
	return b.String()
}

func needsPointer(p *pdl.Property) bool {
	if !p.Optional {
		return false
	}
	if p.Ref == nil {
		return false
	}
	if p.Ref.Items != nil {
		return false
	}
	switch p.Ref.RawType {
	case "string", "any", "object", "binary":
		return false
	case "integer", "number", "boolean":
		return false
	}
	return false
}
