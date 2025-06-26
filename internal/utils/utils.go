package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

// GetEnv returns the environment variable value if it exists, otherwise returns the fallback value.
// The return type matches the type of the fallback parameter.
func GetEnv[T any](key string, fallback T) T {
	if value, exists := os.LookupEnv(key); exists {
		switch any(fallback).(type) {
		case string:
			return any(value).(T)
		case int:
			if intVal, err := strconv.Atoi(value); err == nil {
				return any(intVal).(T)
			}
		case bool:
			if boolVal, err := strconv.ParseBool(value); err == nil {
				return any(boolVal).(T)
			}
		case float64:
			if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
				return any(floatVal).(T)
			}
		}
	}
	return fallback
}

func FetchDataFromRequestBody[T any](request *http.Request) (T, error) {
	var data T

	body, err := io.ReadAll(request.Body)
	if err != nil {
		return data, fmt.Errorf("unable to read request body: %w", err)
	}
	if len(bytes.TrimSpace(body)) == 0 {
		return data, nil
	}
	defer request.Body.Close()

	err = json.Unmarshal(body, &data)
	if err != nil {
		return data, fmt.Errorf("unable to unmarshal request body: %w", err)
	}
	return data, nil
}
