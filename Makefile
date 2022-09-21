repo=github.com/runar-rkmedia/skiver
version := $(shell git describe --tags)
gitHash := $(shell git rev-parse --short HEAD)
buildDate := $(shell TZ=UTC date +"%Y-%m-%dT%H:%M:%SZ")
ldflags=-X 'main.version=$(version)' -X 'main.date=$(buildDate)' -X 'main.commit=$(gitHash)' -X 'main.IsDevStr=0'
watch:
	${MAKE} watch_ -j2
watch_: web_watch go_watch
web_watch:
	cd frontend && yarn watch
go_watch:
	fd -e go -e tmpl  | entr -r  sh -c "echo restarting...; go run main.go"
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
		-t registry.fly.io/skiver:latest \
		-t registry.fly.io/skiver:$(version) \
		-t runardocker/skiver-api:latest \
		-t runardocker/skiver-api:$(version) \
		--target scratch \
    --label "org.opencontainers.image.title=skiver-api" \
    --label "org.opencontainers.image.revision=$(gitHash)" \
    --label "org.opencontainers.image.created=$(buildDate)" \
    --label "org.opencontainers.image.version=$(version)" \
    --build-arg "ldflags=$(ldflags)"
	docker build . \
		-t runardocker/skiver-api:grafana \
		-t runardocker/skiver-api:$(version)-grafana \
		--target grafana \
    --label "org.opencontainers.image.title=skiver-api" \
    --label "org.opencontainers.image.revision=$(gitHash)" \
    --label "org.opencontainers.image.created=$(buildDate)" \
    --label "org.opencontainers.image.version=$(version)" \
    --build-arg "ldflags=$(ldflags)"
container-publish: 
	docker push runardocker/skiver-api:latest 
	docker push runardocker/skiver-api:alpine
	docker push runardocker/skiver-api:grafana
	docker push runardocker/skiver-api:$(version) 
	docker push runardocker/skiver-api:$(version)-alpine
	docker push runardocker/skiver-api:$(version)-grafana
	docker push registry.fly.io/skiver:$(version) 
	docker push registry.fly.io/skiver:latest

publish: check-git-clean gen test release container container-publish fly

release: check-git-clean test
	goreleaser release

check-git-clean: 
	@echo "Latest tag on this branch:"
	git describe --tags
	@echo "Latest tag on any branch"
	git describe --tags `git rev-list --tags --max-count=1`
	git describe --exact-match HEAD --tags
	git diff --quiet

build:
	${MAKE} clean
	${MAKE} build-web
	@GOOS=linux   GOARCH=amd64    SUFFIX="-linux-amd64"  ${MAKE} build-api
	@GOOS=darwin                  SUFFIX="-darwin"       ${MAKE} build-api
	@GOOS=windows                 SUFFIX=".exe"         ${MAKE} build-api

	ls -lah dist

# This is kind of stupid, ans should probably be handled by some linter. however, I don't want to look for a linter
# for both go and svelte that accepts custom rules
list-internal:
	@echo "\ninternal-package should only be used in tests"
	@rg 'skiver/internal' --files-with-matches --glob '**/*.go' --glob '!**/*_test*' || echo "All ok for internal"
list-fmtP:
	@echo "\nFmt.Print* is dissallowed"
	@rg 'fmt\.P' --files-with-matches --glob '**/*.go' --glob '!**/*_test*' --glob '!internal/*' --glob '!cmd/*' || echo "All ok for logger.Debug"
list-logger-debug:
	@echo "\nlogger.Debug is dissallowed (bad naming, debug-logging is of course allowed)"
	@rg 'logger\.Debug' --files-with-matches --glob '**/*.go' --glob '!**/*_test*' --glob '!internal/*' --glob '!cmd/*' || echo "All ok for fmt.P*"
list-deprecated:
	@echo "\nhandlers should not use rc.Write"
	@ rg 'rc\.Write' -g '!handlers/{apiHandler,exportHandler}.go' handlers || echo "All ok for deprecated"
list-pre:
	@echo "\nfrontend should not use <pre>-tags"
	@ rg '<pre' frontend/src  -g '!**/*/JsonDetail*' || echo "All ok for <pre>"
list-invalid: list-fmtP list-internal list-deprecated list-logger-debug list-pre
fly:
	./fly.sh .x/skiver-fly.toml
	@echo "Will deploy image 'registry.fly.io/skiver:$(version)' on fly"
	fly deploy --local-only -i "registry.fly.io/skiver:$(version)" --detach
	fly logs
fly_latest: container
	./fly.sh .x/skiver-fly.toml
	@echo "Will deploy image 'registry.fly.io/skiver:latest' on fly"
	fly deploy -i "registry.fly.io/skiver:latest" --detach
	fly logs
