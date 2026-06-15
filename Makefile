PYTHON ?= python3

.PHONY: validate-registry test build vet check

validate-registry:
	@$(PYTHON) -m pip install --quiet jsonschema
	@$(PYTHON) scripts/validate_registry.py

test:
	go test ./...

build:
	go build ./...

vet:
	go vet ./...

check: validate-registry test build vet
