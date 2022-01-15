# Skiver

[[toc]]

## Current feature-set

- [X] Multi-locale support
- [X] Multi-project support
- [X] i18n-compliant export
- [X] Auto-translate via external translation-service (Bing Translate, Libre Translate)
- [X] Report missing translation
- [ ] Import translations
  - [X] General AST
  - [X] Dry run, with preview of updates and creations
  - [ ] `i18next`-format
    - [X] multiple-language 
    - [X] context-support
    - [ ] inferring of variables and nested keys. 
- [ ] Multi-organization support

## Planned feature-set

- [ ] Variables with examples to help translators and developers.
- [ ] Server-side interpolation via API
- [ ] Client-side live interpolation via library
- [ ] Source-code integration with project, to show usage of translation.
      
     E.g.
     
     > This translation is used in SuperComponent.svelte:74:
       ```jsx
       73: <div>
       74:   <p>{t("feature.awesome", {count: 6})}
       75: </div>
       ```
- [ ] Upload of images to show usage of translation.
- [ ] Sharing of translations between projects.
- [ ] Typescript-type-generation with rich comments


## Things that are a mess, and need refactoring

- Missing translations. They work, but they are terrible. It comes mostly from not trusting the user-input, 
while still attempting to resolve this untrusted information.


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



