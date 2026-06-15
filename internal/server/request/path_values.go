package server_request

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
	errs "github.com/wxvn/golang-messenger/internal/errors"
)

func GetPathValue(r *http.Request, key string) (string, error) {
	pathValue := r.PathValue(key)
	if pathValue == "" {
		return "", fmt.Errorf("no key=%s in path value: %w", key, errs.ErrInvalidArgument)
	}

	return pathValue, nil
}

func GetUUIDPathParam(r *http.Request, key string) (uuid.UUID, error) {
	val := r.PathValue(key)
	if val == "" {
		return uuid.Nil, fmt.Errorf(
			"missing path param %s: %w",
			key,
			errs.ErrInvalidArgument,
		)
	}

	uid, err := uuid.Parse(val)
	if err != nil {
		return uuid.Nil, fmt.Errorf(
			"invalid uuid path param %s=%s: %w",
			key,
			val,
			errs.ErrInvalidArgument,
		)
	}

	return uid, nil
}
