# Skiver


## Swagger and code-generation

We use swagger, with [go-swagger](https://goswagger.io/) for both generating
parts of the Swagger 2.0-document, and for providing Go-Models for server-use
as well as typescript-models for frontend.

The base-swagger is extended with code-generation, and can be used to manually 
define parts of the swagger-file.

This information includes base-information like application description, versioning
etc, routing and user-input (parameters).

The generated user-input-models are output into `./models`.

Extra types are generated from the go-structs itself, and lives in the `./types`-package.

### Committing generated files

Although some people prefer not to commit generated files, the generated swagger files with
models etc. should in this project be committed like any other file.

This makes it a lot easier to reason about changes, and we are then not at the mercy of 
code-generation. We still can at any point drop out of using code-generation for parts of, 
or the whole schema.
