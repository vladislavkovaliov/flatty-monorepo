package model

type SpecSource string

const (
	SourceOpenAPI      SpecSource = "openapi"
	SourceSwagger      SpecSource = "swagger"
	SourceGraphQLSDL   SpecSource = "graphql-sdl"
	SourceGraphQLIntro SpecSource = "graphql-introspection"
)

type ApiSpec struct {
	Title      string
	Version    string
	Source     SpecSource
	Endpoints  []Endpoint
	Schemas    []Schema
	Operations []Operation
}

type Endpoint struct {
	Method      string
	Path        string
	Summary     string
	Description string
	Tags        []string
	Params      []Param
	RequestBody *SchemaRef
	Responses   []Response
}

type Param struct {
	Name        string
	In          string
	Required    bool
	Type        string
	Description string
}

type Response struct {
	StatusCode  string
	Description string
	Schema      *SchemaRef
}

type SchemaRef struct {
	Ref     string
	Schema  *Schema
	IsArray bool
}

type Schema struct {
	Name        string
	Kind        string
	Fields      []Field
	Values      []string
	Description string
}

type Field struct {
	Name        string
	Type        string
	Required    bool
	Description string
	Example     interface{}
	IsArray     bool
}

type Operation struct {
	Name        string
	Kind        string
	Description string
	Args        []Field
	Type        string
}
