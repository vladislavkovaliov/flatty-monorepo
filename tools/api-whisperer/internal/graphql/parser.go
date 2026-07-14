package graphql

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/validator"

	"github.com/vladislavkovaliov/api-whisperer/internal/model"
)

func Parse(path string) (*model.ApiSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read graphql file: %w", err)
	}

	source := string(data)

	schema, err := validator.LoadSchema(
		validator.Prelude,
		&ast.Source{Input: source},
	)
	if err == nil {
		return fromSchema(schema, model.SourceGraphQLSDL), nil
	}

	return parseIntrospection(source)
}

func parseIntrospection(data string) (*model.ApiSpec, error) {
	var intro struct {
		Data struct {
			SchemaData struct {
				QueryType        *struct{ Name string } `json:"queryType"`
				MutationType     *struct{ Name string } `json:"mutationType"`
				SubscriptionType *struct{ Name string } `json:"subscriptionType"`
				Types            []introFullType        `json:"types"`
			} `json:"__schema"`
		} `json:"data"`
	}

	if err := json.Unmarshal([]byte(data), &intro); err != nil {
		return nil, fmt.Errorf("parse introspection json: %w", err)
	}

	if intro.Data.SchemaData.QueryType == nil {
		return nil, fmt.Errorf("invalid introspection: no queryType found")
	}

	spec := &model.ApiSpec{
		Title:  "GraphQL API",
		Source: model.SourceGraphQLIntro,
	}

	typeMap := make(map[string]introFullType)
	for _, t := range intro.Data.SchemaData.Types {
		if strings.HasPrefix(t.Name, "__") {
			continue
		}
		typeMap[t.Name] = t
	}

	for _, t := range intro.Data.SchemaData.Types {
		if strings.HasPrefix(t.Name, "__") {
			continue
		}

		switch t.Kind {
		case "OBJECT", "INPUT_OBJECT":
			s := model.Schema{
				Name:        t.Name,
				Kind:        "struct",
				Description: t.Description,
			}
			if t.Kind == "INPUT_OBJECT" {
				for _, f := range t.InputFields {
					s.Fields = append(s.Fields, model.Field{
						Name:        f.Name,
						Description: f.Description,
						Type:        resolveIntroType(f.Type),
					})
				}
			} else {
				for _, f := range t.Fields {
					s.Fields = append(s.Fields, model.Field{
						Name:        f.Name,
						Description: f.Description,
						Type:        resolveIntroType(f.Type),
					})
				}
			}
			spec.Schemas = append(spec.Schemas, s)

		case "ENUM":
			s := model.Schema{
				Name:        t.Name,
				Kind:        "enum",
				Description: t.Description,
			}
			for _, v := range t.EnumValues {
				s.Values = append(s.Values, v.Name)
			}
			spec.Schemas = append(spec.Schemas, s)

		case "INTERFACE":
			spec.Schemas = append(spec.Schemas, model.Schema{
				Name:        t.Name,
				Kind:        "interface",
				Description: t.Description,
			})

		case "UNION":
			spec.Schemas = append(spec.Schemas, model.Schema{
				Name:        t.Name,
				Kind:        "union",
				Description: t.Description,
			})
		}
	}

	rootTypes := []struct {
		ref  *struct{ Name string }
		kind string
	}{
		{intro.Data.SchemaData.QueryType, "query"},
		{intro.Data.SchemaData.MutationType, "mutation"},
		{intro.Data.SchemaData.SubscriptionType, "subscription"},
	}

	for _, rt := range rootTypes {
		if rt.ref == nil {
			continue
		}
		if t, ok := typeMap[rt.ref.Name]; ok {
			for _, f := range t.Fields {
				op := model.Operation{
					Name:        f.Name,
					Kind:        rt.kind,
					Description: f.Description,
					Type:        resolveIntroType(f.Type),
				}
				for _, arg := range f.Args {
					op.Args = append(op.Args, model.Field{
						Name: arg.Name,
						Type: resolveIntroType(arg.Type),
					})
				}
				spec.Operations = append(spec.Operations, op)
			}
		}
	}

	return spec, nil
}

type introFullType struct {
	Kind        string `json:"kind"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Fields      []struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Args        []argJSON `json:"args"`
		Type        typeJSON `json:"type"`
	} `json:"fields"`
	EnumValues []struct {
		Name string `json:"name"`
	} `json:"enumValues"`
	InputFields []struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Type        typeJSON `json:"type"`
	} `json:"inputFields"`
}

type argJSON struct {
	Name string  `json:"name"`
	Type typeJSON `json:"type"`
}

type typeJSON struct {
	Kind   string    `json:"kind"`
	Name   string    `json:"name"`
	OfType *typeJSON `json:"ofType"`
}

func resolveIntroType(t typeJSON) string {
	switch t.Kind {
	case "NON_NULL":
		if t.OfType != nil {
			return resolveIntroType(*t.OfType)
		}
		return "interface{}"
	case "LIST":
		if t.OfType != nil {
			return "[]" + resolveIntroType(*t.OfType)
		}
		return "[]interface{}"
	case "SCALAR":
		switch t.Name {
		case "String":
			return "string"
		case "Int":
			return "int"
		case "Float":
			return "float64"
		case "Boolean":
			return "bool"
		case "ID":
			return "string"
		default:
			return t.Name
		}
	default:
		if t.Name != "" {
			return t.Name
		}
		return "interface{}"
	}
}

func fromSchema(schema *ast.Schema, source model.SpecSource) *model.ApiSpec {
	spec := &model.ApiSpec{
		Title:  "GraphQL API",
		Source: source,
	}

	for name, def := range schema.Types {
		if def.BuiltIn || strings.HasPrefix(name, "__") {
			continue
		}
		spec.Schemas = append(spec.Schemas, convertASTType(name, def))
	}

	rootKinds := map[string]string{}
	if schema.Query != nil {
		rootKinds[schema.Query.Name] = "query"
	}
	if schema.Mutation != nil {
		rootKinds[schema.Mutation.Name] = "mutation"
	}
	if schema.Subscription != nil {
		rootKinds[schema.Subscription.Name] = "subscription"
	}

	for _, def := range schema.Types {
		if def.BuiltIn || strings.HasPrefix(def.Name, "__") {
			continue
		}
		kind, isRoot := rootKinds[def.Name]
		if !isRoot {
			continue
		}

		for _, field := range def.Fields {
			op := model.Operation{
				Name:        field.Name,
				Kind:        kind,
				Description: field.Description,
				Type:        fieldTypeName(field.Type),
			}
			for _, arg := range field.Arguments {
				op.Args = append(op.Args, model.Field{
					Name:     arg.Name,
					Type:     fieldTypeName(arg.Type),
					Required: arg.Type.NonNull,
				})
			}
			spec.Operations = append(spec.Operations, op)
		}
	}

	return spec
}

func convertASTType(name string, def *ast.Definition) model.Schema {
	s := model.Schema{
		Name:        name,
		Description: def.Description,
	}

	switch def.Kind {
	case ast.Object, ast.InputObject:
		s.Kind = "struct"
		for _, field := range def.Fields {
			f := model.Field{
				Name:        field.Name,
				Description: field.Description,
				Required:    field.Type.NonNull,
			}
			f.Type = fieldTypeName(field.Type)
			s.Fields = append(s.Fields, f)
		}
	case ast.Enum:
		s.Kind = "enum"
		for _, v := range def.EnumValues {
			s.Values = append(s.Values, v.Name)
		}
	case ast.Interface:
		s.Kind = "interface"
	case ast.Union:
		s.Kind = "union"
	case ast.Scalar:
		s.Kind = "scalar"
	}

	return s
}

func fieldTypeName(t *ast.Type) string {
	if t == nil {
		return "interface{}"
	}

	if t.NonNull && t.Elem != nil {
		return "[]" + mapGraphQLScalar(t.Elem.Name())
	}
	if t.Elem != nil {
		return mapGraphQLScalar(t.Elem.Name())
	}
	return mapGraphQLScalar(t.Name())
}

func mapGraphQLScalar(name string) string {
	switch name {
	case "String", "ID":
		return "string"
	case "Int":
		return "int"
	case "Float":
		return "float64"
	case "Boolean":
		return "bool"
	default:
		return name
	}
}
