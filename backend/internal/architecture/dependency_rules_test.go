package architecture

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

const (
	moduleImportPrefix = "computility-ops/backend/internal/modules/"
	sharedImportPrefix = "computility-ops/backend/internal/shared/"
)

var sharedAllowlist = map[string]struct{}{
	"kernel": {},
}

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
			if strings.HasPrefix(imp, sharedImportPrefix) {
				enforceSharedAllowlist(t, path, imp)
				continue
			}
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

			enforceLayerRules(t, path, layer, module, targetModule, targetLayer, imp)
		}
	}
}

func enforceLayerRules(t *testing.T, path, layer, module, targetModule, targetLayer, imp string) {
	t.Helper()

	if layer == "domain" {
		if targetLayer != "domain" {
			t.Fatalf("domain layer cannot import non-domain: %s imports %s", path, imp)
		}
		if targetModule != module {
			t.Fatalf("domain layer cannot import other modules: %s imports %s", path, imp)
		}
	}

	if layer == "application" {
		if targetLayer == "api" || targetLayer == "infrastructure" {
			t.Fatalf("application layer cannot import api/infrastructure: %s imports %s", path, imp)
		}
		if targetModule != module {
			t.Fatalf("application layer cannot import other modules directly: %s imports %s", path, imp)
		}
	}

	if layer == "api" {
		if targetModule != module {
			t.Fatalf("api layer cannot import other modules directly: %s imports %s", path, imp)
		}
		if targetLayer != "application" && targetLayer != "domain" {
			t.Fatalf("api layer may only import application/domain: %s imports %s", path, imp)
		}
	}

	if layer == "infrastructure" {
		if targetLayer == "api" {
			t.Fatalf("infrastructure layer cannot import api: %s imports %s", path, imp)
		}
		if targetModule != module && targetLayer != "domain" {
			t.Fatalf("cross-module imports allowed only to domain: %s imports %s", path, imp)
		}
	}

	if targetModule != module && targetLayer == "infrastructure" {
		t.Fatalf("cross-module infrastructure import forbidden: %s imports %s", path, imp)
	}
}

func enforceSharedAllowlist(t *testing.T, path, imp string) {
	t.Helper()
	trimmed := strings.TrimPrefix(imp, sharedImportPrefix)
	parts := strings.Split(trimmed, "/")
	if len(parts) == 0 || parts[0] == "" {
		t.Fatalf("invalid shared import path: %s imports %s", path, imp)
	}
	if _, ok := sharedAllowlist[parts[0]]; !ok {
		t.Fatalf("shared package not allowlisted: %s imports %s", path, imp)
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
