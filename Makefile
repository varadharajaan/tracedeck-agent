.PHONY: tools verify verify-phase1 verify-phase1b verify-phase2 verify-phase2b verify-phase3 verify-phase4 verify-phase5 check-root-clean test race vet fmt fmt-check validate-config schema smoke backend-smoke backend-newman

CONFIG ?= ./examples/policies/ai-btech-student.yaml

tools:
	powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/setup/install-go-tools.ps1

verify:
	powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase0.ps1

verify-phase1:
	powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase1.ps1

verify-phase1b:
	powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase1b.ps1

verify-phase2:
	powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase2.ps1

verify-phase2b:
	powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase2b.ps1

verify-phase3:
	powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase3.ps1

verify-phase4:
	powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase4.ps1

verify-phase5:
	powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase5.ps1

check-root-clean:
	powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-root-clean.ps1

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

fmt:
	powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/format-go.ps1

fmt-check:
	powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-gofmt.ps1

validate-config:
	go run ./agent/cmd/tracedeck-agent validate-config --config $(CONFIG)

schema:
	go run ./agent/cmd/tracedeck-agent schema

smoke:
	powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase0.ps1

backend-smoke:
	powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase5.ps1

backend-newman:
	powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase5.ps1
