DOCKER_COMPOSE_FILE=_tests/integration/local_docker_test_environment/docker-compose.yml

docker-test: setup-test-environment
	docker exec -it bitrise-main-container bash -c "export INTEGRATION_TEST_BINARY_PATH=\$$PWD/bitrise-cli; go test ./_tests/integration -tags linux_only"

setup-test-environment: build-main-container
	docker exec -it bitrise-main-container bash -c "go build -o bitrise-cli"

build-main-container:
	docker-compose -f $(DOCKER_COMPOSE_FILE) up --build -d
