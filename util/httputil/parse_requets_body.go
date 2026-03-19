package httputil

import (
	"encoding/json"
	"io"
	"net/http"
)

func ParseRequestBody[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	var result T

	limitedBody := http.MaxBytesReader(w, r.Body, int64(1024*10))

	body, err := io.ReadAll(limitedBody)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, err
	}

	return result, nil
}
