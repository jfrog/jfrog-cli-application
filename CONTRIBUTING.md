# Contribution Guide

Welcome to the contribution guide for our project! We appreciate your interest in contributing to the development of this project. Below, you will find essential information on local development, running tests, and guidelines for submitting pull requests.

## Table of Contents

- [🏠🏗️ Local development](#%EF%B8%8F-local-development)
- [🚦 Running Tests](#-running-tests)
- [📖 Submitting PR Guidelines](#-submitting-pr-guidelines)


## 🏠🏗️ Local Development

To run a command locally, use the following command template:

```sh
go run github.com/jfrog/jfrog-cli-application command [options] [arguments...]
```

---

## 🚦 Running Tests

To run unit tests, use the following command:

```
make test
```

To run end-to-end (E2E) tests, refer to the README file located in the `e2e` directory.

---

## 📖 Submitting PR Guidelines

### Before submitting the pull request, ensure:

- Your changes are covered by `unit` and `e2e` tests. If not, please add new tests.
- The code has been validated to compile successfully by running `go vet ./...`
- The code has been formatted properly using `go fmt ./...`

### When creating the pull request, ensure:

- The pull request is targeting the `main` branch.
- The pull request description describes the changes made.
