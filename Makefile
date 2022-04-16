repo=github.com/runar-rkmedia/skiver
version := $(shell git describe --tags)
gitHash := $(shell git rev-parse --short HEAD)
buildDate := $(shell TZ=UTC date +"%Y-%m-%dT%H:%M:%SZ")
ldflags=-X 'main.version=$(version)' -X 'main.date=$(buildDate)' -X 'main.commit=$(gitHash)' -X 'main.IsDevStr=0'
clildflags=-X 'github.com/runar-rkmedia/skiver/cli/cmd.version=$(version)' -X 'github.com/runar-rkmedia/skiver/cli/cmd.date=$(buildDate)' -X 'github.com/runar-rkmedia/skiver/cli/cmd.commit=$(gitHash)' -X 'github.com/runar-rkmedia/skiver/cli/cmd.IsDevStr=0'
watch:
	cd frontend && yarn watch &
	fd -e go -e tmpl  | entr -r  sh -c "echo restarting...; go run main.go"
	wait
gen:
	go generate ./...
build-api:
	go build -ldflags="${ldflags}" -o dist/skiver${SUFFIX} main.go
build-cli:
	go build -ldflags="${clildflags}" -o dist/skiver-cli${SUFFIX} cli/main.go
clean:
	rm -rf dist
	rm -rf frontend/dist
test:
	go test ./...
lint:
	golangci-lint run
test-watch:
	fd -e go -e tmpl | entr -r sh -c 'printf "%*s\n" "${COLUMNS:-$(tput cols)}" "" | tr " " - && go test ./... | grep -v "no test files"'
swagger-clean:
	echo "removing files generated by swagger"
	rg -tgo 'Code generated by go-swagger' models --files-with-matches &&  rg -tgo 'Code generated by go-swagger' models --files-with-matches | xargs rm || echo "no cleanup needed"
swagger-watch:
	echo "Watch the base-swagger and the types-folder"
	printf "base-swagger.yml\n$(fd '' types)" | entr -r go generate ./...
build-web:
	cd frontend && yarn build
generate:
	${MAKE} swagger-clean
	go generate ./...
	swagger validate swagger.yml
container: build-web
	docker build . \
		-t runardocker/skiver-api:alpine \
		-t runardocker/skiver-api:$(version)-alpine \
		--target alpine \
    --label "org.opencontainers.image.title=skiver-api" \
    --label "org.opencontainers.image.revision=$(gitHash)" \
    --label "org.opencontainers.image.created=$(buildDate)" \
    --label "org.opencontainers.image.version=$(version)" \
    --build-arg "ldflags=$(ldflags)"
	docker build . \
		-t runardocker/skiver-api:latest \
		-t runardocker/skiver-api:$(version) \
		--target scratch \
    --label "org.opencontainers.image.title=skiver-api" \
    --label "org.opencontainers.image.revision=$(gitHash)" \
    --label "org.opencontainers.image.created=$(buildDate)" \
    --label "org.opencontainers.image.version=$(version)" \
    --build-arg "ldflags=$(ldflags)"
container-publish: 
	docker push runardocker/skiver-api:latest 
	docker push runardocker/skiver-api:alpine
	docker push runardocker/skiver-api:$(version) 
	docker push runardocker/skiver-api:$(version)-alpine

publish: container container-publish release

release:
	goreleaser release

build:
	${MAKE} clean
	${MAKE} build-web
	@GOOS=linux   GOARCH=amd64    SUFFIX="-linux-amd64"  ${MAKE} build-api
	@GOOS=darwin                  SUFFIX="-darwin"       ${MAKE} build-api
	@GOOS=windows                 SUFFIX=".exe"         ${MAKE} build-api

	ls -lah dist

list-internal:
	@rg 'skiver/internal' --files-with-matches --glob '**/*.go' --glob '!**/*_test*' || echo "All ok for internal"
list-fmtP:
	@rg 'fmt\.P' --files-with-matches --glob '**/*.go' --glob '!**/*_test*' --glob '!cli/*' --glob '!internal/*' --glob '!cmd/*' || echo "All ok for fmt.P*"
list-invalid: list-fmtP list-internal


