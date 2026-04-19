package gen

//go:generate oapi-codegen --config ../oapi-codegen-types.yaml ../api/openapi.yaml
//go:generate oapi-codegen --config ../oapi-codegen-server.yaml ../api/openapi.yaml
//go:generate oapi-codegen --config ../oapi-codegen-client.yaml ../api/openapi.yaml
