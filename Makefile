.PHONY: tools verify verify-phase1 verify-phase1b check-root-clean test race vet fmt fmt-check validate-config schema smoke

CONFIG ?= ./examples/policies/ai-btech-student.yaml

tools:
	powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/setup/install-go-tools.ps1

verify:
	powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase0.ps1

verify-phase1:
	powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase1.ps1

verify-phase1b:
	powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase1b.ps1

check-root-clean:
	powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-root-clean.ps1

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -w ./agent

fmt-check:
	powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-gofmt.ps1

validate-config:
	go run ./agent/cmd/tracedeck-agent validate-config --config $(CONFIG)

schema:
	go run ./agent/cmd/tracedeck-agent schema

smoke:
	powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase0.ps1
