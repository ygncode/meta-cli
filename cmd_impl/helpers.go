package cmd_impl

import (
	"encoding/json"
	"fmt"
	"os"
)

func readJSONInput(jsonStr, filePath string) (json.RawMessage, error) {
	if jsonStr != "" && filePath != "" {
		return nil, fmt.Errorf("specify --json or --file, not both")
	}
	if jsonStr != "" {
		if !json.Valid([]byte(jsonStr)) {
			return nil, fmt.Errorf("invalid JSON string")
		}
		return json.RawMessage(jsonStr), nil
	}
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("read file: %w", err)
		}
		if !json.Valid(data) {
			return nil, fmt.Errorf("file does not contain valid JSON")
		}
		return json.RawMessage(data), nil
	}
	return nil, fmt.Errorf("--json or --file is required")
}
