package server_request

import (
	"fmt"
	"net/http"

	errs "github.com/wxvn/golang-messenger/internal/errors"
)

func GetPathValue(r *http.Request, key string) (string, error) {
	pathValue := r.PathValue(key)
	if pathValue == "" {
		return "", fmt.Errorf("no key=%s in path value: %w", key, errs.ErrInvalidArgument)
	}

	return pathValue, nil
}
