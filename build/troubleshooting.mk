
.PHONY: docker-snmp-troubleshooting/build
docker-snmp-troubleshooting/build:
	@echo "=== $(INTEGRATION) === [ docker-snmp/build ]: Building SNMP Docker image..."
	@docker build -t $(BINARY_NAME)-test -f ./build/troubleshooting/Dockerfile .

.PHONY: docker-snmp-troubleshooting
docker-snmp-troubleshooting: docker-snmp-troubleshooting/build
	@echo "=== $(INTEGRATION) === [ docker-snmp ]: Running SNMP Docker image..."
	@docker run --rm -it -p 1024:1024/udp $(BINARY_NAME)-test

.PHONY: docker-snmp-example/build
docker-snmp-example/build:
	@echo "=== $(INTEGRATION) === [ docker-snmp/build ]: Building SNMP Docker image..."
	@docker build -t $(BINARY_NAME)-example --target base -f ./build/troubleshooting/Dockerfile .

.PHONY: docker-snmp-example
docker-snmp-example: docker-snmp-example/build
	@echo "=== $(INTEGRATION) === [ docker-snmp ]: Running SNMP Docker image..."
	@docker run --rm -it -p 1024:1024/udp $(BINARY_NAME)-example
