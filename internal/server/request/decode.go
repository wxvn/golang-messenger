package server_request

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	errs "github.com/wxvn/golang-messenger/internal/errors"
)

var requestValidator = validator.New()

type validateble interface {
	Validate() error
}

func DecodeAndValidateRequest(r *http.Request, dest any) error {
	if err := json.NewDecoder(r.Body).Decode(dest); err != nil {
		return fmt.Errorf("decode json: %v: %w", err, errs.ErrInvalidArgument)
	}

	var (
		err error
	)

	v, ok := dest.(validateble)
	if ok {
		err = v.Validate()
	} else {
		err = requestValidator.Struct(dest)
	}

	if err != nil {
		return fmt.Errorf("requst validation: %v: %w", err, errs.ErrInvalidArgument)
	}

	return nil
}
