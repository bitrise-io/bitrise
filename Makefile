# this is intended to be used only for local testing of container related integration tests
# everything else CI related is in the bitrise.yml

DOCKER_COMPOSE_FILE=integrationtests/docker/local_docker_test_environment/docker-compose.yml
SRC_DIR_IN_GOPATH=/bitrise/src
DOCKERCOMPOSE=$(shell which docker-compose 2> /dev/null || echo '')

.PHONY: docker-with-group-test docker-step-based-test setup-test-environment build-main-container docker-clean

docker-step-based-test: setup-test-environment
	@echo "Running docker integration tests..."
	docker exec -it bitrise-main-container bash -c "export INTEGRATION_TEST_BINARY_PATH=\$$PWD/bitrise-cli; cd integrationtests/docker && go test -v -p 1 -timeout 20m --tags linux_only ./stepbased/"

docker-with-group-test: setup-test-environment
	@echo "Running docker integration tests..."
	docker exec -it bitrise-main-container bash -c "export INTEGRATION_TEST_BINARY_PATH=\$$PWD/bitrise-cli; cd integrationtests/docker && go test -v -p 1 -timeout 20m --tags linux_only ./withgroupbased/"

setup-test-environment: build-main-container
	@echo "Building bitrise binary inside container..."
	docker exec -it bitrise-main-container bash -c "go build -o bitrise-cli"

build-main-container:
	@echo "Building and starting test container..."
	@if [ "$$DOCKERCOMPOSE" ]; then \
		docker-compose -f $(DOCKER_COMPOSE_FILE) up --build -d; \
	else \
		docker compose -f $(DOCKER_COMPOSE_FILE) up --build -d; \
	fi

docker-clean:
	@echo "Stopping and removing all test containers..."
	@if [ "$$DOCKERCOMPOSE" ]; then \
		docker-compose -f $(DOCKER_COMPOSE_FILE) down; \
	else \
		docker compose -f $(DOCKER_COMPOSE_FILE) down; \
	fi
	@rm -rf _tmp /tmp/auth_* 2>/dev/null || true
	@echo "Cleanup complete!"
