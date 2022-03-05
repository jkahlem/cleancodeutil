package jsonschema

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/utils"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

type SchemaBuilder struct {
	root      string
	topLevel  string
	resources []string
	compiler  *jsonschema.Compiler
}

type TopLevelSchemaBuilder interface {
	WithTopLevel(resourceUri string) ResourcesSchemaBuilder
}

type ResourcesSchemaBuilder interface {
	WithResources(resourceUris ...string) ResourcesSchemaBuilder
	Compile() (Schema, errors.Error)
	MustCompile() Schema
}

type TopLevel string

func AtRoot(root string) TopLevelSchemaBuilder {
	return &SchemaBuilder{root: root}
}

func (b *SchemaBuilder) WithTopLevel(resourceUri string) ResourcesSchemaBuilder {
	b.topLevel = resourceUri
	return b
}

func (b *SchemaBuilder) WithResources(resourceUris ...string) ResourcesSchemaBuilder {
	b.resources = resourceUris
	return b
}

func (b *SchemaBuilder) Compile() (Schema, errors.Error) {
	b.compiler = jsonschema.NewCompiler()
	b.compiler.Draft = jsonschema.Draft2020
	b.addResourceToCompiler(b.topLevel)
	for _, resource := range b.resources {
		b.addResourceToCompiler(resource)
	}
	if schema, err := b.compiler.Compile(path.Join(b.root, string(b.topLevel))); err != nil {
		return Schema{}, errors.Wrap(err, "Error", "Could not compile schema files")
	} else {
		return Schema{
			schema: schema,
		}, nil
	}
}

func (b *SchemaBuilder) MustCompile() Schema {
	if schema, err := b.Compile(); err != nil {
		panic(err)
	} else {
		return schema
	}
}

func (b *SchemaBuilder) addResourceToCompiler(resource string) errors.Error {
	uri := path.Join(b.root, resource)
	if file, err := os.Open(filepath.FromSlash(uri)); err != nil {
		return errors.Wrap(err, "Error", "Could not load schema file")
	} else if err := b.compiler.AddResource("file:///c:/Users/work/Documents/bachelor/02-project/returntypes-predictor/mainapp/schemas/", file); err != nil {
		return errors.Wrap(err, "Error", "Could not compile schema file")
	}
	return nil
}

type Schema struct {
	schema *jsonschema.Schema
}

func (s Schema) Validate(v interface{}) errors.Error {
	if s.schema == nil {
		return errors.New("Error", "No json schema available to validate against")
	}
	if err := s.schema.Validate(v); err != nil {
		return errors.Wrap(err, "Error", "Validation against json schema has failed")
	}
	return nil
}

func UnmarshalJSONStrict(source []byte, destination interface{}, schema Schema) errors.Error {
	var v interface{}
	if err := json.Unmarshal(source, &v); err != nil {
		return errors.Wrap(err, "JSON Error", "Could not unmarshal JSON")
	}
	if err := schema.Validate(v); err != nil {
		return err
	}
	if err := utils.DecodeMapToStructStrict(v, destination); err != nil {
		return err
	}
	return nil
}
