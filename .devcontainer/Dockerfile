ARG VARIANT=1-1-bookworm
FROM mcr.microsoft.com/devcontainers/go:${VARIANT}

# [Optional] Uncomment this section to install additional OS packages.
# RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
#     && apt-get -y install --no-install-recommends <your-package-list-here>

# [Optional] Uncomment the next lines to use go get to install anything else you need
USER vscode
RUN go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.2 && \
    go install github.com/golang-migrate/migrate/v4/cmd/migrate@v4.15.2 && \
    go install github.com/a-h/templ/cmd/templ@v0.2.1
USER root

# [Optional] Uncomment this line to install global node packages.
# RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && npm install -g <your-package-here>" 2>&1