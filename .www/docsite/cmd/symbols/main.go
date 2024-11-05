package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/tylermmorton/torque/.www/docsite/model"
)

func main() {
	inputPath := flag.String("i", ".", "Path to the package directory")
	outputPath := flag.String("o", "symbols.json", "Output JSON file path")
	flag.Parse() // Parse the command-line flags

	symbols, err := analyzePackage(*inputPath)
	if err != nil {
		log.Fatalf("Error analyzing package: %v", err)
	}

	jsonData, err := json.MarshalIndent(symbols, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling JSON: %v", err)
	}

	err = os.WriteFile(*outputPath, jsonData, 0666)
	if err != nil {
		log.Fatalf("Error outputting json file: %v", err)
	}
}

func analyzePackage(path string) ([]*model.Symbol, error) {
	fset := token.NewFileSet()
	symbols := make([]*model.Symbol, 0)

	// Parse all Go files in the directory, skipping test files
	pkgs, err := parser.ParseDir(fset, path, func(info os.FileInfo) bool {
		// Skip test files
		return !info.IsDir() && !strings.HasSuffix(info.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return symbols, fmt.Errorf("failed to parse package: %w", err)
	}

	// Analyze each package found in the directory
	for _, pkg := range pkgs {
		config := &types.Config{
			Importer: importer.For("source", nil), // Set up the importer for type-checking
			Error:    func(err error) {},
		}
		info := &types.Info{
			Defs: make(map[*ast.Ident]types.Object),
		}

		// Perform type-checking on all parsed files
		if _, err := config.Check(pkg.Name, fset, collectFiles(pkg), info); err != nil {
			return symbols, fmt.Errorf("type checking failed: %w", err)
		}

		// Analyze each file in the package
		for _, file := range pkg.Files {
			symbols = analyzeAST(file, info, symbols, fset)
		}
	}

	return symbols, nil
}

// collectFiles returns a slice of parsed files for type-checking.
func collectFiles(pkg *ast.Package) []*ast.File {
	var files []*ast.File
	for _, file := range pkg.Files {
		files = append(files, file)
	}
	return files
}

func analyzeAST(file *ast.File, info *types.Info, symbols []*model.Symbol, fset *token.FileSet) []*model.Symbol {
	for _, decl := range file.Decls {
		switch decl := decl.(type) {
		case *ast.GenDecl:
			if decl.Tok == token.TYPE {
				for _, spec := range decl.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						pos := fset.Position(typeSpec.Pos())
						symbol := model.Symbol{
							Package:      "",
							Name:         typeSpec.Name.Name,
							Kind:         determineTypeKind(typeSpec.Type),
							IsExported:   ast.IsExported(typeSpec.Name.Name),
							FileName:     filepath.Base(fset.File(typeSpec.Pos()).Name()),
							LineNumber:   uint(pos.Line),
							LinePosition: uint(pos.Column),
						}

						if decl.Doc != nil {
							symbol.Comments = decl.Doc.Text()
						}

						start := fset.Position(typeSpec.Pos()).Offset
						end := fset.Position(typeSpec.End()).Offset
						sourceCode, _ := extractSourceCode(fset, file, start, end)
						symbol.Source = "type " + sourceCode

						symbols = append(symbols, &symbol)
					}
				}
			}

		case *ast.FuncDecl:
			pos := fset.Position(decl.Pos())
			symbol := model.Symbol{
				Package:      "",
				Name:         decl.Name.Name,
				Kind:         "function",
				IsExported:   decl.Name.IsExported(),
				FileName:     filepath.Base(fset.File(decl.Pos()).Name()),
				LineNumber:   uint(pos.Line),
				LinePosition: uint(pos.Column),
			}

			if decl.Doc != nil {
				symbol.Comments = decl.Doc.Text()
			}

			// Get source code for the function
			start := fset.Position(decl.Pos()).Offset
			end := fset.Position(decl.End()).Offset
			sourceCode, _ := extractSourceCode(fset, file, start, end)
			symbol.Source = sourceCode // Include function source code

			// Extract parameters and returns
			symbol.Parameters = getFieldTypes(decl.Type.Params)
			symbol.Returns = getFieldTypes(decl.Type.Results)

			// Extract receiver type if it exists
			if decl.Recv != nil && len(decl.Recv.List) > 0 {
				recv := decl.Recv.List[0] // Assuming a single receiver for now
				if starExpr, ok := recv.Type.(*ast.StarExpr); ok {
					if ident, ok := starExpr.X.(*ast.Ident); ok {
						symbol.Receiver = &ident.Name
					}
				} else if ident, ok := recv.Type.(*ast.Ident); ok {
					symbol.Receiver = &ident.Name
				}
			}

			// Add the symbol to the slice
			symbols = append(symbols, &symbol)
		}
	}
	return symbols
}

func determineTypeKind(expr ast.Expr) string {
	return strings.Split(strings.TrimPrefix(fmt.Sprintf("%T", expr), "*"), ".")[1]
}

func extractSourceCode(fset *token.FileSet, file *ast.File, start, end int) (string, error) {
	// Get the filename from the FileSet
	filename := fset.File(file.Pos()).Name()

	// Read the full file content
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Build a filtered code string
	var filteredCode string
	codeMap := make(map[token.Pos]struct{}) // To track positions of comments

	// Traverse the AST to collect comments' positions
	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		switch n := n.(type) {
		case *ast.CommentGroup:
			// Store positions of comments
			for _, comment := range n.List {
				commentStart := fset.Position(comment.Pos()).Offset
				commentEnd := fset.Position(comment.End()).Offset
				for i := commentStart; i < commentEnd; i++ {
					codeMap[token.Pos(i)] = struct{}{}
				}
			}
		case *ast.GenDecl:
			//if n.Tok == token.TYPE { // Check for type declarations
			//	for _, spec := range n.Specs {
			//		if _, ok := spec.(*ast.TypeSpec); ok {
			//			typeFound = true
			//		}
			//	}
			//}
		}

		return true
	})

	// Extract the source code without comments
	for pos := start; pos < end; pos++ {
		if _, exists := codeMap[token.Pos(pos)]; !exists {
			filteredCode += string(content[pos]) // Append character if not a comment
		}
	}

	return filteredCode, nil
}

func getFieldTypes(fields *ast.FieldList) []string {
	if fields == nil {
		return nil
	}

	typeNames := make([]string, 0, fields.NumFields())
	for _, field := range fields.List {
		typeStr := types.ExprString(field.Type)
		typeNames = append(typeNames, typeStr)
	}
	return typeNames
}
