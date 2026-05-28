package server_request

import (
	"fmt"
	"net/http"
	"strconv"

	errs "github.com/wxvn/golang-messenger/internal/errors"
)

func GetIntQueryParam(r *http.Request, key string) (*int, error) {
	param := r.URL.Query().Get(key)

	if param == "" {
		return nil, nil
	}

	val, err := strconv.Atoi(param)
	if err != nil {
		return nil, fmt.Errorf(
			"invalid int query param %s=%s: %w",
			key,
			param,
			errs.ErrInvalidArgument,
		)
	}

	return &val, nil
}

func GetStringQueryParam(r *http.Request, key string) (*string, error) {
	param := r.URL.Query().Get(key)

	if param == "" {
		return nil, nil
	}

	return &param, nil
}
