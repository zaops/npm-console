package cmd

import (
	"encoding/json"
	"fmt"
)

// outputJSON outputs any data structure as formatted JSON
func outputJSON(data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	fmt.Println(string(jsonData))
	return nil
}
