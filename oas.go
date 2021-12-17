package main

import (
	"github.com/getkin/kin-openapi/openapi3"
)

const (
	OpenAPIVersion = "3.0.0"

	ContentTypeText = "text/plain"
	ContentTypeJson = "application/json"
	ContentTypeForm = "multipart/form-data"
)

func ApplyScopes(flows *openapi3.OAuthFlows, scopes map[string]string) {
	if flows.Implicit != nil {
		flows.Implicit.Scopes = scopes
	}

	if flows.AuthorizationCode != nil {
		flows.AuthorizationCode.Scopes = scopes
	}

	if flows.Password != nil {
		flows.Password.Scopes = scopes
	}

	if flows.ClientCredentials != nil {
		flows.ClientCredentials.Scopes = scopes
	}
}
