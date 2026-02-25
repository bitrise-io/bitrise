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

# Run all tests (takes ~10-20 minutes on Mac ARM64)
make docker-test

# Clean up everything and start fresh
make docker-clean
make docker-test

# Run a specific test only
docker exec -it bitrise-main-container bash -c \
  "export INTEGRATION_TEST_BINARY_PATH=\$PWD/bitrise-cli; \
   cd integrationtests && \
   go test -v -timeout 20m --tags linux_only ./docker -run 'Test_Docker/test_name_here'"
```