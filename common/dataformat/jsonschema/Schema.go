package jsonschema

import (
	"encoding/json"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/utils"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

type SchemaBuilder struct {
	loader    ResourceLoader
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
	return &SchemaBuilder{loader: &FileResourceLoader{
		RootPath: root,
	}}
}

func FromMap(resourceMap map[string]string) TopLevelSchemaBuilder {
	return &SchemaBuilder{loader: &MapResourceLoader{
		ResourceMap: resourceMap,
	}}
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
	if err := b.addResourceToCompiler(b.topLevel); err != nil {
		return Schema{}, err
	}
	for _, resource := range b.resources {
		if err := b.addResourceToCompiler(resource); err != nil {
			return Schema{}, err
		}
	}

	if topLevelURI, err := joinURI(b.loader.Root(), b.topLevel); err != nil {
		return Schema{}, err
	} else if schema, err := b.compiler.Compile(topLevelURI); err != nil {
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
	if reader, err := b.loader.Load(resource); err != nil {
		return err
	} else if uri, err := joinURI(b.loader.Root(), resource); err != nil {
		return err
	} else if err := b.compiler.AddResource(uri, reader); err != nil {
		return errors.Wrap(err, "Error", "Could not compile schema file")
	}
	return nil
}

type Schema struct {
	schema *jsonschema.Schema
}

func (s Schema) Validate(v interface{}) errors.Error {
	if s.schema == nil {
		// No error thrown here to not interrupt running of tests (which cannot load the schema files)
		return nil
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

type ResourceLoader interface {
	Load(relativePath string) (io.Reader, errors.Error)
	Root() string
}

type FileResourceLoader struct {
	RootPath string
}

func (l *FileResourceLoader) Root() string {
	return l.RootPath
}

func (l *FileResourceLoader) Load(relativePath string) (io.Reader, errors.Error) {
	uri := path.Join(l.RootPath, relativePath)
	if file, err := os.Open(filepath.FromSlash(uri)); err != nil {
		return nil, errors.Wrap(err, "Error", "Could not load schema file")
	} else {
		return file, nil
	}
}

type MapResourceLoader struct {
	ResourceMap map[string]string
}

func (l *MapResourceLoader) Root() string {
	return "resources://root/"
}

func (l *MapResourceLoader) Load(relativePath string) (io.Reader, errors.Error) {
	if content, ok := l.ResourceMap[relativePath]; !ok {
		return nil, errors.New("Error", "Resource not found: %s", relativePath)
	} else {
		return strings.NewReader(content), nil
	}
}

func joinURI(parts ...string) (string, errors.Error) {
	p, err := url.Parse(parts[0])
	if err != nil {
		return "", errors.Wrap(err, "Error", "Could not parse schema URI")
	}
	parts[0] = p.Path
	p.Path = path.Join(parts...)
	return p.String(), nil
}
