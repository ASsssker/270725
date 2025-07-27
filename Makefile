## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

## gen-bp: generate boilerplate files
.PHONY: gen-bp
gen-bp:
	oapi-codegen -package boilerplate -generate echo -o internal/rest/v1/boileplate/server.go api/openapi.yaml
	oapi-codegen -package boilerplate -generate spec -o internal/rest/v1/boileplate/spec.go api/openapi.yaml
	oapi-codegen -package boilerplate -generate types -o internal/rest/v1/boileplate/types.go api/openapi.yaml
	oapi-codegen -package boilerplate -generate client -o internal/rest/v1/boileplate/client.go api/openapi.yaml
