FROM mcr.microsoft.com/devcontainers/go:2-1.27-trixie

RUN apt-get update \
    && apt-get install -y --no-install-recommends postgresql-client \
    && rm -rf /var/lib/apt/lists/*

USER vscode

RUN go install github.com/pressly/goose/v3/cmd/goose@latest \
    && go install golang.org/x/vuln/cmd/govulncheck@latest \
    && go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
