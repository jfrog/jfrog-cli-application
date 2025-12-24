# E2E Tests

## Running E2E Tests

### Prerequisites

1. Setup a JPD (JFrog Platform Deployment)
2. Create an identity token with admin permissions

### Configuration

You can configure the tests using either environment variables or command-line flags.

#### Option 1: Environment Variables

Set the following environment variables:

* **JFROG_APPTRUST_CLI_TESTS_JFROG_URL**: The JFrog Platform URL (defaults to `http://localhost:8082/` if not set)
* **JFROG_APPTRUST_CLI_TESTS_JFROG_ACCESS_TOKEN**: The JFrog Platform access token

Example:

```bash
export JFROG_APPTRUST_CLI_TESTS_JFROG_URL=http://localhost:8082
export JFROG_APPTRUST_CLI_TESTS_JFROG_ACCESS_TOKEN=your-access-token
go test -tags=e2e ./e2e/...
```

#### Option 2: Command-Line Flags

Use the `-jfrog.url` and `-jfrog.adminToken` flags:

Example:

```bash
go test -tags=e2e -jfrog.url=http://localhost:8082 -jfrog.adminToken=your-access-token ./e2e/...
```
