repo=github.com/runar-rkmedia/skiver
version := $(shell git describe --tags)
gitHash := $(shell git rev-parse --short HEAD)
buildDate := $(shell TZ=UTC date +"%Y-%m-%dT%H:%M:%S")
ldflags=-X 'main.Version=$(version)' -X 'main.BuildDateStr=$(buildDate)' -X 'main.GitHash=$(gitHash)' -X 'main.IsDevStr=0'
watch:
	cd frontend && yarn watch &
	fd -e go  | entr -r  sh -c "echo restarting...; go run main.go"
gen:
	go generate ./...
build-api:
	go build -ldflags="${ldflags}" -o dist/skiver${SUFFIX} main.go
clean:
	rm -rf dist
	rm -rf frontend/dist
test:
	go test ./...
lint:
	golangci-lint run
test-watch:
	fd -e go | entr -r sh -c 'printf "%*s\n" "${COLUMNS:-$(tput cols)}" "" | tr " " - && gotest ./... | grep -v "no test files"'
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
build:
	${MAKE} clean
	${MAKE} build-web
	@GOOS=linux   GOARCH=amd64    SUFFIX="-linux-amd64"  ${MAKE} build-api
	@GOOS=darwin                  SUFFIX="-darwin"       ${MAKE} build-api
	@GOOS=windows                 SUFFIX=".exe"         ${MAKE} build-api

	ls -lah dist
