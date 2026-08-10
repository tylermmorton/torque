package analysis

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

// handlerInterfaces maps each torque handler interface name to its identifying method.
// All methods are exported, so nil-package lookup in a MethodSet works correctly.
var handlerInterfaces = []struct{ iface, method string }{
	{"TemplateProvider", "Template"},
	{"StyleSheetProvider", "StyleSheet"},
	{"FuncMapProvider", "FuncMap"},
	{"Loader", "Load"},
	{"Renderer", "Render"},
	{"ContextProvider", "Context"},
	{"Action", "Action"},
	{"RouterProvider", "Router"},
	{"LayoutProvider", "Layout"},
}

const loadMode = packages.NeedName |
	packages.NeedSyntax |
	packages.NeedTypes |
	packages.NeedFiles |
	packages.NeedCompiledGoFiles

type analyzer struct {
	dir      string
	pkgCache map[string]*packages.Package
	visited  map[string]bool
	manifest *ProjectManifest
}

// AnalyzePackages loads all packages matched by patterns and analyzes every type
// that implements TemplateProvider, returning a merged ProjectManifest.
// dir is the working directory for package loading (typically the module root).
func AnalyzePackages(patterns []string, dir string) (*ProjectManifest, error) {
	a := &analyzer{
		dir:      dir,
		pkgCache: make(map[string]*packages.Package),
		visited:  make(map[string]bool),
		manifest: &ProjectManifest{
			Components: make(map[string]*Component),
		},
	}

	cfg := &packages.Config{
		Mode: loadMode,
		Dir:  dir,
	}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, fmt.Errorf("packages.Load: %w", err)
	}

	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			var msgs []string
			for _, e := range pkg.Errors {
				msgs = append(msgs, e.Msg)
			}
			return nil, fmt.Errorf("errors in package %q: %s", pkg.PkgPath, strings.Join(msgs, "; "))
		}
		a.pkgCache[pkg.PkgPath] = pkg

		for _, name := range pkg.Types.Scope().Names() {
			obj := pkg.Types.Scope().Lookup(name)
			tn, ok := obj.(*types.TypeName)
			if !ok {
				continue
			}
			named, ok := tn.Type().(*types.Named)
			if !ok {
				continue
			}
			if !hasMethod(named, "Template") {
				continue
			}
			if err := a.analyzeType(named, pkg); err != nil {
				return nil, err
			}
		}
	}

	return a.manifest, nil
}

func (a *analyzer) analyzeType(named *types.Named, pkg *packages.Package) error {
	key := qualKey(named)
	if a.visited[key] {
		return nil
	}
	a.visited[key] = true

	comp := &Component{
		TypeName:    named.Obj().Name(),
		PackageName: named.Obj().Pkg().Name(),
		PackagePath: named.Obj().Pkg().Path(),
		SourceFile:  sourceFileForType(pkg, named.Obj().Name()),
		Interfaces:  detectInterfaces(named),
		Children:    []ComponentRef{},
	}

	if hasMethod(named, "Template") {
		tmpl, err := a.extractStringLiteral(named, pkg, "Template")
		if err != nil {
			return err
		}
		comp.Template = tmpl
	}

	if hasMethod(named, "StyleSheet") {
		css, err := a.extractStringLiteral(named, pkg, "StyleSheet")
		if err != nil {
			return err
		}
		comp.StyleSheet = css
	}

	if s, ok := named.Underlying().(*types.Struct); ok {
		if err := a.collectChildren(s, pkg, comp); err != nil {
			return err
		}
	}

	a.manifest.Components[key] = comp
	return nil
}

// collectChildren walks the fields of s, adding TemplateProvider fields as children
// of comp and recursing into non-TemplateProvider struct fields transparently.
// Transparent intermediates are not added to the component tree — only types with
// Template() are registered. This mirrors the runtime recurseFieldsImplementing behaviour.
func (a *analyzer) collectChildren(s *types.Struct, pkg *packages.Package, comp *Component) error {
	for i := 0; i < s.NumFields(); i++ {
		field := s.Field(i)
		tag := s.Tag(i)

		childNamed := unwrapNamed(field.Type())
		if childNamed == nil {
			continue
		}

		if hasMethod(childNamed, "Template") {
			// Case 1: field is a TemplateProvider — add as child and recurse via analyzeType.
			templateName := field.Name()
			if tv := reflect.StructTag(tag).Get("template"); tv != "" {
				templateName = tv
			}

			childKey := qualKey(childNamed)
			comp.Children = append(comp.Children, ComponentRef{
				FieldName:    field.Name(),
				TemplateName: templateName,
				TypeRef:      childKey,
			})

			if !a.visited[childKey] {
				childPkg, err := a.loadPkg(childNamed.Obj().Pkg().Path())
				if err != nil {
					return fmt.Errorf("loading package for %s: %w", childKey, err)
				}
				if err := a.analyzeType(childNamed, childPkg); err != nil {
					return err
				}
			}
		} else if innerStruct, ok := childNamed.Underlying().(*types.Struct); ok {
			// Case 2: field is not a TemplateProvider but is a struct — recurse
			// transparently, attaching any discovered children to the current comp.
			innerPkg, err := a.loadPkg(childNamed.Obj().Pkg().Path())
			if err != nil {
				return fmt.Errorf("loading package for intermediate %s: %w", qualKey(childNamed), err)
			}
			if err := a.collectChildren(innerStruct, innerPkg, comp); err != nil {
				return err
			}
		}
		// Case 3: neither TemplateProvider nor struct — skip.
	}
	return nil
}

func (a *analyzer) loadPkg(pattern string) (*packages.Package, error) {
	if pkg, ok := a.pkgCache[pattern]; ok {
		return pkg, nil
	}
	cfg := &packages.Config{
		Mode: loadMode,
		Dir:  a.dir,
	}
	pkgs, err := packages.Load(cfg, pattern)
	if err != nil {
		return nil, fmt.Errorf("packages.Load(%q): %w", pattern, err)
	}
	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no packages found for %q", pattern)
	}
	for _, p := range pkgs {
		if len(p.Errors) > 0 {
			var msgs []string
			for _, e := range p.Errors {
				msgs = append(msgs, e.Msg)
			}
			return nil, fmt.Errorf("errors in package %q: %s", p.PkgPath, strings.Join(msgs, "; "))
		}
		a.pkgCache[p.PkgPath] = p
	}
	return pkgs[0], nil
}

// extractStringLiteral finds the named method on the type and extracts its return
// value as a string. Returns an error if the body is not a single return of a
// string literal — dynamic values (variables, //go:embed, function calls) are not supported.
func (a *analyzer) extractStringLiteral(named *types.Named, pkg *packages.Package, methodName string) (*string, error) {
	fn := lookupMethod(named, methodName)
	if fn == nil {
		return nil, nil
	}

	// The method may be declared in a different package (e.g. promoted from an embedded type).
	// Use the method's signature to find the actual receiver type name, which is what we search
	// for in the AST — positions are not comparable across separately-loaded token.FileSets.
	sig, _ := fn.Type().(*types.Signature)
	if sig == nil || sig.Recv() == nil {
		return nil, nil
	}
	recvType := sig.Recv().Type()
	if ptr, ok := recvType.(*types.Pointer); ok {
		recvType = ptr.Elem()
	}
	recvNamed, ok := recvType.(*types.Named)
	if !ok {
		return nil, nil
	}
	receiverTypeName := recvNamed.Obj().Name()

	methodPkg := pkg
	if fn.Pkg() != nil && fn.Pkg().Path() != pkg.PkgPath {
		var err error
		methodPkg, err = a.loadPkg(fn.Pkg().Path())
		if err != nil {
			return nil, err
		}
	}

	funcDecl := findFuncDecl(methodPkg, receiverTypeName, methodName)
	if funcDecl == nil || funcDecl.Body == nil {
		return nil, nil
	}

	if len(funcDecl.Body.List) != 1 {
		return nil, fmt.Errorf(
			"%s.%s(): body must be a single return statement (dynamic values are not supported)",
			named.Obj().Name(), methodName,
		)
	}
	ret, ok := funcDecl.Body.List[0].(*ast.ReturnStmt)
	if !ok {
		return nil, fmt.Errorf("%s.%s(): body must be a return statement", named.Obj().Name(), methodName)
	}
	if len(ret.Results) != 1 {
		return nil, fmt.Errorf("%s.%s(): must return exactly one value", named.Obj().Name(), methodName)
	}
	lit, ok := ret.Results[0].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return nil, fmt.Errorf(
			"%s.%s(): return value must be a string literal (dynamic values, variables, and //go:embed are not supported)",
			named.Obj().Name(), methodName,
		)
	}
	val, err := strconv.Unquote(lit.Value)
	if err != nil {
		return nil, fmt.Errorf("%s.%s(): unquoting string literal: %w", named.Obj().Name(), methodName, err)
	}
	return &val, nil
}

// sourceFileForType finds the absolute path of the file declaring typeName in pkg
// by searching the AST — safe across separately-loaded token.FileSets.
func sourceFileForType(pkg *packages.Package, typeName string) string {
	for i, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok && ts.Name.Name == typeName {
					if i < len(pkg.CompiledGoFiles) {
						return pkg.CompiledGoFiles[i]
					}
				}
			}
		}
	}
	return ""
}

// findFuncDecl searches pkg's AST for a method on receiverTypeName with the given name.
// Matches by name rather than position — safe across separately-loaded token.FileSets.
func findFuncDecl(pkg *packages.Package, receiverTypeName, methodName string) *ast.FuncDecl {
	for _, file := range pkg.Syntax {
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Name.Name != methodName || fd.Recv == nil || len(fd.Recv.List) == 0 {
				continue
			}
			recvType := fd.Recv.List[0].Type
			if star, ok := recvType.(*ast.StarExpr); ok {
				recvType = star.X
			}
			if ident, ok := recvType.(*ast.Ident); ok && ident.Name == receiverTypeName {
				return fd
			}
		}
	}
	return nil
}

// detectInterfaces returns the names of torque handler API interfaces implemented by named.
func detectInterfaces(named *types.Named) []string {
	ptrMS := types.NewMethodSet(types.NewPointer(named))
	valMS := types.NewMethodSet(named)
	var result []string
	for _, hi := range handlerInterfaces {
		if ptrMS.Lookup(nil, hi.method) != nil || valMS.Lookup(nil, hi.method) != nil {
			result = append(result, hi.iface)
		}
	}
	return result
}

// hasMethod reports whether named (or *named) has an exported method with the given name.
func hasMethod(named *types.Named, name string) bool {
	if types.NewMethodSet(types.NewPointer(named)).Lookup(nil, name) != nil {
		return true
	}
	return types.NewMethodSet(named).Lookup(nil, name) != nil
}

// lookupMethod returns the *types.Func for the named method on named (or *named), or nil.
func lookupMethod(named *types.Named, name string) *types.Func {
	for _, ms := range []*types.MethodSet{
		types.NewMethodSet(types.NewPointer(named)),
		types.NewMethodSet(named),
	} {
		if sel := ms.Lookup(nil, name); sel != nil {
			if fn, ok := sel.Obj().(*types.Func); ok {
				return fn
			}
		}
	}
	return nil
}

// unwrapNamed unwraps a pointer and returns the underlying *types.Named, or nil.
func unwrapNamed(t types.Type) *types.Named {
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, _ := t.(*types.Named)
	return named
}

// qualKey returns the flat-map key for a type: "pkgImportPath.TypeName".
func qualKey(named *types.Named) string {
	return named.Obj().Pkg().Path() + "." + named.Obj().Name()
}
