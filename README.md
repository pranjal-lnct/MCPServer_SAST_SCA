# SAST + SCA MCP Server

A Model Context Protocol (MCP) server that exposes two tools:

- `semgrep_scan` — runs Semgrep for SAST (source static analysis)
- `grype_scan` — runs Grype for software composition analysis

## Prerequisites

- Go 1.22+
- Semgrep (CLI) available on PATH
- Grype (CLI) available on PATH

## Build and Run

```bash
# Compile the server
$ go build ./cmd/mcp-server

# Run the server (stdio JSON-RPC)
$ ./mcp-server
```

The server speaks JSON-RPC 2.0 over stdin/stdout. Any MCP-compatible agent (e.g., Codex, Amazon Q, etc.) can launch the binary, call `initialize`, `list_tools`, and then invoke `call_tool` for `semgrep_scan` or `grype_scan`.

### Example Invocation

```bash
$ printf '{"jsonrpc":"2.0","id":1,"method":"call_tool","params":{"name":"semgrep_scan","arguments":{"target_path":"/workspace"}}}\n' \
  | ./mcp-server
```

## Docker

A Dockerfile is provided to bundle the MCP server with Semgrep and Grype so clients only need Docker installed. Build it with:

```bash
$ docker build -t sast-sca-mcp .
```

Run scans by mounting the repository inside the container and using JSON-RPC via stdin/stdout:

```bash
$ docker run --rm -i -v "$PWD":/workspace sast-sca-mcp \
    printf '{"jsonrpc":"2.0","id":1,"method":"call_tool","params":{"name":"grype_scan","arguments":{"target_path":"/workspace"}}}\n'
```

## Configuration

Both tools accept optional `timeout_seconds`. `semgrep_scan` also accepts a `config` string (defaults to `auto`). Errors/timeouts are returned in the JSON result payload as `{ "error": "message" }`.

## License

MIT
