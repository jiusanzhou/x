// Package gen provides code generation for talk endpoints.
package gen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"go.zoe.im/x/talk"
	"go.zoe.im/x/talk/extract"
)

// Generator generates endpoint registration code from Go interfaces.
type Generator struct {
	PackageName string
	TypeName    string
	OutputFile  string
}

// MethodInfo holds parsed method information for code generation.
type MethodInfo struct {
	Name         string
	Path         string
	HTTPMethod   string
	StreamMode   talk.StreamMode
	RequestType  string
	ResponseType string
	HasRequest   bool
	HasResponse  bool
	Comments     []string
}

// InterfaceInfo holds parsed interface information.
type InterfaceInfo struct {
	PackageName string
	TypeName    string
	Methods     []MethodInfo
	Imports     []string
}

// Generate parses the source file and generates endpoint code.
func (g *Generator) Generate(sourceFile string) error {
	info, err := g.parseInterface(sourceFile)
	if err != nil {
		return err
	}

	code, err := g.generateCode(info)
	if err != nil {
		return err
	}

	outputFile := g.OutputFile
	if outputFile == "" {
		ext := filepath.Ext(sourceFile)
		outputFile = strings.TrimSuffix(sourceFile, ext) + "_talk" + ext
	}

	return os.WriteFile(outputFile, code, 0644)
}

func (g *Generator) parseInterface(sourceFile string) (*InterfaceInfo, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, sourceFile, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse file: %w", err)
	}

	info := &InterfaceInfo{
		PackageName: f.Name.Name,
		TypeName:    g.TypeName,
	}

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != g.TypeName {
				continue
			}

			ifaceType, ok := typeSpec.Type.(*ast.InterfaceType)
			if !ok {
				return nil, fmt.Errorf("%s is not an interface", g.TypeName)
			}

			info.Methods = g.parseMethods(ifaceType, f.Comments)
		}
	}

	if len(info.Methods) == 0 {
		return nil, fmt.Errorf("no methods found in interface %s", g.TypeName)
	}

	return info, nil
}

func (g *Generator) parseMethods(iface *ast.InterfaceType, comments []*ast.CommentGroup) []MethodInfo {
	var methods []MethodInfo

	for _, method := range iface.Methods.List {
		if len(method.Names) == 0 {
			continue
		}

		funcType, ok := method.Type.(*ast.FuncType)
		if !ok {
			continue
		}

		methodName := method.Names[0].Name
		mi := MethodInfo{
			Name: methodName,
		}

		// Extract comments
		if method.Doc != nil {
			for _, c := range method.Doc.List {
				mi.Comments = append(mi.Comments, c.Text)
			}
		}

		// Parse @talk annotation
		ann := extract.ParseAnnotations(mi.Comments)
		if ann != nil {
			mi.Path = ann.Path
			mi.HTTPMethod = ann.Method
			mi.StreamMode = ann.StreamMode
		}

		// Derive from method name if not specified
		if mi.Path == "" || mi.HTTPMethod == "" {
			httpMethod, path := deriveMethodAndPath(methodName, funcType)
			if mi.Path == "" {
				mi.Path = path
			}
			if mi.HTTPMethod == "" {
				mi.HTTPMethod = httpMethod
			}
		}

		// Parse request/response types
		mi.RequestType, mi.HasRequest = parseRequestType(funcType)
		mi.ResponseType, mi.HasResponse = parseResponseType(funcType)

		// Detect stream mode from signature if not specified
		if mi.StreamMode == talk.StreamNone {
			mi.StreamMode = detectStreamMode(funcType)
		}

		methods = append(methods, mi)
	}

	return methods
}

func deriveMethodAndPath(name string, funcType *ast.FuncType) (httpMethod, path string) {
	resource := extractResource(name)
	hasIDParam := hasSimpleParam(funcType)

	switch {
	case strings.HasPrefix(name, "Get"):
		httpMethod = "GET"
		if hasIDParam {
			path = "/" + resource + "/{id}"
		} else {
			path = "/" + resource
		}
	case strings.HasPrefix(name, "List"):
		httpMethod = "GET"
		path = "/" + resource
	case strings.HasPrefix(name, "Create"):
		httpMethod = "POST"
		path = "/" + resource
	case strings.HasPrefix(name, "Update"):
		httpMethod = "PUT"
		path = "/" + resource + "/{id}"
	case strings.HasPrefix(name, "Delete"):
		httpMethod = "DELETE"
		path = "/" + resource + "/{id}"
	case strings.HasPrefix(name, "Watch"):
		httpMethod = "GET"
		path = "/" + resource + "/watch"
	default:
		httpMethod = "POST"
		path = "/" + toKebabCase(name)
	}

	return
}

func extractResource(methodName string) string {
	prefixes := []string{"Get", "List", "Create", "Update", "Delete", "Watch"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(methodName, prefix) {
			resource := methodName[len(prefix):]
			return strings.ToLower(resource)
		}
	}
	return strings.ToLower(methodName)
}

func toKebabCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteRune('-')
			}
			result.WriteRune(r + 32)
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func hasSimpleParam(funcType *ast.FuncType) bool {
	if funcType == nil || funcType.Params == nil || len(funcType.Params.List) < 2 {
		return false
	}
	// Skip context, check second param
	for i, param := range funcType.Params.List {
		if i == 0 {
			continue
		}
		if ident, ok := param.Type.(*ast.Ident); ok {
			if ident.Name == "string" || ident.Name == "int" || ident.Name == "int64" {
				return true
			}
		}
	}
	return false
}

func parseRequestType(funcType *ast.FuncType) (string, bool) {
	if funcType.Params == nil || len(funcType.Params.List) < 2 {
		return "", false
	}

	// Skip context.Context (first param)
	for i, param := range funcType.Params.List {
		if i == 0 {
			continue
		}
		return typeToString(param.Type), true
	}
	return "", false
}

func parseResponseType(funcType *ast.FuncType) (string, bool) {
	if funcType.Results == nil || len(funcType.Results.List) == 0 {
		return "", false
	}

	// First result is response, last is error
	if len(funcType.Results.List) > 1 {
		return typeToString(funcType.Results.List[0].Type), true
	}
	return "", false
}

func detectStreamMode(funcType *ast.FuncType) talk.StreamMode {
	hasInputChan := false
	hasOutputChan := false

	if funcType.Params != nil {
		for _, param := range funcType.Params.List {
			if _, ok := param.Type.(*ast.ChanType); ok {
				hasInputChan = true
			}
		}
	}

	if funcType.Results != nil {
		for _, result := range funcType.Results.List {
			if _, ok := result.Type.(*ast.ChanType); ok {
				hasOutputChan = true
			}
		}
	}

	switch {
	case hasInputChan && hasOutputChan:
		return talk.StreamBidirect
	case hasInputChan:
		return talk.StreamClientSide
	case hasOutputChan:
		return talk.StreamServerSide
	default:
		return talk.StreamNone
	}
}

func typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeToString(t.X)
	case *ast.SelectorExpr:
		return typeToString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + typeToString(t.Elt)
	case *ast.ChanType:
		return "<-chan " + typeToString(t.Value)
	case *ast.MapType:
		return "map[" + typeToString(t.Key) + "]" + typeToString(t.Value)
	default:
		return "any"
	}
}

func (g *Generator) generateCode(info *InterfaceInfo) ([]byte, error) {
	tmpl, err := template.New("endpoints").Parse(endpointTemplate)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, info); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

const endpointTemplate = `// Code generated by go.zoe.im/x/talk/gen. DO NOT EDIT.

package {{.PackageName}}

import (
	"context"

	"go.zoe.im/x/talk"
)

// {{.TypeName}}Endpoints returns the endpoints for {{.TypeName}}.
func {{.TypeName}}Endpoints(svc {{.TypeName}}) []*talk.Endpoint {
	return []*talk.Endpoint{
{{- range .Methods}}
		{
			Name:       "{{.Name}}",
			Path:       "{{.Path}}",
			Method:     "{{.HTTPMethod}}",
			StreamMode: {{.StreamMode}},
			Handler: func(ctx context.Context, req any) (any, error) {
{{- if .HasRequest}}
				r, _ := req.({{.RequestType}})
				return svc.{{.Name}}(ctx, r)
{{- else}}
				return svc.{{.Name}}(ctx)
{{- end}}
			},
		},
{{- end}}
	}
}
`
