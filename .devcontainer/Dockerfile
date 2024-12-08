FROM mcr.microsoft.com/devcontainers/go:1-1.23-bookworm

# Install PostgreSQL client tools and golang-migrate
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends postgresql-client \
    && curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz \
    && mv migrate /usr/local/bin/migrate \
    && chmod +x /usr/local/bin/migrate \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# [Optional] Uncomment this section to install additional OS packages.
# RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
#     && apt-get -y install --no-install-recommends <your-package-list-here>

# [Optional] Uncomment the next lines to use go get to install anything else you need
# USER vscode
# RUN go get -x <your-dependency-or-tool>
# USER root

# [Optional] Uncomment this line to install global node packages.
# RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && npm install -g <your-package-here>" 2>&1
