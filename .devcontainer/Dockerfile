# Base image for Go 1.24
FROM debian:bookworm

# Install prerequisites and Go 1.24
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get install -y --no-install-recommends \
        wget \
        curl \
        tar \
        ca-certificates \
        git \
        make \
        gcc \
        g++ \
        pkg-config \
        libssl-dev \
    && wget https://go.dev/dl/go1.24.2.linux-amd64.tar.gz \
    && rm -rf /usr/local/go \
    && tar -C /usr/local -xzf go1.24.2.linux-amd64.tar.gz \
    && rm go1.24.2.linux-amd64.tar.gz \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Set Go environment variables
ENV PATH="/usr/local/go/bin:${PATH}" \
    GOPATH="/go" \
    GOCACHE="/go/cache" \
    GO111MODULE="on" \
    CGO_ENABLED=1

# Install MariaDB client tools, golang-migrate, and go-task
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
        mariadb-client \
        default-mysql-client \
    && curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz \
    && mv migrate /usr/local/bin/migrate \
    && chmod +x /usr/local/bin/migrate \
    && sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install Docker Compose
RUN curl -L "https://github.com/docker/compose/releases/download/2.32.4/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose \
    && chmod +x /usr/local/bin/docker-compose

# Install act for local GitHub Actions testing
RUN curl -s https://raw.githubusercontent.com/nektos/act/master/install.sh | bash

# Create vscode user and set up Go directories
RUN groupadd -r vscode && useradd -r -g vscode -m -s /bin/bash vscode \
    && mkdir -p /go/{bin,pkg,cache,src} \
    && chown -R vscode:vscode /go

# Switch to vscode user for subsequent operations
USER vscode

# Install Go tools
RUN go install github.com/go-delve/delve/cmd/dlv@latest && \
    go install golang.org/x/tools/gopls@latest && \
    go install github.com/fatih/gomodifytags@latest && \
    go install github.com/cweill/gotests/gotests@latest && \
    go install github.com/a-h/templ/cmd/templ@latest && \
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest && \
    go install github.com/vektra/mockery/v2@latest && \
    go install github.com/swaggo/swag/cmd/swag@latest && \
    go install github.com/securego/gosec/v2/cmd/gosec@latest && \
    go install github.com/kyoh86/richgo@latest && \
    go install github.com/rakyll/gotest@latest && \
    go install github.com/segmentio/golines@latest && \
    go install github.com/daixiang0/gci@latest

# Set up shell environment
RUN echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bashrc && \
    echo 'export GOCACHE=/go/cache' >> ~/.bashrc && \
    echo 'export GO111MODULE=on' >> ~/.bashrc && \
    echo 'export CGO_ENABLED=1' >> ~/.bashrc

# Create workspace directory
RUN mkdir -p /workspaces/goforms && \
    chown -R vscode:vscode /workspaces/goforms

WORKDIR /workspaces/goforms
