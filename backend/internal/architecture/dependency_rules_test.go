package architecture

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

const moduleImportPrefix = "computility-ops/backend/internal/modules/"

func TestModuleLayeringRules(t *testing.T) {
	root := filepath.Join("..", "modules")
	files, err := filepath.Glob(filepath.Join(root, "*", "*", "*.go"))
	if err != nil {
		t.Fatalf("glob failed: %v", err)
	}

	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		layer := filepath.Base(filepath.Dir(path))
		module := filepath.Base(filepath.Dir(filepath.Dir(path)))
		imports := parseImports(t, path)
		for _, imp := range imports {
			if !strings.HasPrefix(imp, moduleImportPrefix) {
				continue
			}
			trimmed := strings.TrimPrefix(imp, moduleImportPrefix)
			parts := strings.Split(trimmed, "/")
			if len(parts) < 2 {
				continue
			}
			targetModule := parts[0]
			targetLayer := parts[1]

			if layer == "domain" && targetLayer != "domain" {
				t.Fatalf("domain layer cannot import non-domain: %s imports %s", path, imp)
			}
			if layer == "application" && targetLayer == "infrastructure" {
				t.Fatalf("application layer cannot import infrastructure: %s imports %s", path, imp)
			}
			if layer == "api" && targetLayer == "infrastructure" {
				t.Fatalf("api layer cannot import infrastructure: %s imports %s", path, imp)
			}
			if targetModule != module && targetLayer == "infrastructure" {
				t.Fatalf("cross-module infrastructure import forbidden: %s imports %s", path, imp)
			}
		}
	}
}

func parseImports(t *testing.T, path string) []string {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse imports failed (%s): %v", path, err)
	}
	out := make([]string, 0, len(f.Imports))
	for _, imp := range f.Imports {
		out = append(out, strings.Trim(imp.Path.Value, "\""))
	}
	return out
}

func _ensureASTUsed(_ ast.Node) {}
