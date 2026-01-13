package yml

import (
	"fmt"
	"go/ast"
	"go/token"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestGenerateDocs(t *testing.T) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Dir:  ".",
		Fset: token.NewFileSet(),
	}
	pkgs, err := packages.Load(cfg)

	if err != nil {
		t.Fatalf("Failed to load packages: %v", err)
	}

	results := strings.Builder{}

	for _, pkg := range pkgs {
		for _, syntaxTree := range pkg.Syntax {
			for _, decl := range syntaxTree.Decls {
				genDecl, ok := decl.(*ast.GenDecl)
				if !ok {
					// Not a declaration containing types
					continue
				}

				if len(genDecl.Specs) > 1 {
					// Block declaration, we don't do that for types
					continue
				}

				typeNode, ok := genDecl.Specs[0].(*ast.TypeSpec)
				if !ok {
					// Not a type
					continue
				}

				results.WriteString("\n---\n")
				results.WriteString(getTypeDocs(typeNode.Name.Name, genDecl.Doc.Text()))

				structNode, ok := typeNode.Type.(*ast.StructType)
				if !ok || structNode.Fields == nil {
					// Not a struct type or has no fields
					continue
				}

				for _, field := range structNode.Fields.List {
					if len(field.Names) != 1 {
						// Embedded field or other node
						continue
					}

					results.WriteString(getFieldDocs(field.Names[0].Name, field.Doc.Text(), field.Tag.Value))
				}
			}
		}
	}

	fmt.Println(results.String())
}

func getTypeDocs(typeName string, docs string) string {
	if docs == "" {
		return fmt.Sprintf("\n## Documentation missing for %s\n\n", typeName)
	}

	// No transformation for types, we don't care about type info
	return docs + "\n"
}

func getFieldDocs(name string, doc string, tag string) string {
	structTag := reflect.StructTag(tag)
	yamlTag, ok := structTag.Lookup("yaml")
	if ok {
		yamlName := strings.Split(yamlTag, ",")[0]

		return "`" + yamlName + "`\n" + doc + "\n"
	}

	// Not YML field
	return ""
}
