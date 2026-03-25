// Package pdl provides parsing for the Protocol Description Language.
package pdl

// Protocol is the root of a parsed PDL file.
type Protocol struct {
	Version  Version
	Domains  []*Domain
	Includes []string
}

// Version holds the protocol version.
type Version struct {
	Major string
	Minor string
}

// Domain represents a CDP domain such as "Page" or "Network".
type Domain struct {
	Name         string
	Description  string
	Experimental bool
	Deprecated   bool
	Dependencies []string
	Types        []*Type
	Commands     []*Command
	Events       []*Event
}

// Type represents a CDP type definition.
type Type struct {
	ID           string
	Description  string
	Experimental bool
	Deprecated   bool
	BaseType     string
	Enum         []string
	Properties   []*Property
	Items        *TypeRef
}

// Command represents a CDP command.
type Command struct {
	Name         string
	Description  string
	Experimental bool
	Deprecated   bool
	Parameters   []*Property
	Returns      []*Property
	Redirect     string
}

// Event represents a CDP event.
type Event struct {
	Name         string
	Description  string
	Experimental bool
	Deprecated   bool
	Parameters   []*Property
}

// Property represents a property, parameter, or return value.
type Property struct {
	Name         string
	Description  string
	Ref          *TypeRef
	Optional     bool
	Experimental bool
	Deprecated   bool
	Enum         []string
}

// TypeRef references a type, either a primitive or a named type from a domain.
type TypeRef struct {
	Domain  string
	Name    string
	RawType string
	Items   *TypeRef
}

// IsPrimitive reports whether the TypeRef refers to a primitive type.
func (r *TypeRef) IsPrimitive() bool {
	switch r.RawType {
	case "string", "integer", "number", "boolean", "any", "binary", "object":
		return r.Name == "" && r.Domain == ""
	}
	return false
}

// IsArray reports whether the TypeRef refers to an array type.
func (r *TypeRef) IsArray() bool {
	return r.Items != nil
}
