DOCKER_IMAGE := golang:1.9
INTEGRATION_DIR := nri-$(INTEGRATION)

docker-fmt:
	@echo "=== $(INTEGRATION) === [ docker-fmt ]: Running gofmt in Docker..."
	@echo "Using Docker image $(DOCKER_IMAGE)"
	@docker run -it --rm -v $(PWD):/go/src/github.com/newrelic/$(INTEGRATION_DIR) -w /go/src/github.com/newrelic/$(INTEGRATION_DIR) $(DOCKER_IMAGE) "gofmt" "-s" "-w" "."

docker-make:
	@echo "=== $(INTEGRATION) === [ docker-fmt ]: Running make in Docker..."
	@echo "Using Docker image $(DOCKER_IMAGE)"
	@docker run -it --rm -v $(PWD):/go/src/github.com/newrelic/$(INTEGRATION_DIR) -w /go/src/github.com/newrelic/$(INTEGRATION_DIR) $(DOCKER_IMAGE) "make"

.PHONY: docker-fmt docker-make

.PHONY: docker-snmp/build
docker-snmp/build:
	@echo "=== $(INTEGRATION) === [ docker-snmp/build ]: Building SNMP Docker image..."
	@docker build -t $(INTEGRATION_DIR)-test -f ./test/Dockerfile .

.PHONY: docker-snmp
docker-snmp: docker-snmp/build
	@echo "=== $(INTEGRATION) === [ docker-snmp ]: Running SNMP Docker image..."
	@docker run --rm -it -p 1024:1024/udp $(INTEGRATION_DIR)-test
