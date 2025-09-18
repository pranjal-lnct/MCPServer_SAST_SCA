# Test MCP Server
$serverPath = ".\mcp-server.exe"

# Test initialize
$init = '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}'
$message = "Content-Length: $($init.Length)`r`n`r`n$init"

Write-Host "Testing MCP Server..."
Write-Host "Sending initialize request..."

$response = $message | & $serverPath
Write-Host "Response: $response"

# Test list_tools
$listTools = '{"jsonrpc":"2.0","id":2,"method":"list_tools"}'
$message2 = "Content-Length: $($listTools.Length)`r`n`r`n$listTools"

Write-Host "`nSending list_tools request..."
$response2 = $message2 | & $serverPath
Write-Host "Response: $response2"