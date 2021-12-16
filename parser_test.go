package main

import (
	"encoding/json"
	"go/ast"
	"go/token"
	"io/ioutil"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

func setupParser() (*parser, error) {
	return newParser("example/", "example/main.go", "", "", false, false, false)
}
func TestExample(t *testing.T) {
	p, err := setupParser()
	require.NoError(t, err)

	err = p.parse()
	require.NoError(t, err)

	bts, err := json.MarshalIndent(p.OpenAPI, "", "    ")
	require.NoError(t, err)

	expected, _ := ioutil.ReadFile("./example/example.json")
	require.JSONEq(t, string(expected), string(bts))
}

func TestShowHiddenExample(t *testing.T) {
	p, err := newParser("example/", "example/main.go", "", "", false, false, true)
	require.NoError(t, err)

	err = p.parse()
	require.NoError(t, err)

	bts, err := json.MarshalIndent(p.OpenAPI, "", "    ")
	require.NoError(t, err)

	expected, _ := ioutil.ReadFile("./example/example-show-hidden.json")
	require.JSONEq(t, string(expected), string(bts))
}

func TestDeterministic(t *testing.T) {
	var allOutputs []string
	for i := 0; i < 10; i++ {
		p, err := setupParser()
		require.NoError(t, err)

		err = p.parse()
		require.NoError(t, err)

		bts, err := json.Marshal(p.OpenAPI)
		require.NoError(t, err)
		allOutputs = append(allOutputs, string(bts))
	}

	for i := 0; i < len(allOutputs)-1; i++ {
		require.Equal(t, allOutputs[i], allOutputs[i+1])
	}
}

func Test_parseRouteComment(t *testing.T) {
	p, err := setupParser()
	require.NoError(t, err)

	operation := &OperationObject{
		Responses: map[string]*ResponseObject{},
	}
	p.OpenAPI.Paths["v2/foo/bar"] = &PathItemObject{}
	p.OpenAPI.Paths["v2/foo/bar"].Get = operation

	duplicateError := p.parseRouteComment(operation, "@Router v2/foo/bar [get]")
	require.Error(t, duplicateError)
}

func Test_infoDescriptionRef(t *testing.T) {
	p, err := setupParser()
	require.NoError(t, err)
	p.OpenAPI.Info.Description = &ReffableString{Value: "$ref:http://dopeoplescroll.com/"}

	result, err := json.Marshal(p.OpenAPI.Info.Description)

	require.NoError(t, err)
	require.Equal(t, "{\"$ref\":\"http://dopeoplescroll.com/\"}", string(result))
}

func Test_parseTags(t *testing.T) {
	t.Run("name", func(t *testing.T) {
		result, err := parseTags("@Tags \"Foo\"")

		require.NoError(t, err)
		require.Equal(t, &TagDefinition{Name: "Foo"}, result)
	})

	t.Run("name and description", func(t *testing.T) {
		result, err := parseTags("@Tags \"Foobar\" \"Barbaz\"")

		require.NoError(t, err)
		require.Equal(t, &TagDefinition{Name: "Foobar", Description: &ReffableString{Value: "Barbaz"}}, result)
	})

	t.Run("name and description including ref ", func(t *testing.T) {
		result, err := parseTags("@Tags \"Foobar\" \"$ref:path/to/baz\"")
		require.NoError(t, err)
		b, err := json.Marshal(result)
		require.NoError(t, err)
		require.Equal(t, "{\"name\":\"Foobar\",\"description\":{\"$ref\":\"path/to/baz\"}}", string(b))
	})

	t.Run("invalid tag", func(t *testing.T) {
		_, err := parseTags("@Tags Foobar Barbaz")

		require.Error(t, err)
	})
}

func Test_parseCliGroups(t *testing.T) {
	t.Run("group and aliases", func(t *testing.T) {
		// we only expect the value here, not the whole comment (omitting "@CliGroups")
		group, cfg, err := parseCliGroups("feature-flags aliases:flags,featureflags")
		require.NoError(t, err)
		require.Equal(t, "feature-flags", group)
		require.Equal(t, CliConfigObject{Aliases: []string{"flags", "featureflags"}}, cfg)
	})

	t.Run("group only", func(t *testing.T) {
		group, cfg, err := parseCliGroups("feature-flags")
		require.NoError(t, err)
		require.Equal(t, "feature-flags", group)
		require.Equal(t, CliConfigObject{}, cfg)
	})

	t.Run("missing aliases label", func(t *testing.T) {
		_, _, err := parseCliGroups("feature-flags flags,featureflags")
		require.Error(t, err)
		require.Equal(t, "Expected: @CliGroups <command> (optional) aliases:<alias1,alias2,etc> Received: @CliGroups feature-flags flags,featureflags. Did you forget the \"aliases\" label?", err.Error())
	})

	t.Run("too many spaces", func(t *testing.T) {
		_, _, err := parseCliGroups("projects aliases: proj,project")
		require.Error(t, err)
	})

	t.Run("empty value", func(t *testing.T) {
		_, _, err := parseCliGroups("")
		require.Error(t, err)
	})
}

func Test_handleCompoundType(t *testing.T) {
	t.Run("oneOf", func(t *testing.T) {
		p, err := setupParser()
		require.NoError(t, err)
		result, err := p.handleCompoundType("./example", "example.com/example", "oneOf(string,[]string)")
		require.NoError(t, err)
		s, err := json.Marshal(result)
		require.NoError(t, err)
		require.Equal(t, "{\"oneOf\":[{\"type\":\"string\"},{\"type\":\"array\",\"items\":{\"type\":\"string\"}}]}", string(s))
	})

	t.Run("anyOf", func(t *testing.T) {
		p, err := setupParser()
		require.NoError(t, err)
		result, err := p.handleCompoundType("./example", "example.com/example", "anyOf(string,[]string)")
		require.NoError(t, err)
		s, err := json.Marshal(result)
		require.NoError(t, err)
		require.Equal(t, "{\"anyOf\":[{\"type\":\"string\"},{\"type\":\"array\",\"items\":{\"type\":\"string\"}}]}", string(s))
	})

	t.Run("allOf", func(t *testing.T) {
		p, err := setupParser()
		require.NoError(t, err)
		result, err := p.handleCompoundType("./example", "example.com/example", "allOf(string,[]string)")
		require.NoError(t, err)
		s, err := json.Marshal(result)
		require.NoError(t, err)
		require.Equal(t, "{\"allOf\":[{\"type\":\"string\"},{\"type\":\"array\",\"items\":{\"type\":\"string\"}}]}", string(s))
	})

	t.Run("case insensitive oneOf", func(t *testing.T) {
		p, err := setupParser()
		require.NoError(t, err)
		result, err := p.handleCompoundType("./example", "example.com/example", "oneof(string,[]string)")
		require.NoError(t, err)
		s, err := json.Marshal(result)
		require.NoError(t, err)
		require.Equal(t, "{\"oneOf\":[{\"type\":\"string\"},{\"type\":\"array\",\"items\":{\"type\":\"string\"}}]}", string(s))
	})

	t.Run("case insensitive anyOf", func(t *testing.T) {
		p, err := setupParser()
		require.NoError(t, err)
		result, err := p.handleCompoundType("./example", "example.com/example", "anyof(string,[]string)")
		require.NoError(t, err)
		s, err := json.Marshal(result)
		require.NoError(t, err)
		require.Equal(t, "{\"anyOf\":[{\"type\":\"string\"},{\"type\":\"array\",\"items\":{\"type\":\"string\"}}]}", string(s))
	})

	t.Run("case insensitive allOf", func(t *testing.T) {
		p, err := setupParser()
		require.NoError(t, err)
		result, err := p.handleCompoundType("./example", "example.com/example", "allof(string,[]string)")
		require.NoError(t, err)
		s, err := json.Marshal(result)
		require.NoError(t, err)
		require.Equal(t, "{\"allOf\":[{\"type\":\"string\"},{\"type\":\"array\",\"items\":{\"type\":\"string\"}}]}", string(s))
	})

	t.Run("not", func(t *testing.T) {
		p, err := setupParser()
		require.NoError(t, err)
		result, err := p.handleCompoundType("./example", "example.com/example", "not(string)")
		require.NoError(t, err)
		s, err := json.Marshal(result)
		require.NoError(t, err)
		require.Equal(t, "{\"not\":{\"type\":\"string\"}}", string(s))
	})

	t.Run("handles whitespace", func(t *testing.T) {
		p, err := setupParser()
		require.NoError(t, err)
		result, err := p.handleCompoundType("./example", "example.com/example", "allOf(  string, []string )")
		require.NoError(t, err)
		s, err := json.Marshal(result)
		require.NoError(t, err)
		require.Equal(t, "{\"allOf\":[{\"type\":\"string\"},{\"type\":\"array\",\"items\":{\"type\":\"string\"}}]}", string(s))
	})

	t.Run("not only accepts 1 arg", func(t *testing.T) {
		p, err := setupParser()
		require.NoError(t, err)
		_, notErr := p.handleCompoundType("./example", "example.com/example", "not(string,int32)")
		require.Error(t, notErr)
	})

	t.Run("error when no args", func(t *testing.T) {
		p, err := setupParser()
		require.NoError(t, err)
		_, notErr := p.handleCompoundType("./example", "example.com/example", "oneOf()")
		require.Error(t, notErr)
	})
}

func Test_explodeRefs(t *testing.T) {
	t.Run("Info.Description unchanged when not a ref", func(t *testing.T) {
		p, err := setupParser()
		require.NoError(t, err)
		p.OpenAPI.Info.Description = &ReffableString{Value: "Foo"}

		err = p.explodeRefs()
		require.NoError(t, err)

		require.Equal(t, "Foo", p.OpenAPI.Info.Description.Value)
	})

	t.Run("Info.Description inlined when a ref", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("GET", "https://example.com",
			httpmock.NewStringResponder(200, "The quick brown fox jumped over the lazy dog"))
		p, err := setupParser()
		require.NoError(t, err)
		p.OpenAPI.Info.Description = &ReffableString{Value: "$ref:https://example.com"}

		err = p.explodeRefs()
		require.NoError(t, err)

		require.Equal(t, "The quick brown fox jumped over the lazy dog", p.OpenAPI.Info.Description.Value)
	})

	t.Run("Tags[].Description unchanged when not a ref", func(t *testing.T) {
		p, err := setupParser()
		require.NoError(t, err)
		p.OpenAPI.Tags = []TagDefinition{{Name: "Foo", Description: &ReffableString{Value: "Foobar"}}}

		err = p.explodeRefs()
		require.NoError(t, err)

		require.Equal(t, "Foobar", p.OpenAPI.Tags[0].Description.Value)
	})

	t.Run("Tags[].Description inlined when a ref", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("GET", "https://example.com",
			httpmock.NewStringResponder(200, "The quick brown fox jumped over the lazy dog"))
		p, err := setupParser()
		require.NoError(t, err)
		p.OpenAPI.Tags = []TagDefinition{{Name: "Foo", Description: &ReffableString{Value: "$ref:https://example.com"}}}

		err = p.explodeRefs()
		require.NoError(t, err)

		require.Equal(t, "The quick brown fox jumped over the lazy dog", p.OpenAPI.Tags[0].Description.Value)
	})

	t.Run("Mixed of tag refs and non-refs", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("GET", "https://example.com",
			httpmock.NewStringResponder(200, "The quick brown fox jumped over the lazy dog"))
		p, err := setupParser()
		require.NoError(t, err)
		p.OpenAPI.Tags = []TagDefinition{{Name: "Foo", Description: &ReffableString{Value: "$ref:https://example.com"}}, {Name: "Bar", Description: &ReffableString{Value: "Baz"}}}

		err = p.explodeRefs()
		require.NoError(t, err)

		require.Equal(t, "The quick brown fox jumped over the lazy dog", p.OpenAPI.Tags[0].Description.Value)
		require.Equal(t, "Baz", p.OpenAPI.Tags[1].Description.Value)
	})
}

func Test_fetchRef(t *testing.T) {
	t.Run("fetches local file ref", func(t *testing.T) {
		desc, err := fetchRef(".", "$ref:file://example/example.md")
		require.NoError(t, err)

		require.Equal(t, "Example description", desc)
	})
}

func Test_descriptions(t *testing.T) {
	t.Run("Description unchanged when not a ref", func(t *testing.T) {
		p, err := setupParser()
		require.NoError(t, err)

		operation := &OperationObject{
			Responses: map[string]*ResponseObject{},
		}

		err = p.parseDescription(operation, "testing")
		require.NoError(t, err)

		require.Equal(t, "testing", operation.Description)
	})

	t.Run("Description inline when a ref", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder("GET", "https://example.com",
			httpmock.NewStringResponder(200, "The quick brown fox jumped over the lazy dog"))

		p, err := setupParser()
		require.NoError(t, err)

		operation := &OperationObject{
			Responses: map[string]*ResponseObject{},
		}

		err = p.parseDescription(operation, "$ref:https://example.com")
		require.NoError(t, err)

		require.Equal(t, "The quick brown fox jumped over the lazy dog", operation.Description)
	})
}

func Test_parseRequestBodyExample(t *testing.T) {
	t.Run("Parses example object request body", func(t *testing.T) {
		exampleRequestBody, err := parseRequestBodyExample("{\\\"name\\\":\\\"Bilbo\\\"}")
		require.NoError(t, err)

		require.Equal(t, map[string]interface{}(map[string]interface{}{"name": "Bilbo"}), exampleRequestBody)
	})

	t.Run("Parses example array request body", func(t *testing.T) {
		exampleRequestBody, err := parseRequestBodyExample("[{\\\"name\\\":\\\"Bilbo\\\"}]")
		require.NoError(t, err)

		require.Equal(t, []interface{}([]interface{}{map[string]interface{}{"name": "Bilbo"}}), exampleRequestBody)
	})

	t.Run("Errors if example is invalid", func(t *testing.T) {
		_, err := parseRequestBodyExample("{name:\\\"Smaug\\\"}")
		require.Error(t, err)
	})
}

func Test_genSchemaObjectID(t *testing.T) {
	t.Run("empty package name", func(t *testing.T) {
		p, err := setupParser()
		require.NoError(t, err)

		result := p.genSchemaObjectID("", "sample")

		require.Equal(t, "sample", string(result))
	})
	t.Run("simple package name", func(t *testing.T) {
		p, err := setupParser()
		require.NoError(t, err)

		result := p.genSchemaObjectID("sample", "sample")

		require.Equal(t, "sample.sample", string(result))
	})
	t.Run("multidepth package name", func(t *testing.T) {
		p, err := setupParser()
		require.NoError(t, err)

		result := p.genSchemaObjectID("test.sample", "sample")

		require.Equal(t, "test.sample.sample", string(result))
	})
	t.Run("omit package name", func(t *testing.T) {
		p, err := newParser("example/", "example/main.go", "", "", false, true, false)
		require.NoError(t, err)

		result := p.genSchemaObjectID("test.sample", "sample")

		require.Equal(t, "sample", string(result))
	})
}

func Test_parseOperationTags(t *testing.T) {
	t.Run("Parses operation tags", func(t *testing.T) {
		p, err := setupParser()
		require.NoError(t, err)

		p.OpenAPI.Tags = append(p.OpenAPI.Tags, TagDefinition{Name: "foo", Description: &ReffableString{Value: "bar"}})

		var comment []*ast.Comment
		comment = append(comment, &ast.Comment{Slash: 0, Text: "// @Tag foo"})
		err = p.parseOperation(p.ModulePath, "", comment)
		require.NoError(t, err)
	})

	t.Run("Errors when tag in operation is not in list of tags", func(t *testing.T) {
		p, err := setupParser()
		require.NoError(t, err)

		p.OpenAPI.Tags = append(p.OpenAPI.Tags, TagDefinition{Name: "foo", Description: &ReffableString{Value: "bar"}})

		var comment []*ast.Comment
		comment = append(comment, &ast.Comment{Slash: 0, Text: "// @Tag Foo"})
		err = p.parseOperation(p.ModulePath, "", comment)
		require.Error(t, err)
	})
}

func Test_validateSchemaNames(t *testing.T) {
	t.Run("Returns no conflicts", func(t *testing.T) {
		p, err := setupParser()
		require.NoError(t, err)

		conflicts := p.validateSchemaNames()

		require.Empty(t, conflicts)
	})

	t.Run("Returns conflicts", func(t *testing.T) {
		p, err := setupParser()
		require.NoError(t, err)
		p.ApiSchemaNames["pkg/foo/bar"] = map[string]string{}
		p.ApiSchemaNames["pkg/foo/bar"]["BarRecord"] = "Record"
		p.ApiSchemaNames["pkg/baz/qux"] = map[string]string{}
		p.ApiSchemaNames["pkg/baz/qux"]["QuxRecord"] = "Record"

		conflicts := p.validateSchemaNames()

		require.Len(t, conflicts, 1)
		require.Contains(t, conflicts[0], "pkg/foo/bar#BarRecord")
		require.Contains(t, conflicts[0], "pkg/baz/qux#QuxRecord")
	})
}

func Test_parseOverrideStructTag(t *testing.T) {
	t.Run("found tag", func(t *testing.T) {
		ast := &ast.Field{
			Doc:   nil,
			Names: nil,
			Type:  nil,
			Tag: &ast.BasicLit{
				ValuePos: 0,
				Kind:     token.STRING,
				Value:    `overrideApiSchemaType:"Test"`},
		}
		result := parseOverrideStructTag(ast)

		require.Equal(t, "Test", result)
	})
}
