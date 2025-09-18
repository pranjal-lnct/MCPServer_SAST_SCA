# syntax=docker/dockerfile:1.6

ARG SEMGREP_VERSION=1.70.0
ARG GRYPE_VERSION=0.77.0

FROM golang:1.22-bookworm AS builder
WORKDIR /src

# Prime the module cache before bringing in the full tree.
COPY go.mod go.sum ./
RUN go mod download

# Build the MCP server binary.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/mcp-server ./cmd/mcp-server

FROM debian:bookworm-slim

ARG SEMGREP_VERSION
ARG GRYPE_VERSION

# Install runtime dependencies and tooling prerequisites.
RUN apt-get update && \
    apt-get install -y --no-install-recommends curl gnupg python3 python3-pip ca-certificates && \
    rm -rf /var/lib/apt/lists/*

ENV PIP_BREAK_SYSTEM_PACKAGES=1

# Install Semgrep via pip.
RUN pip3 install --no-cache-dir "semgrep==${SEMGREP_VERSION}"

# Install Grype from the official release tarball.
RUN curl -fsSL "https://github.com/anchore/grype/releases/download/v${GRYPE_VERSION}/grype_${GRYPE_VERSION}_linux_amd64.tar.gz" -o /tmp/grype.tgz && \
    tar -xzf /tmp/grype.tgz -C /usr/local/bin grype && \
    rm /tmp/grype.tgz

# Add the compiled MCP server.
COPY --from=builder /out/mcp-server /usr/local/bin/mcp-server
RUN chmod +x /usr/local/bin/mcp-server

# Create an unprivileged user and workspace.
RUN useradd --uid 1000 --create-home scanner && \
    mkdir -p /workspace && \
    chown scanner:scanner /workspace
USER scanner
WORKDIR /workspace

ENV PATH="/usr/local/bin:${PATH}"

ENTRYPOINT ["/usr/local/bin/mcp-server"]
