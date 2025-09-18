# Scan Electrolux project with Semgrep
$repo = "C:\Users\pranjal.sharma\source\repos\ExternalJars\sr-2101607-electrolux"

$init = '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}'
$semgrep = '{"jsonrpc":"2.0","id":2,"method":"call_tool","params":{"name":"semgrep_scan","arguments":{"target_path":"/workspace","timeout_seconds":600}}}'

$message = "Content-Length: $($init.Length)`r`n`r`n$init"
$message += "Content-Length: $($semgrep.Length)`r`n`r`n$semgrep"

Write-Host "Scanning Electrolux project with Semgrep..."
$message | docker run --rm -i -v "${repo}:/workspace" sast-sca-mcp