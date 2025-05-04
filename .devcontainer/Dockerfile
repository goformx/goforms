ARG VARIANT=1-1-bookworm
FROM mcr.microsoft.com/devcontainers/go:${VARIANT}

# [Optional] Uncomment this section to install additional OS packages.
# RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
#     && apt-get -y install --no-install-recommends <your-package-list-here>

# [Optional] Uncomment the next lines to use go get to install anything else you need
USER vscode

RUN go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest && \
    go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest && \
    go install github.com/a-h/templ/cmd/templ@latest && \
    go install github.com/air-verse/air@latest

# Install dependencies and build the template package
RUN cd /workspaces/goforms-template && npm install && npm run build && npm link || true

USER root

# [Optional] Uncomment this line to install global node packages.
# RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && npm install -g <your-package-here>" 2>&1ce