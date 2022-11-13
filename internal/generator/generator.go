package generator

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/google/go-github/v48/github"
)

type Generator struct {
	f        *jen.File
	t        reflect.Type
	configs  []reflect.StructField
	services []reflect.StructField
}

func New() *Generator {
	return &Generator{
		f:        jen.NewFile("gogithubmockable"),
		t:        reflect.TypeOf(github.Client{}),
		configs:  make([]reflect.StructField, 0),
		services: make([]reflect.StructField, 0),
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

	g.GenClient()
	g.GenClientAPI()

	for _, service := range g.services {
		g.GenServiceInterface(service)
	}

	return g.f.Render(w)
}

func (g *Generator) GenClient() {
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
		getter := g.f.Func().Op("(").Id("c").Op("*").Qual("", "Client").Op(")").Id(fmt.Sprintf("Get%s", config.Name)).
			Params()

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
		g.f.Func().Op("(").Id("c").Op("*").Qual("", "Client").Op(")").Id(fmt.Sprintf("Get%s", service.Name)).
			Params().
			Qual("", fmt.Sprintf("%sService", service.Name)).
			Block(jen.Return(jen.Id("c").Dot("client").Dot(service.Name))).
			Line()
	}
}

func (g *Generator) GenClientAPI() {
	configs := g.configs
	services := g.services

	g.f.Type().Id("ClientAPI").
		InterfaceFunc(func(g *jen.Group) {
			for _, config := range configs {
				getter := fmt.Sprintf("Get%s", config.Name)
				setter := fmt.Sprintf("Set%s", config.Name)

				ct := config.Type
				gid := g.Id(getter).Params()
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
				name := fmt.Sprintf("Get%s", service.Name)
				g.Id(name).Params().Qual("", fmt.Sprintf("%sService", service.Name))
			}
		}).
		Line()
}

func (g *Generator) GenServiceInterface(service reflect.StructField) {
	g.f.Type().Id(fmt.Sprintf("%sService", service.Name)).
		InterfaceFunc(func(g *jen.Group) {
			for i := 0; i < service.Type.NumMethod(); i++ {
				m := service.Type.Method(i)
				if !m.IsExported() {
					continue
				}

				g.Id(m.Name).
					ParamsFunc(func(g *jen.Group) {
						for i := 1; i < m.Type.NumIn(); i++ {
							p := m.Type.In(i)
							id := g.Id("")

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
