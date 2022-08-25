GOLANGCI_IMAGE=golangci/golangci-lint:latest-alpine

lint:
	docker run --rm -v ${PWD}:/app -w /app ${GOLANGCI_IMAGE} golangci-lint run --fix --timeout 20m --sort-results