NATIVEOS     := $(shell go version | awk -F '[ /]' '{print $$4}')
NATIVEARCH   := $(shell go version | awk -F '[ /]' '{print $$5}')
INTEGRATION  := snmp
BINARY_NAME   = nri-$(INTEGRATION)
WORKDIR      := $(shell pwd)
GO_FILES     := ./src/
GOTOOLS      = github.com/kardianos/govendor \
               github.com/xeipuuv/gojsonschema

all: build

build: clean validate test compile

clean:
	@echo "=== $(INTEGRATION) === [ clean ]: Removing binaries and coverage file..."
	@rm -rfv bin coverage.xml

tools:
	@echo "=== $(INTEGRATION) === [ tools ]: Installing tools required by the project..."
	@go get $(GOTOOLS)
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.33.0

tools-update:
	@echo "=== $(INTEGRATION) === [ tools-update ]: Updating tools required by the project..."
	@go get -u $(GOTOOLS)
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.33.0

deps: tools deps-only

deps-only:
	@echo "=== $(INTEGRATION) === [ deps ]: Installing package dependencies required by the project..."
	@govendor sync

validate:
	@printf "=== $(INTEGRATION) === [ validate ]: running golangci-lint & semgrep... "
	@./bin/golangci-lint run --verbose
	@[ -f .semgrep.yml ] && semgrep_config=".semgrep.yml" || semgrep_config="p/golang" ; \
	docker run --rm -v "${PWD}:/src:ro" --workdir /src returntocorp/semgrep -c "$$semgrep_config"

compile: tools deps-only
	@echo "=== $(INTEGRATION) === [ compile ]: Building $(BINARY_NAME)..."
	@go build -o bin/$(BINARY_NAME) ./src

compile-only: deps-only
	@echo "=== $(INTEGRATION) === [ compile ]: Building $(BINARY_NAME)..."
	@go build -o bin/$(BINARY_NAME) ./src

test: deps
	@echo "=== $(INTEGRATION) === [ test ]: Running unit tests..."
	@go test -race $(GO_FILES) -count=1

integration-test: deps
	@echo "=== $(INTEGRATION) === [ test ]: running integration tests..."
	@docker-compose -f tests/integration/docker-compose.yml up -d --build
	@go test -v -tags=integration ./tests/integration/. -count=1 ; (ret=$$?; docker-compose -f tests/integration/docker-compose.yml down && exit $$ret)

# Include thematic Makefiles
include $(CURDIR)/build/ci.mk
include $(CURDIR)/build/release.mk
include $(CURDIR)/build/troubleshooting.mk

.PHONY: all build clean tools tools-update deps deps-only validate compile compile-only test integration-test
