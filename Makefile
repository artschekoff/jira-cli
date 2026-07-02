.PHONY: build install run test lint fmt vet vulncheck validate clean rulesync-install

BIN := bin/jira-cli
CMD := ./cmd/jira-cli
INSTALL_DIR := /usr/local/bin

build:
	go build -o $(BIN) $(CMD)

install: build
	sudo install -m 755 $(BIN) $(INSTALL_DIR)/jira-cli

run:
	go run $(CMD)

test:
	go test ./... -race -cover -count=1

lint:
	golangci-lint run

fmt:
	gofumpt -l -w .

vet:
	go vet ./...

vulncheck:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

validate: fmt vet lint test vulncheck

clean:
	rm -rf bin/

rulesync-env-install:
	curl -sL -H "Authorization: token $$(gh auth token)" https://raw.githubusercontent.com/artschekoff/golang-rules/refs/heads/main/scripts/install.sh | sh

rulesync-install:
	cp commands/*.md .rulesync/commands/
	rulesync generate --targets cursor --features "*"

install-here:
	rulesync fetch artschekoff/jira-cli --features rules,commands --token "$$(gh auth token)"
	rulesync generate --targets cursor --features "*"
