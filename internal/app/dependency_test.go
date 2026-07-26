package app

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// apiKitSiblingPairs 列出禁止相互导入的兄弟 API Kit（双向）。
var apiKitSiblingPairs = [][2]string{
	{"apimtauth", "apimtbrowse"},
	{"apimtbrowse", "apimtauth"},
}

func TestProductionPackageDependencies(t *testing.T) {
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test path")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(current), "../.."))

	err := filepath.WalkDir(filepath.Join(root, "pkg"), func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		for _, spec := range file.Imports {
			importPath, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			if strings.HasPrefix(importPath, "mtlogin/internal/") {
				t.Errorf("%s imports private application package %s", rel, importPath)
			}
			// 兄弟 API Kit 不得相互导入（双向禁止）。
			slashRel := filepath.ToSlash(rel)
			for _, pair := range apiKitSiblingPairs {
				if strings.HasPrefix(slashRel, "pkg/apikit/"+pair[0]+"/") &&
					strings.HasPrefix(importPath, "mtlogin/pkg/apikit/"+pair[1]) {
					t.Errorf("%s imports sibling API kit %s", rel, importPath)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestInternalDependencyBoundaries(t *testing.T) {
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test path")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(current), "../.."))
	cases := []struct {
		dir       string
		forbidden []string
	}{
		{dir: "internal/refresh", forbidden: []string{"mtlogin/pkg/conf", "mtlogin/pkg/apikit/apinotify"}},
		{dir: "internal/scheduler", forbidden: []string{"mtlogin/pkg/conf"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.dir, func(t *testing.T) {
			assertImportsExcluded(t, root, testCase.dir, testCase.forbidden)
		})
	}
	assertSourceExcludes(t, filepath.Join(root, "internal/app/app.go"), "apicore.Chain(")
}

func assertImportsExcluded(t *testing.T, root, relativeDir string, forbidden []string) {
	t.Helper()
	err := filepath.WalkDir(filepath.Join(root, relativeDir), func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, spec := range file.Imports {
			importPath, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			for _, blocked := range forbidden {
				if importPath == blocked || strings.HasPrefix(importPath, blocked+"/") {
					t.Errorf("%s imports forbidden package %s", relativeDir, importPath)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func assertSourceExcludes(t *testing.T, path, forbidden string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), forbidden) {
		t.Fatalf("%s contains forbidden source %q", path, forbidden)
	}
}

// TestDomainHasNoInternalDependencies 守护 domain 作为零依赖叶子包，
// 防止未来引入循环依赖。
func TestDomainHasNoInternalDependencies(t *testing.T) {
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test path")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(current), "../.."))
	err := filepath.WalkDir(filepath.Join(root, "internal/domain"), func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, spec := range file.Imports {
			importPath, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			if strings.HasPrefix(importPath, "mtlogin/") {
				t.Errorf("internal/domain must stay dependency-free but imports %s", importPath)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
