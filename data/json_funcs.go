package data

import (
	"encoding/json"
	"fmt"
)

func jsonString(str string) (string, error) {
	byts, err := json.Marshal(str)
	if err != nil {
		return "", fmt.Errorf("Unable to convert to JSON string: %v", err)
	}
	return string(byts), nil
}
