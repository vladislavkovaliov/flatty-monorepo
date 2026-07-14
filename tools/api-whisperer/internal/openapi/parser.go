package openapi

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"

	"github.com/vladislavkovaliov/api-whisperer/internal/model"
)

func Parse(path string) (*model.ApiSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	doc3, err := openapi3.NewLoader().LoadFromData(data)
	if err == nil {
		return fromOpenAPI3(doc3), nil
	}

	var doc2 openapi2.T
	if err := yaml.Unmarshal(data, &doc2); err != nil {
		if err := json.Unmarshal(data, &doc2); err != nil {
			return nil, fmt.Errorf("parse as openapi2: %w", err)
		}
	}

	if doc2.Swagger == "" {
		return nil, fmt.Errorf("not a valid OpenAPI or Swagger document")
	}

	doc3, err = openapi2conv.ToV3(&doc2)
	if err != nil {
		return nil, fmt.Errorf("convert swagger to v3: %w", err)
	}

	spec := fromOpenAPI3(doc3)
	spec.Source = model.SourceSwagger
	return spec, nil
}

func ParseURL(rawURL string) (*model.ApiSpec, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	doc3, err := openapi3.NewLoader().LoadFromURI(parsedURL)
	if err != nil {
		return nil, fmt.Errorf("load from url: %w", err)
	}

	return fromOpenAPI3(doc3), nil
}

func fromOpenAPI3(doc *openapi3.T) *model.ApiSpec {
	spec := &model.ApiSpec{
		Title:   doc.Info.Title,
		Version: doc.Info.Version,
		Source:  model.SourceOpenAPI,
	}

	schemas := make(map[string]*openapi3.Schema)
	if doc.Components != nil {
		for name, ref := range doc.Components.Schemas {
			if ref != nil {
				schemas[name] = ref.Value
			}
		}
	}

	for name, schema := range schemas {
		spec.Schemas = append(spec.Schemas, convertSchema(name, schema))
	}

	for _, path := range doc.Paths.InMatchingOrder() {
		pathItem := doc.Paths.Find(path)
		if pathItem == nil {
			continue
		}

		ops := []struct {
			method string
			op     *openapi3.Operation
		}{
			{"GET", pathItem.Get},
			{"POST", pathItem.Post},
			{"PUT", pathItem.Put},
			{"DELETE", pathItem.Delete},
			{"PATCH", pathItem.Patch},
			{"HEAD", pathItem.Head},
			{"OPTIONS", pathItem.Options},
		}

		for _, entry := range ops {
			if entry.op == nil {
				continue
			}
			spec.Endpoints = append(spec.Endpoints, convertEndpoint(entry.method, path, entry.op, schemas))
		}
	}

	seen := make(map[string]bool)
	for _, s := range spec.Schemas {
		seen[s.Name] = true
	}
	for _, ep := range spec.Endpoints {
		if ep.RequestBody != nil && ep.RequestBody.Schema != nil {
			name := ep.RequestBody.Schema.Name
			if !seen[name] {
				spec.Schemas = append(spec.Schemas, *ep.RequestBody.Schema)
				seen[name] = true
			}
		}
		for _, r := range ep.Responses {
			if r.Schema != nil && r.Schema.Schema != nil {
				name := r.Schema.Schema.Name
				if !seen[name] {
					spec.Schemas = append(spec.Schemas, *r.Schema.Schema)
					seen[name] = true
				}
			}
		}
	}

	return spec
}

func convertEndpoint(method, path string, op *openapi3.Operation, schemas map[string]*openapi3.Schema) model.Endpoint {
	ep := model.Endpoint{
		Method:      method,
		Path:        path,
		Summary:     op.Summary,
		Description: op.Description,
		Tags:        op.Tags,
	}

	for _, p := range op.Parameters {
		if p == nil || p.Value == nil {
			continue
		}
		paramType := typeString(nil)
		if p.Value.Schema != nil {
			paramType = typeString(p.Value.Schema.Value)
		}
		ep.Params = append(ep.Params, model.Param{
			Name:        p.Value.Name,
			In:          p.Value.In,
			Required:    p.Value.Required,
			Type:        paramType,
			Description: p.Value.Description,
		})
	}

	if op.RequestBody != nil && op.RequestBody.Value != nil {
		for _, media := range op.RequestBody.Value.Content {
			if media.Schema != nil {
				schemaName := fmt.Sprintf("%s %s Request", method, path)
				ep.RequestBody = convertSchemaRef(schemaName, media.Schema, schemas)
				break
			}
		}
	}

	if op.Responses != nil {
		for code, respRef := range op.Responses.Map() {
			if respRef == nil || respRef.Value == nil {
				continue
			}
			desc := ""
			if respRef.Value.Description != nil {
				desc = *respRef.Value.Description
			}
			r := model.Response{
				StatusCode:  code,
				Description: desc,
			}
			for _, media := range respRef.Value.Content {
				if media.Schema != nil {
					schemaName := fmt.Sprintf("%s %s Response %s", method, path, code)
					r.Schema = convertSchemaRef(schemaName, media.Schema, schemas)
					break
				}
			}
			ep.Responses = append(ep.Responses, r)
		}
	}

	return ep
}

func convertSchema(name string, s *openapi3.Schema) model.Schema {
	sch := model.Schema{
		Name: name,
	}

	if s == nil {
		return sch
	}

	sch.Description = s.Description

	switch {
	case s.Type != nil && s.Type.Is("object"):
		sch.Kind = "struct"
		for propName, propRef := range s.Properties {
			if propRef == nil || propRef.Value == nil {
				continue
			}
			f := model.Field{
				Name:        propName,
				Description: propRef.Value.Description,
				Required:    contains(s.Required, propName),
			}
			f.Type, f.IsArray = resolveType(propRef.Value)
			sch.Fields = append(sch.Fields, f)
		}
	case len(s.Enum) > 0:
		sch.Kind = "enum"
		for _, v := range s.Enum {
			sch.Values = append(sch.Values, fmt.Sprintf("%v", v))
		}
	default:
		if s.Type != nil {
			sch.Kind = s.Type.Slice()[0]
		} else {
			sch.Kind = "object"
		}
	}

	return sch
}

func convertSchemaRef(name string, ref *openapi3.SchemaRef, schemas map[string]*openapi3.Schema) *model.SchemaRef {
	r := &model.SchemaRef{}
	if ref == nil {
		return r
	}

	if ref.Value != nil && ref.Value.Type != nil {
		r.IsArray = ref.Value.Type.Is("array")
	}

	if ref.Ref != "" {
		r.Ref = resolveRefName(ref.Ref)
		if s, ok := schemas[r.Ref]; ok {
			sch := convertSchema(r.Ref, s)
			r.Schema = &sch
		}
		return r
	}

	if r.IsArray && ref.Value.Items != nil {
		if ref.Value.Items.Ref != "" {
			itemRef := resolveRefName(ref.Value.Items.Ref)
			r.Ref = itemRef
			if s, ok := schemas[itemRef]; ok {
				sch := convertSchema(itemRef, s)
				r.Schema = &sch
			}
			return r
		}
	}

	sch := convertSchema(name, ref.Value)
	r.Schema = &sch
	return r
}

func resolveType(s *openapi3.Schema) (goType string, isArray bool) {
	if s == nil {
		return "interface{}", false
	}
	if s.Type != nil && s.Type.Is("array") && s.Items != nil {
		t, _ := resolveType(s.Items.Value)
		return t, true
	}

	switch {
	case s.Type == nil:
		ref := s.OneOf
		if len(ref) > 0 {
			return "interface{}", false
		}
		return "map[string]interface{}", false
	case s.Type.Is("string"):
		return "string", false
	case s.Type.Is("integer"):
		return "int", false
	case s.Type.Is("number"):
		return "float64", false
	case s.Type.Is("boolean"):
		return "bool", false
	case s.Type.Is("object"):
		return "map[string]interface{}", false
	default:
		return "interface{}", false
	}
}

func typeString(s *openapi3.Schema) string {
	if s == nil || s.Type == nil {
		return "string"
	}
	return s.Type.Slice()[0]
}

func resolveRefName(ref string) string {
	parts := strings.Split(ref, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ref
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
