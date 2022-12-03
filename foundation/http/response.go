package http

import "encoding/json"

func JsonStatus(message string) []byte {
	b, _ := json.Marshal(struct {
		Message string `json:"message"`
	}{
		Message: message,
	})
	return b
}
