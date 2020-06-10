BUILD_FOLDER = ".build-artifacts"

ensure-bin:
	[ -d .bin ] || mkdir .bin

setup: ensure-bin ## Install tools
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | bash -s v1.27.0
	mkdir .bin || mv bin/golangci-lint .bin/golangci-lint && rm -rf bin

lint: ## Run the linters
	golangci-lint run

fmt: ## gofmt and goimports all go files
	find . -name '*.go' -not -wholename './vendor/*' | while read -r file; do gofmt -w -s "$$file"; goimports -w "$$file"; done

run:
	docker-compose up
	
build:
	go version
	go env
	mkdir -p $(BUILD_FOLDER)
	go build -o $(BUILD_FOLDER)/dnsserver dnsserver/cmd/main.go
	chmod a+rwx -R $(BUILD_FOLDER)

ci: # todo
	echo "YOYO"

functional-tests:
	go run .test/main.go 
	
local-dev:
	go mod edit -replace "github.com/Oppodelldog/dockertest@v0.0.3 = ../dockertest"
    
# Self-Documented Makefile see https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help