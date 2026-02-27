# Integration Tests

Integration tests are treated as separate from the main Go package.
This is intentional in order for these to be executed on-demand and decoupled from the main test suite. Running these integration tests can take a long time.

> [!WARNING]
> Because of this whenever dependencies change you'll have to run `go mod tidy` in this directory too!
> 
> Keep in mind that the integration test do not use vendoring. You SHOULD NOT run `go mod vendor` and MUST NOT push vendor folder here.

## Docker Tests

### Prerequisites

- Docker Desktop (Mac) or Docker Engine (Linux)

### Running Tests

```bash
# From the repository root

# Run all Step based containerization tests (takes ~10-20 minutes on Mac ARM64)
make docker-step-based-test

# Run all With group based containerization tests (takes ~10-20 minutes on Mac ARM64)
make docker-with-group-test

# Clean up everything and start fresh
make docker-clean
make docker-test

# Run a specific Step based containerization test only
docker exec -it bitrise-main-container bash -c \
  "export INTEGRATION_TEST_BINARY_PATH=\$PWD/bitrise-cli; \
   cd integrationtests/docker && \
   go test -v -timeout 20m -count=1 --tags linux_only ./stepbased/ -run 'Test_Docker/test_name_here'"
   
# Run a specific With group based containerization test only
docker exec -it bitrise-main-container bash -c \
  "export INTEGRATION_TEST_BINARY_PATH=\$PWD/bitrise-cli; \
   cd integrationtests/docker && \
   go test -v -timeout 20m -count=1 --tags linux_only ./withgroupbased/ -run 'Test_Docker/test_name_here'"
```