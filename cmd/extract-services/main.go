// Command extract-services walks the Go service layer in services/ and emits
// a JSON manifest of every exported method on every *Service receiver.
//
// The output (services-contract.json) is the single source of truth for the
// surface that frontend adapters (Wails desktop bindings, web HTTP/IndexedDB
// adapters) must implement. Run via `make contract`. CI fails when the diff
// is non-empty, forcing service authors to commit the new contract.
package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Method struct {
	Name    string   `json:"name"`
	Params  []string `json:"params"`
	Returns []string `json:"returns"`
}

type Service struct {
	File    string   `json:"file"`
	Methods []Method `json:"methods"`
}

func main() {
	servicesDir := "services"
	outPath := "services-contract.json"
	if len(os.Args) > 1 {
		servicesDir = os.Args[1]
	}
	if len(os.Args) > 2 {
		outPath = os.Args[2]
	}

	services := map[string]*Service{}

	entries, err := os.ReadDir(servicesDir)
	if err != nil {
		fail("read services dir: %v", err)
	}

	fset := token.NewFileSet()
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		if strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		path := filepath.Join(servicesDir, e.Name())
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			fail("parse %s: %v", path, err)
		}
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
				continue
			}
			if !ast.IsExported(fn.Name.Name) {
				continue
			}
			recvType := receiverTypeName(fn.Recv.List[0].Type)
			if !strings.HasSuffix(recvType, "Service") {
				continue
			}

			svc, ok := services[recvType]
			if !ok {
				svc = &Service{File: filepath.ToSlash(path)}
				services[recvType] = svc
			}
			svc.Methods = append(svc.Methods, Method{
				Name:    fn.Name.Name,
				Params:  formatFields(fn.Type.Params),
				Returns: formatFields(fn.Type.Results),
			})
		}
	}

	for _, svc := range services {
		sort.Slice(svc.Methods, func(i, j int) bool { return svc.Methods[i].Name < svc.Methods[j].Name })
	}

	out, err := os.Create(outPath)
	if err != nil {
		fail("create %s: %v", outPath, err)
	}
	defer out.Close()
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(services); err != nil {
		fail("encode: %v", err)
	}

	fmt.Printf("wrote %s (%d services)\n", outPath, len(services))
}

func receiverTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return receiverTypeName(t.X)
	case *ast.Ident:
		return t.Name
	}
	return ""
}

func formatFields(fl *ast.FieldList) []string {
	if fl == nil {
		return nil
	}
	var out []string
	for _, f := range fl.List {
		typeStr := exprString(f.Type)
		if len(f.Names) == 0 {
			out = append(out, typeStr)
			continue
		}
		for _, n := range f.Names {
			out = append(out, n.Name+":"+typeStr)
		}
	}
	return out
}

func exprString(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprString(t.X)
	case *ast.SelectorExpr:
		return exprString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + exprString(t.Elt)
	case *ast.MapType:
		return "map[" + exprString(t.Key) + "]" + exprString(t.Value)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.Ellipsis:
		return "..." + exprString(t.Elt)
	case *ast.FuncType:
		return "func"
	case *ast.ChanType:
		return "chan " + exprString(t.Value)
	}
	return "?"
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
