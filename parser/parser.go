package parser

/*
	https://github.com/pacedotdev/oto.git
	The MIT License (MIT)
	Copyright (c) 2021 Pace Software Ltd

	Modifications:
	Removed oto specific fields and functions.
	Removed the 1 param 1 return value constraint.
	Changes def to map[string]*Definition.
	Moved structs to models.go file.
*/

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/token"
	"go/types"
	"os"
	"sort"
	"strings"

	"github.com/fatih/structtag"
	"github.com/pkg/errors"

	"golang.org/x/tools/go/packages"
)

// Parser parses packages.
type Parser struct {
	Verbose bool

	ExcludeInterfaces []string

	patterns []string
	def      map[string]*Definition

	// outputObjects marks output object names.
	outputObjects map[string]struct{}
	// objects marks object names.
	objects map[string]struct{}

	// docs are the docs for extracting comments.
	docs *doc.Package
}

// New makes a fresh parser using the specified patterns.
// The patterns should be the args passed into the tool (after any flags)
// and will be passed to the underlying build system.
func New(patterns ...string) *Parser {
	return &Parser{
		patterns: patterns,
	}
}

func (p Parser) Parse() (map[string]*Definition, error) {
	cfg := &packages.Config{
		Mode:  packages.NeedTypes | packages.NeedName | packages.NeedTypesInfo | packages.NeedDeps | packages.NeedName | packages.NeedSyntax,
		Tests: false,
	}

	pkgs, err := packages.Load(cfg, p.patterns...)
	if err != nil {
		fmt.Println(fmt.Errorf("error loading packages %s", err))
		os.Exit(1)
	}

	p.def = make(map[string]*Definition)
	p.outputObjects = make(map[string]struct{})
	p.objects = make(map[string]struct{})
	var excludedObjectsTypeIDs []string
	for _, pkg := range pkgs {
		p.docs, err = doc.NewFromFiles(pkg.Fset, pkg.Syntax, "")
		if err != nil {
			panic(err)
		}
		d := &Definition{}
		p.def[pkg.PkgPath] = d
		d.PackageName = pkg.Name
		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			obj := scope.Lookup(name)
			switch item := obj.Type().Underlying().(type) {
			case *types.Interface:
				s, err := p.parseService(pkg, obj, item)
				if err != nil {
					return p.def, err
				}
				if isInSlice(p.ExcludeInterfaces, name) {
					for _, method := range s.Methods {
						for i := range method.InputObjects {
							excludedObjectsTypeIDs = append(excludedObjectsTypeIDs, method.InputObjects[i].TypeID)
						}
						for i := range method.InputObjects {
							excludedObjectsTypeIDs = append(excludedObjectsTypeIDs, method.OutputObjects[i].TypeID)
						}

					}
					continue
				}
				d.Services = append(d.Services, s)
			case *types.Struct:
				p.parseObject(pkg, obj, item)
			}
		}

		// remove any excluded objects
		nonExcludedObjects := make([]Object, 0, len(d.Objects))
		for _, object := range d.Objects {
			excluded := false
			for _, excludedTypeID := range excludedObjectsTypeIDs {
				if object.TypeID == excludedTypeID {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}
			nonExcludedObjects = append(nonExcludedObjects, object)
		}
		d.Objects = nonExcludedObjects
		// sort services
		sort.Slice(d.Services, func(i, j int) bool {
			return d.Services[i].Name < d.Services[j].Name
		})
		// sort objects
		sort.Slice(d.Objects, func(i, j int) bool {
			return d.Objects[i].Name < d.Objects[j].Name
		})
	}

	return p.def, nil
}

func (p *Parser) parseService(pkg *packages.Package, obj types.Object, interfaceType *types.Interface) (Service, error) {
	var s Service
	s.Name = obj.Name()
	s.Comment = p.commentForType(s.Name)
	if p.Verbose {
		fmt.Printf("%s ", s.Name)
	}
	l := interfaceType.NumMethods()
	for i := 0; i < l; i++ {
		m := interfaceType.Method(i)
		method, err := p.parseMethod(pkg, s.Name, m)
		if err != nil {
			return s, err
		}
		s.Methods = append(s.Methods, method)
	}
	return s, nil
}

func (p *Parser) parseMethod(pkg *packages.Package, serviceName string, methodType *types.Func) (Method, error) {
	var m Method
	m.Name = methodType.Name()
	m.Comment = p.commentForMethod(serviceName, m.Name)
	sig := methodType.Type().(*types.Signature)
	inputParams := sig.Params()

	for i := 0; i < inputParams.Len(); i++ {
		field, err := p.parseFieldType(pkg, inputParams.At(i))
		if err != nil {
			return m, errors.Wrap(err, "parse input object type")
		}

		m.InputObjects = append(m.InputObjects, field)
	}

	outputParams := sig.Results()

	for i := 0; i < outputParams.Len(); i++ {
		field, err := p.parseFieldType(pkg, outputParams.At(i))
		if err != nil {
			return m, errors.Wrap(err, "parse output object type")
		}

		m.OutputObjects = append(m.OutputObjects, field)
		p.outputObjects[field.TypeName] = struct{}{}
	}
	return m, nil
}

func (p *Parser) parseFieldType(pkg *packages.Package, obj types.Object) (FieldType, error) {
	var ftype FieldType
	pkgPath := pkg.PkgPath
	d := p.def[pkgPath]
	resolver := func(other *types.Package) string {
		if other.Name() != pkg.Name {
			if d.Imports == nil {
				d.Imports = make(map[string]string)
			}
			d.Imports[other.Path()] = other.Name()
			ftype.Package = other.Path()
			pkgPath = other.Path()
			return other.Name()
		}
		return "" // no package prefix
	}

	typ := obj.Type()
	if slice, ok := obj.Type().(*types.Slice); ok {
		typ = slice.Elem()
		ftype.Multiple = true
	}
	isPointer := true
	originalTyp := typ
	pointerType, isPointer := typ.(*types.Pointer)
	if isPointer {
		typ = pointerType.Elem()
		isPointer = true
	}
	if named, ok := typ.(*types.Named); ok {
		if structure, ok := named.Underlying().(*types.Struct); ok {
			if err := p.parseObject(pkg, named.Obj(), structure); err != nil {
				return ftype, err
			}
			ftype.IsObject = true
		}
	}
	// disallow nested structs
	switch typ.(type) {
	case *types.Struct:
		return ftype, p.wrapErr(errors.New("nested structs not supported (create another type instead)"), pkg, obj.Pos())
	}
	ftype.TypeName = types.TypeString(originalTyp, resolver)
	ftype.ObjectName = types.TypeString(originalTyp, func(other *types.Package) string { return "" })
	ftype.TypeID = pkgPath + "." + ftype.ObjectName
	ftype.CleanObjectName = strings.TrimPrefix(ftype.ObjectName, "*")

	return ftype, nil
}

// parseObject parses a struct type and adds it to the Definition.
func (p *Parser) parseObject(pkg *packages.Package, o types.Object, v *types.Struct) error {
	var obj Object
	obj.Name = o.Name()
	obj.Comment = p.commentForType(obj.Name)
	var err error
	if err != nil {
		return p.wrapErr(errors.New("extract comment metadata"), pkg, o.Pos())
	}
	if _, found := p.objects[obj.Name]; found {
		// if this has already been parsed, skip it
		return nil
	}
	if o.Pkg().Name() != pkg.Name {
		obj.Imported = true
	}
	typ := v.Underlying()
	st, ok := typ.(*types.Struct)
	if !ok {
		return p.wrapErr(errors.New(obj.Name+" must be a struct"), pkg, o.Pos())
	}
	obj.TypeID = o.Pkg().Path() + "." + obj.Name
	obj.Fields = []Field{}
	for i := 0; i < st.NumFields(); i++ {
		field, err := p.parseField(pkg, obj.Name, st.Field(i), st.Tag(i))
		if err != nil {
			return err
		}
		field.Tag = v.Tag(i)
		field.ParsedTags, err = p.parseTags(field.Tag)
		if err != nil {
			return errors.Wrap(err, "parse field tag")
		}
		obj.Fields = append(obj.Fields, field)
	}
	d := p.def[pkg.PkgPath]
	d.Objects = append(d.Objects, obj)
	p.objects[obj.Name] = struct{}{}
	return nil
}

func (p *Parser) parseTags(tag string) (map[string]FieldTag, error) {
	tags, err := structtag.Parse(tag)
	if err != nil {
		return nil, err
	}
	fieldTags := make(map[string]FieldTag)
	for _, tag := range tags.Tags() {
		fieldTags[tag.Key] = FieldTag{
			Value:   tag.Name,
			Options: tag.Options,
		}
	}
	return fieldTags, nil
}

func (p *Parser) parseField(pkg *packages.Package, objectName string, v *types.Var, tag string) (Field, error) {
	var f Field
	f.Name = v.Name()

	f.Comment = p.commentForField(objectName, f.Name)

	var err error
	f.Type, err = p.parseFieldType(pkg, v)
	if err != nil {
		return f, errors.Wrap(err, "parse type")
	}
	return f, nil
}

func (p *Parser) wrapErr(err error, pkg *packages.Package, pos token.Pos) error {
	position := pkg.Fset.Position(pos)
	return errors.Wrap(err, position.String())
}

func (p *Parser) commentForField(typeName, field string) string {
	typ := p.lookupType(typeName)
	if typ == nil {
		return ""
	}
	spec, ok := typ.Decl.Specs[0].(*ast.TypeSpec)
	if !ok {
		return ""
	}
	obj, ok := spec.Type.(*ast.StructType)
	if !ok {
		return ""
	}
	var f *ast.Field
outer:
	for i := range obj.Fields.List {
		for _, name := range obj.Fields.List[i].Names {
			if name.Name == field {
				f = obj.Fields.List[i]
				break outer
			}
		}
	}
	if f == nil {
		return ""
	}
	return cleanComment(f.Doc.Text())
}

func (p *Parser) commentForType(name string) string {
	typ := p.lookupType(name)
	if typ == nil {
		return ""
	}
	return cleanComment(typ.Doc)
}

// I think this looks at the interface to grab the comment
func (p *Parser) commentForMethod(service, method string) string {
	typ := p.lookupType(service)
	if typ == nil {
		return ""
	}
	spec, ok := typ.Decl.Specs[0].(*ast.TypeSpec)
	if !ok {
		return ""
	}
	iface, ok := spec.Type.(*ast.InterfaceType)
	if !ok {
		return ""
	}
	var m *ast.Field
outer:
	for i := range iface.Methods.List {
		for _, name := range iface.Methods.List[i].Names {
			if name.Name == method {
				m = iface.Methods.List[i]
				break outer
			}
		}
	}
	if m == nil {
		return ""
	}
	return cleanComment(m.Doc.Text())
}

func (p *Parser) lookupType(name string) *doc.Type {
	for i := range p.docs.Types {
		if p.docs.Types[i].Name == name {
			return p.docs.Types[i]
		}
	}
	return nil
}

func cleanComment(s string) string {
	return strings.TrimSpace(s)
}

func isInSlice(slice []string, s string) bool {
	for i := range slice {
		if slice[i] == s {
			return true
		}
	}
	return false
}
