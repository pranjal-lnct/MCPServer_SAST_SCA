# Test MCP Server Tools
$repo = "C:\Users\pranjal.sharma\source\repos\ExternalJars\MCPServer_SAST_SCA"

# Initialize
$init = '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}'
$list = '{"jsonrpc":"2.0","id":2,"method":"list_tools"}'
$semgrep = '{"jsonrpc":"2.0","id":3,"method":"call_tool","params":{"name":"semgrep_scan","arguments":{"target_path":"/workspace","timeout_seconds":60}}}'

$message = "Content-Length: $($init.Length)`r`n`r`n$init"
$message += "Content-Length: $($list.Length)`r`n`r`n$list"
$message += "Content-Length: $($semgrep.Length)`r`n`r`n$semgrep"

Write-Host "Testing with Docker..."
$message | docker run --rm -i -v "${repo}:/workspace" sast-sca-mcp