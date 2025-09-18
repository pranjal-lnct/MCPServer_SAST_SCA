package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	scan "github.com/your-org/sast-sca-mcp/internal/scan"
)

const (
	jsonrpcVersion  = "2.0"
	protocolVersion = "2024-05-01"
)

type jsonrpcRequest struct {
	Version string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type jsonrpcResponse struct {
	Version string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *jsonrpcError   `json:"error,omitempty"`
}

type jsonrpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type initializeResult struct {
	ProtocolVersion string            `json:"protocolVersion"`
	ServerInfo      map[string]string `json:"serverInfo"`
	Capabilities    initializeCaps    `json:"capabilities"`
}

type initializeCaps struct {
	Tools initializeToolCaps `json:"tools"`
}

type initializeToolCaps struct {
	List bool `json:"list"`
	Call bool `json:"call"`
}

type listToolsResult struct {
	Tools []toolDescription `json:"tools"`
}

type toolDescription struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

type callToolParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type callToolResult struct {
	Content []toolOutput `json:"content"`
}

type toolOutput struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	for {
		payload, err := readFrame(reader)
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			writeProtocolError(err)
			continue
		}

		var req jsonrpcRequest
		if err := json.Unmarshal(payload, &req); err != nil {
			writeError(nil, fmt.Errorf("decode request: %w", err))
			continue
		}

		switch req.Method {
		case "initialize":
			writeResponse(req.ID, initialize())
		case "list_tools":
			writeResponse(req.ID, listTools())
		case "call_tool":
			writeResponse(req.ID, callTool(req.Params))
		default:
			writeError(req.ID, fmt.Errorf("method %q not supported", req.Method))
		}
	}
}

func initialize() initializeResult {
	return initializeResult{
		ProtocolVersion: protocolVersion,
		ServerInfo: map[string]string{
			"name":    "sast-sca-mcp",
			"version": "0.1.0",
		},
		Capabilities: initializeCaps{
			Tools: initializeToolCaps{List: true, Call: true},
		},
	}
}

func listTools() listToolsResult {
	return listToolsResult{
		Tools: []toolDescription{
			{
				Name:        "semgrep_scan",
				Description: "Runs Semgrep SAST scan on the supplied source tree.",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"target_path": map[string]any{
							"type":        "string",
							"description": "Absolute path to the directory containing source code.",
						},
						"config": map[string]any{
							"type":        "string",
							"description": "Optional Semgrep config (rule set URI). Defaults to auto.",
						},
						"timeout_seconds": map[string]any{
							"type":        "integer",
							"description": "Soft timeout for Semgrep execution.",
						},
					},
					"required": []string{"target_path"},
				},
			},
			{
				Name:        "grype_scan",
				Description: "Runs Grype SCA/vulnerability scan on dependencies detected in the target.",
				InputSchema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"target_path": map[string]any{
							"type":        "string",
							"description": "Absolute path to the directory containing the project.",
						},
						"timeout_seconds": map[string]any{
							"type":        "integer",
							"description": "Soft timeout for Grype execution.",
						},
					},
					"required": []string{"target_path"},
				},
			},
		},
	}
}

func callTool(raw json.RawMessage) callToolResult {
	var params callToolParams
	if err := json.Unmarshal(raw, &params); err != nil {
		return textError(fmt.Errorf("decode call_tool params: %w", err))
	}

	if params.Arguments == nil {
		params.Arguments = map[string]any{}
	}

	switch params.Name {
	case "semgrep_scan":
		return runSemgrep(params.Arguments)
	case "grype_scan":
		return runGrype(params.Arguments)
	default:
		return textError(fmt.Errorf("tool %q not registered", params.Name))
	}
}

func runSemgrep(args map[string]any) callToolResult {
	targetValue, err := requireString(args, "target_path")
	if err != nil {
		return textError(err)
	}
	targetPath, err := scan.ResolveDirectory(targetValue)
	if err != nil {
		return textError(err)
	}
	config := getString(args, "config", "auto")
	timeout := getDuration(args, "timeout_seconds", 10*time.Minute)

	output, err := scan.RunSemgrep(context.Background(), targetPath, config, timeout)
	if err != nil {
		return textError(err)
	}

	return callToolResult{Content: []toolOutput{{Type: "text", Text: string(output)}}}
}
func runGrype(args map[string]any) callToolResult {
	targetValue, err := requireString(args, "target_path")
	if err != nil {
		return textError(err)
	}
	targetPath, err := scan.ResolveDirectory(targetValue)
	if err != nil {
		return textError(err)
	}
	timeout := getDuration(args, "timeout_seconds", 5*time.Minute)

	output, err := scan.RunGrype(context.Background(), targetPath, timeout)
	if err != nil {
		return textError(err)
	}

	return callToolResult{Content: []toolOutput{{Type: "text", Text: string(output)}}}
}
func requireString(m map[string]any, key string) (string, error) {
	val, ok := m[key]
	if !ok {
		return "", fmt.Errorf("missing required argument %q", key)
	}
	str, ok := val.(string)
	if !ok || str == "" {
		return "", fmt.Errorf("argument %q must be a non-empty string", key)
	}
	return str, nil
}

func getString(m map[string]any, key, fallback string) string {
	val, ok := m[key]
	if !ok {
		return fallback
	}
	if str, ok := val.(string); ok && str != "" {
		return str
	}
	return fallback
}

func getDuration(m map[string]any, key string, fallback time.Duration) time.Duration {
	val, ok := m[key]
	if !ok {
		return fallback
	}

	switch v := val.(type) {
	case float64:
		if v <= 0 {
			return fallback
		}
		return time.Duration(v) * time.Second
	case int:
		if v <= 0 {
			return fallback
		}
		return time.Duration(v) * time.Second
	default:
		return fallback
	}
}

func textError(err error) callToolResult {
	return callToolResult{
		Content: []toolOutput{
			{Type: "text", Text: fmt.Sprintf("{\"error\": %q}", err.Error())},
		},
	}
}

func writeResponse(id json.RawMessage, result any) {
	resp := jsonrpcResponse{Version: jsonrpcVersion, ID: id, Result: result}
	writeJSON(resp)
}

func writeError(id json.RawMessage, err error) {
	resp := jsonrpcResponse{
		Version: jsonrpcVersion,
		ID:      id,
		Error: &jsonrpcError{
			Code:    -32000,
			Message: err.Error(),
		},
	}
	writeJSON(resp)
}

func writeProtocolError(err error) {
	writeError(nil, fmt.Errorf("protocol error: %w", err))
}

func writeJSON(v any) {
	data, err := json.Marshal(v)
	if err != nil {
		data = []byte(fmt.Sprintf("{\"jsonrpc\":\"2.0\",\"error\":{\"code\":-32603,\"message\":%q}}", err.Error()))
	}
	writeFrame(data)
}

func writeFrame(payload []byte) {
	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(payload))
	if _, err := os.Stdout.Write([]byte(header)); err != nil {
		return
	}
	if _, err := os.Stdout.Write(payload); err != nil {
		return
	}
}

func readFrame(r *bufio.Reader) ([]byte, error) {
	contentLength := -1
	seenHeader := false

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				if len(line) == 0 {
					return nil, io.EOF
				}
			} else {
				return nil, err
			}
		}

		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			if !seenHeader {
				continue
			}
			break
		}

		seenHeader = true

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])

		if key == "content-length" {
			length, err := strconv.Atoi(value)
			if err != nil {
				return nil, fmt.Errorf("invalid Content-Length %q", value)
			}
			contentLength = length
		}
	}

	if contentLength < 0 {
		return nil, fmt.Errorf("missing Content-Length header")
	}

	payload := make([]byte, contentLength)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}

	return payload, nil
}
