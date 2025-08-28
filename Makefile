SHELL := /bin/sh

.PHONY: test test-backend test-frontend install-frontend

# Install frontend dependencies using the lockfile
install-frontend:
	npm --prefix web ci

# Run Go backend tests
test-backend:
	go test ./...

# Run frontend tests (Vitest)
# Ensures dependencies are installed first
test-frontend: install-frontend
	npm --prefix web run test:run

# Run all tests
test: test-backend test-frontend
