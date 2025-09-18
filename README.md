# SAST + SCA MCP Server

A self-contained Model Context Protocol (MCP) server that exposes two security scanning tools:

- `semgrep_scan` &mdash; static application security testing (SAST) via [Semgrep](https://github.com/semgrep/semgrep)
- `grype_scan` &mdash; dependency/software composition analysis (SCA) via [Grype](https://github.com/anchore/grype)

The server speaks JSON-RPC 2.0 over stdio so any MCP-capable agent (Amazon Q, Codex, CLI scripts, etc.) can launch it and delegate scans.

## Prerequisites (native build)

- Go 1.22+
- Semgrep CLI on `PATH`
- Grype CLI on `PATH`

```bash
# Compile the server
$ go build ./cmd/mcp-server

# Run in stdio mode (reads JSON-RPC frames from stdin)
$ ./mcp-server
```

## Local smoke tests

You can drive the server directly with framed JSON. Example PowerShell snippet (works the same with the Docker image, just swap `go run` for `docker run ...`):

```powershell
$repo    = 'C:\path\to\your\project'
$init   = '{"jsonrpc":"2.0","id":1,"method":"initialize"}'
$list   = '{"jsonrpc":"2.0","id":2,"method":"list_tools"}'
$semgrep = '{"jsonrpc":"2.0","id":3,"method":"call_tool","params":{"name":"semgrep_scan","arguments":{"target_path":"/workspace","timeout_seconds":600}}}'

$message  = "Content-Length: $($init.Length)`r`n`r`n$init"
$message += "Content-Length: $($list.Length)`r`n`r`n$list"
$message += "Content-Length: $($semgrep.Length)`r`n`r`n$semgrep"

$message | docker run --rm -i -v "$repo:/workspace" sast-sca-mcp | Tee-Object semgrep-output.json
```

Swap the final request body with `{"name":"grype_scan",...}` to exercise the SCA tool.

## CLI wrapper

Prefer a direct command-line experience? Build the lightweight wrapper and invoke the tools without crafting JSON manually:

```bash
$ go build -o mcp-cli ./cmd/mcp-cli

# Semgrep
$ ./mcp-cli semgrep --target /path/to/repo --config auto --timeout 10m

# Grype
$ ./mcp-cli grype --target /path/to/repo --timeout 5m
```

The CLI prints the raw JSON emitted by Semgrep/Grype, making it easy to pipe into `jq` or save for later triage:

```bash
$ ./mcp-cli semgrep --target /workspace | jq '.results[].extra.message'
```

`--timeout` accepts Go duration strings (`300s`, `5m`). Set it to `0` to disable the client-side deadline.
## Docker image

A Dockerfile bundles the MCP server, Semgrep, and Grype:

```bash
$ docker build -t sast-sca-mcp .
```

Run scans by mounting a repository at `/workspace` and streaming JSON-RPC frames:

```bash
# Windows command form used by Amazon Q / PowerShell
$env:CODE_DIR='C:\path\to\repo'
$payload = '{"jsonrpc":"2.0","id":1,"method":"initialize"}'
"Content-Length: $($payload.Length)`r`n`r`n$payload" | \
  docker run --rm -i -v "$env:CODE_DIR:/workspace" sast-sca-mcp
```

Inside the container the repo *must* already exist. The server validates that `target_path` resolves to a directory before executing Semgrep or Grype.

### Amazon Q integration

1. Build/pull the Docker image on the host where Amazon Q runs.
2. Register a custom MCP tool with transport `stdio` and command:
   ```
   docker run --rm -i -v "%CODE_DIR%:/workspace" sast-sca-mcp
   ```
   Set `CODE_DIR` to the project you want to scan (e.g. `C:\Users\you\repo`).
3. Increase the timeout to >=180 seconds for the first run (image pulls).
4. Trigger `initialize`/`list_tools`, then call `semgrep_scan` or `grype_scan` with `{"target_path":"/workspace"}`.

## Tool arguments

| Tool          | Required args                    | Optional args                               |
|---------------|----------------------------------|---------------------------------------------|
| `semgrep_scan`| `target_path` (string, directory)| `config` (Semgrep rule set, default `auto`), `timeout_seconds` (int, default 600) |
| `grype_scan`  | `target_path` (string, directory)| `timeout_seconds` (int, default 300)        |

Timeouts provide a soft limit; the server cancels the subprocess when exceeded and returns a JSON error payload. Errors are reported as `{"error":"message"}` objects inside the response content.

## Troubleshooting

- **`target path ... not accessible`** &mdash; ensure the directory exists inside the process/container and the MCP user has read rights.
- **Semgrep `auto` config errors** &mdash; supply an explicit `config` URI (e.g. `p/security-audit`) or point at a custom ruleset: `{"config":"/workspace/.semgrep.yml"}`.
- **Grype false positives** &mdash; export the JSON result (`--output json`) and feed it into your vulnerability triage workflow.

## License

MIT
