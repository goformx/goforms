ARG VARIANT=1.24-bookworm
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
    && go install github.com/a-h/templ/cmd/templ@latest \
    && go install go.uber.org/mock/mockgen@latest \
    && go install github.com/air-verse/air@latest \
    && sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b ~/.local/bin

# Install dependencies and build the template package
# RUN cd /workspaces/goforms-template && npm install && npm run build && npm link || true

USER root

# Optional: Install Node.js dependencies globally if needed
# RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && npm install -g @types/node typescript" 2>&1
