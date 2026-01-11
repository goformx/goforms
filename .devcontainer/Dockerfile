ARG VARIANT=1.25-bookworm
FROM mcr.microsoft.com/devcontainers/go:${VARIANT}

# Install system dependencies for PostgreSQL and development
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
    postgresql-client \
    curl \
    git \
    && apt-get autoremove -y && apt-get clean -y

# Switch to vscode user for Go tool installation
USER vscode

RUN go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest \
    && go install go.uber.org/mock/mockgen@latest \
    && go install github.com/air-verse/air@latest \
    && go install github.com/bombsimon/wsl/v4/cmd/wsl@latest \
    && go install golang.org/x/tools/gopls@latest \
    && go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install golang.org/x/tools/cmd/goimports@latest \
    && go install github.com/securego/gosec/v2/cmd/gosec@latest \
    && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Install Task runner
RUN sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b ~/.local/bin

# Add Go tools to PATH for the vscode user
RUN echo 'export PATH=$PATH:$(go env GOPATH)/bin:~/.local/bin' >> ~/.bashrc \
    && echo 'export PATH=$PATH:$(go env GOPATH)/bin:~/.local/bin' >> ~/.zshrc

USER root

# Optional: Install Node.js dependencies globally if needed
# RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && npm install -g @types/node typescript" 2>&1
