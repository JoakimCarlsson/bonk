package gen

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/joakimcarlsson/bonk/internal/pdl"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// Generator produces Go source files from a parsed PDL protocol.
type Generator struct {
	Proto      *pdl.Protocol
	OutputDir  string
	ModulePath string
}

type templateData struct {
	Domain  *pdl.Domain
	Imports []string
	Types   []*pdl.Type
}

// Generate writes Go source files for all domains.
// All domain code is generated into a single flat package (the output dir).
// Each domain gets its own set of files prefixed with the domain name.
func (g *Generator) Generate() error {
	for _, domain := range g.Proto.Domains {
		expandInlineEnums(domain)
	}

	for _, domain := range g.Proto.Domains {
		if err := g.generateDomain(domain); err != nil {
			return fmt.Errorf("domain %s: %w", domain.Name, err)
		}
	}
	return nil
}

func runGoimports(path string) error {
	goimports, err := exec.LookPath("goimports")
	if err != nil {
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			home, _ := os.UserHomeDir()
			gopath = filepath.Join(home, "go")
		}
		goimports = filepath.Join(gopath, "bin", "goimports")
	}

	cmd := exec.Command(goimports, "-w", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("goimports %s: %s: %w", path, out, err)
	}
	return nil
}

func expandInlineEnums(domain *pdl.Domain) {
	var extraTypes []*pdl.Type

	processProps := func(parentName string, props []*pdl.Property) {
		for _, p := range props {
			if len(p.Enum) == 0 {
				continue
			}
			typeName := parentName + pdl.ExportedName(p.Name)
			extraTypes = append(extraTypes, &pdl.Type{
				ID:          typeName,
				Description: p.Description,
				BaseType:    "string",
				Enum:        p.Enum,
			})
			p.Ref = &pdl.TypeRef{Name: typeName}
			p.Enum = nil
		}
	}

	for _, t := range domain.Types {
		processProps(pdl.ExportedName(t.ID), t.Properties)
	}
	for _, c := range domain.Commands {
		processProps(pdl.ExportedName(c.Name), c.Parameters)
		processProps(pdl.ExportedName(c.Name)+"Return", c.Returns)
	}
	for _, e := range domain.Events {
		processProps(pdl.ExportedName(e.Name), e.Parameters)
	}

	domain.Types = append(domain.Types, extraTypes...)
}

func filterHandWrittenTypes(types []*pdl.Type) []*pdl.Type {
	var result []*pdl.Type
	for _, t := range types {
		exported := pdl.ExportedName(t.ID)
		if handWrittenTypes[t.ID] || handWrittenTypes[exported] {
			continue
		}
		result = append(result, t)
	}
	return result
}

var handWrittenTypes = map[string]bool{
	"FrameID":          true,
	"FrameId":          true,
	"LoaderID":         true,
	"LoaderId":         true,
	"RequestID":        true,
	"RequestId":        true,
	"RemoteObjectID":   true,
	"RemoteObjectId":   true,
	"BrowserContextID": true,
	"BrowserContextId": true,
	"NodeID":           true,
	"NodeId":           true,
	"BackendNodeID":    true,
	"BackendNodeId":    true,
	"TimeSinceEpoch":   true,
	"MonotonicTime":    true,
}

func (g *Generator) generateDomain(domain *pdl.Domain) error {
	if err := os.MkdirAll(g.OutputDir, 0o755); err != nil {
		return err
	}

	prefix := strings.ToLower(domain.Name)

	nonHandwritten := filterHandWrittenTypes(domain.Types)
	if len(nonHandwritten) > 0 {
		typeDomain := &pdl.Domain{
			Name:        domain.Name,
			Description: domain.Description,
			Types:       nonHandwritten,
		}
		if err := g.writeTemplate("templates/types.go.tmpl", g.OutputDir, prefix+"_types.go", typeDomain); err != nil {
			return err
		}
	}

	if len(domain.Commands) > 0 {
		if err := g.writeTemplate("templates/domain.go.tmpl", g.OutputDir, prefix+"_commands.go", domain); err != nil {
			return err
		}
	}

	if len(domain.Events) > 0 {
		if err := g.writeTemplate("templates/events.go.tmpl", g.OutputDir, prefix+"_events.go", domain); err != nil {
			return err
		}
	}

	return nil
}

func (g *Generator) writeTemplate(
	tmplName, dir, filename string,
	domain *pdl.Domain,
) error {
	content, err := templateFS.ReadFile(tmplName)
	if err != nil {
		return fmt.Errorf("read template %s: %w", tmplName, err)
	}

	funcs := TemplateFuncs(g.ModulePath, domain)
	tmpl, err := template.New(tmplName).Funcs(funcs).Parse(string(content))
	if err != nil {
		return fmt.Errorf("parse template %s: %w", tmplName, err)
	}

	data := templateData{
		Domain: domain,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("execute template %s: %w", tmplName, err)
	}

	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	return runGoimports(path)
}
