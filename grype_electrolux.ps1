# Scan Electrolux project with Grype
$repo = "C:\Users\pranjal.sharma\source\repos\ExternalJars\sr-2101607-electrolux"

$init = '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}'
$grype = '{"jsonrpc":"2.0","id":2,"method":"call_tool","params":{"name":"grype_scan","arguments":{"target_path":"/workspace","timeout_seconds":300}}}'

$message = "Content-Length: $($init.Length)`r`n`r`n$init"
$message += "Content-Length: $($grype.Length)`r`n`r`n$grype"

Write-Host "Scanning Electrolux project with Grype..."
$message | docker run --rm -i -v "${repo}:/workspace" sast-sca-mcp