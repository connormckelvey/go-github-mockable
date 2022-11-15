package generator

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/doc"
	"go/parser"
	"go/token"
	"io"
	"reflect"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/google/go-github/v48/github"
	"github.com/google/uuid"
)

type Generator struct {
	f           *jen.File
	t           reflect.Type
	pt          reflect.Type
	configs     []reflect.StructField
	services    []reflect.StructField
	methods     []reflect.Method
	packageInfo *PackageInfo
}

func New() *Generator {
	return &Generator{
		f:           jen.NewFile("gogithubmockable"),
		t:           reflect.TypeOf(github.Client{}),
		pt:          reflect.TypeOf(&github.Client{}),
		configs:     make([]reflect.StructField, 0),
		services:    make([]reflect.StructField, 0),
		methods:     make([]reflect.Method, 0),
		packageInfo: newPackageInfo(),
	}
}

func (g *Generator) Generate(w io.Writer) error {
	for i := 0; i < g.t.NumField(); i++ {
		f := g.t.Field(i)
		if !f.IsExported() {
			continue
		}
		if f.Type.Kind() == reflect.Pointer && strings.HasSuffix(f.Type.Elem().Name(), "Service") {
			g.services = append(g.services, f)
		} else {
			g.configs = append(g.configs, f)
		}
	}

	for i := 0; i < g.pt.NumMethod(); i++ {
		m := g.pt.Method(i)
		if !m.IsExported() {
			continue
		}
		g.methods = append(g.methods, m)
	}

	pkg, err := build.Import(
		g.t.PkgPath(),
		"",
		build.FindOnly,
	)
	if err != nil {
		return err
	}

	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, pkg.Dir, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	for _, typ := range doc.New(pkgs["github"], "", doc.AllDecls).Types {
		typeInfo := newTypeInfo(typ.Name)
		typeInfo.Doc = typ.Doc

		st, ok := typ.Decl.Specs[0].(*ast.TypeSpec).Type.(*ast.StructType)
		if ok {
			for _, f := range st.Fields.List {
				name := uuid.NewString()
				if len(f.Names) > 0 {
					name = f.Names[0].Name
				}
				typeInfo.FieldDocs[name] = f.Doc.Text()
			}
		}
		for _, m := range typ.Methods {
			typeInfo.MethodDocs[m.Name] = m.Doc
			params := []string{}
			for _, p := range m.Decl.Type.Params.List {
				for _, name := range p.Names {
					params = append(params, name.Name)

				}
			}
			typeInfo.MethodParams[m.Name] = params
		}

		g.packageInfo.Types[typeInfo.Name] = typeInfo
	}

	g.GenClient()
	g.GenClientAPI()

	for _, service := range g.services {
		g.GenServiceInterface(service)
	}

	return g.f.Render(w)
}

func (g *Generator) GenClient() {
	g.f.Comment(g.packageInfo.Types["Client"].Doc)
	g.f.Type().Id("Client").Struct(
		jen.Id("client").Op("*").Qual(g.t.PkgPath(), g.t.Name()),
	)

	g.f.Func().Id("NewClient").
		Params(jen.Id("client").Op("*").Qual(g.t.PkgPath(), g.t.Name())).
		Op("*").Qual("", "Client").
		BlockFunc(func(g *jen.Group) {
			g.Id("c").Op(":=").New(jen.Id("Client"))
			g.Id("c").Dot("client").Op("=").Id("client")
			g.Return(jen.Id("c"))
		}).
		Line()

	for _, config := range g.configs {
		g.f.Comment(fmt.Sprintf("Get%s returns the %s", config.Name, g.packageInfo.Types["Client"].FieldDocs[config.Name]))
		getter := g.f.Func().Op("(").Id("c").Op("*").Qual("", "Client").Op(")").Id(fmt.Sprintf("Get%s", config.Name)).
			Params()

		g.f.Comment(fmt.Sprintf("Set%s sets the %s", config.Name, g.packageInfo.Types["Client"].FieldDocs[config.Name]))
		setter := g.f.Func().Op("(").Id("c").Op("*").Qual("", "Client").Op(")").Id(fmt.Sprintf("Set%s", config.Name))

		ct := config.Type
		param := jen.Id(strings.ToLower(config.Name))
		if config.Type.Kind() == reflect.Pointer {
			ct = config.Type.Elem()
			getter.Op("*")
			param.Op("*")
		}
		getter.Qual(ct.PkgPath(), ct.Name())
		param.Qual(ct.PkgPath(), ct.Name())

		setter.Params(param)

		getter.Block(jen.Return(jen.Id("c").Dot("client").Dot(config.Name))).
			Line()

		setter.BlockFunc(func(g *jen.Group) {
			g.Id("c").Dot("client").Dot(config.Name).Op("=").Id(strings.ToLower(config.Name))
		}).
			Line()
	}

	for _, service := range g.services {
		serviceInfo := g.packageInfo.Types[service.Name+"Service"]
		if serviceInfo != nil {
			g.f.Comment(serviceInfo.Doc)
		}

		g.f.Func().Op("(").Id("c").Op("*").Qual("", "Client").Op(")").Id(fmt.Sprintf("%s", service.Name)).
			Params().
			Qual("", fmt.Sprintf("%sService", service.Name)).
			Block(jen.Return(jen.Id("c").Dot("client").Dot(service.Name))).
			Line()
	}

	for _, m := range g.methods {
		mDocs := g.packageInfo.Types["Client"].MethodDocs[m.Name]
		mParams := g.packageInfo.Types["Client"].MethodParams[m.Name]

		g.f.Comment(mDocs)
		g.f.Func().Op("(").Id("c").Op("*").Qual("", "Client").Op(")").Id(m.Name).
			ParamsFunc(func(g *jen.Group) {
				for i := 1; i < m.Type.NumIn(); i++ {
					p := m.Type.In(i)

					var idStr string
					if len(mParams) > i-1 {
						idStr = mParams[i-1]
					}
					id := g.Id(idStr)

					if p.Kind() == reflect.Slice && p.Name() == "" {
						p = p.Elem()
						id = id.Op("[]")
					}
					if p.Kind() == reflect.Pointer {
						p = p.Elem()
						id = id.Op("*")
					} else if p.Kind() == reflect.Interface && p.Name() == "" {
						id = id.Interface()
					} else if p.Kind() == reflect.Map {
						id.Map(jen.Qual(p.Key().PkgPath(), p.Key().Name())).
							Qual(p.Elem().PkgPath(), p.Elem().Name())
						continue
					}
					id.Qual(p.PkgPath(), p.Name())
				}
			}).
			Op("(").
			ListFunc(func(g *jen.Group) {
				for i := 0; i < m.Type.NumOut(); i++ {
					p := m.Type.Out(i)
					id := g.Id("")

					if p.Kind() == reflect.Slice {
						p = p.Elem()
						id = id.Op("[]")
					}
					if p.Kind() == reflect.Pointer {
						p = p.Elem()
						id = id.Op("*")
					} else if p.Kind() == reflect.Map {
						id.Map(jen.Qual(p.Key().PkgPath(), p.Key().Name())).
							Qual(p.Elem().PkgPath(), p.Elem().Name())
						continue
					}
					id.Qual(p.PkgPath(), p.Name())
				}
			}).
			Op(")").
			BlockFunc(func(g *jen.Group) {
				g.Return(jen.Id("c").Dot("client").Dot(m.Name).CallFunc(func(g *jen.Group) {
					for _, p := range mParams {
						g.Id(p)
					}
				}))
			})
	}
}

func (g *Generator) GenClientAPI() {
	configs := g.configs
	services := g.services
	methods := g.methods
	packageInfo := g.packageInfo

	g.f.Type().Id("ClientAPI").
		InterfaceFunc(func(g *jen.Group) {
			for _, config := range configs {
				getter := fmt.Sprintf("Get%s", config.Name)
				setter := fmt.Sprintf("Set%s", config.Name)

				ct := config.Type

				g.Comment(fmt.Sprintf("Get%s returns the %s", config.Name, packageInfo.Types["Client"].FieldDocs[config.Name]))
				gid := g.Id(getter).Params()

				g.Comment(fmt.Sprintf("Set%s sets the %s", config.Name, packageInfo.Types["Client"].FieldDocs[config.Name]))
				sid := g.Id(setter)

				if config.Type.Kind() == reflect.Pointer {
					ct = config.Type.Elem()
					gid = gid.Op("*")
					sid = sid.Params(jen.Op("*").Qual(ct.PkgPath(), ct.Name()))
				} else {
					sid.Params(jen.Qual(ct.PkgPath(), ct.Name()))
				}
				gid.Qual(ct.PkgPath(), ct.Name())
			}

			for _, service := range services {
				name := fmt.Sprintf("%s", service.Name)
				serviceInfo := packageInfo.Types[service.Name+"Service"]

				if serviceInfo != nil {
					g.Comment(serviceInfo.Doc)
				}
				g.Id(name).Params().Qual("", fmt.Sprintf("%sService", service.Name))
			}

			for _, m := range methods {
				mDocs := packageInfo.Types["Client"].MethodDocs[m.Name]
				mParams := packageInfo.Types["Client"].MethodParams[m.Name]

				g.Comment(mDocs)
				g.Id(m.Name).
					ParamsFunc(func(g *jen.Group) {
						for i := 1; i < m.Type.NumIn(); i++ {
							p := m.Type.In(i)

							var idStr string
							if len(mParams) > i-1 {
								idStr = mParams[i-1]
							}
							id := g.Id(idStr)

							if p.Kind() == reflect.Slice && p.Name() == "" {
								p = p.Elem()
								id = id.Op("[]")
							}
							if p.Kind() == reflect.Pointer {
								p = p.Elem()
								id = id.Op("*")
							} else if p.Kind() == reflect.Interface && p.Name() == "" {
								id = id.Interface()
							} else if p.Kind() == reflect.Map {
								id.Map(jen.Qual(p.Key().PkgPath(), p.Key().Name())).
									Qual(p.Elem().PkgPath(), p.Elem().Name())
								continue
							}
							id.Qual(p.PkgPath(), p.Name())
						}
					}).
					Op("(").
					ListFunc(func(g *jen.Group) {
						for i := 0; i < m.Type.NumOut(); i++ {
							p := m.Type.Out(i)
							id := g.Id("")

							if p.Kind() == reflect.Slice {
								p = p.Elem()
								id = id.Op("[]")
							}
							if p.Kind() == reflect.Pointer {
								p = p.Elem()
								id = id.Op("*")
							} else if p.Kind() == reflect.Map {
								id.Map(jen.Qual(p.Key().PkgPath(), p.Key().Name())).
									Qual(p.Elem().PkgPath(), p.Elem().Name())
								continue
							}
							id.Qual(p.PkgPath(), p.Name())
						}
					}).
					Op(")")
			}
		}).
		Line()
}

func (g *Generator) GenServiceInterface(service reflect.StructField) {
	serviceInfo := g.packageInfo.Types[service.Name+"Service"]
	if serviceInfo != nil {
		g.f.Comment(serviceInfo.Doc)
	}
	g.f.Type().Id(fmt.Sprintf("%sService", service.Name)).
		InterfaceFunc(func(g *jen.Group) {
			for i := 0; i < service.Type.NumMethod(); i++ {
				m := service.Type.Method(i)
				if !m.IsExported() {
					continue
				}
				if serviceInfo != nil {
					g.Comment(serviceInfo.MethodDocs[m.Name])

				}
				g.Id(m.Name).
					ParamsFunc(func(g *jen.Group) {
						for i := 1; i < m.Type.NumIn(); i++ {
							p := m.Type.In(i)

							var idStr string
							if serviceInfo != nil {
								paramsInfo := serviceInfo.MethodParams[m.Name]
								if len(paramsInfo) > i-1 {
									idStr = paramsInfo[i-1]
								}
							}

							id := g.Id(idStr)

							if p.Kind() == reflect.Slice && p.Name() == "" {
								p = p.Elem()
								id = id.Op("[]")
							}
							if p.Kind() == reflect.Pointer {
								p = p.Elem()
								id = id.Op("*")
							} else if p.Kind() == reflect.Map {
								id.Map(jen.Qual(p.Key().PkgPath(), p.Key().Name())).
									Qual(p.Elem().PkgPath(), p.Elem().Name())
								continue
							}
							id.Qual(p.PkgPath(), p.Name())
						}
					}).
					Op("(").
					ListFunc(func(g *jen.Group) {
						for i := 0; i < m.Type.NumOut(); i++ {
							p := m.Type.Out(i)
							id := g.Id("")

							if p.Kind() == reflect.Slice {
								p = p.Elem()
								id = id.Op("[]")
							}
							if p.Kind() == reflect.Pointer {
								p = p.Elem()
								id = id.Op("*")
							} else if p.Kind() == reflect.Map {
								id.Map(jen.Qual(p.Key().PkgPath(), p.Key().Name())).
									Qual(p.Elem().PkgPath(), p.Elem().Name())
								continue
							}
							id.Qual(p.PkgPath(), p.Name())
						}
					}).
					Op(")")
			}
		}).
		Line()
}
