# this is intended to be used only for local testing of container related integration tests
# everything else CI related is in the bitrise.yml

DOCKER_COMPOSE_FILE=integrationtests/docker/local_docker_test_environment/docker-compose.yml
SRC_DIR_IN_GOPATH=/bitrise/src
DOCKERCOMPOSE=$(shell which docker-compose 2> /dev/null || echo '')

docker-test: setup-test-environment
	docker exec -it bitrise-main-container bash -c "export INTEGRATION_TEST_BINARY_PATH=\$$PWD/bitrise-cli; go test ./integrationtests/docker -tags linux_only"

setup-test-environment: build-main-container
	docker exec -it bitrise-main-container bash -c "go build -o bitrise-cli"

build-main-container:
	@if [ "$$DOCKERCOMPOSE" ]; then \
		docker-compose -f $(DOCKER_COMPOSE_FILE) up --build -d; \
	else \
		docker compose -f $(DOCKER_COMPOSE_FILE) up --build -d; \
	fi
