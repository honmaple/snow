package encrypt

import (
	"errors"

	"github.com/flosch/pongo2/v7"
)

func (e *EncryptHook) encryptFilter(in *pongo2.Value, param *pongo2.Value) (out *pongo2.Value, err error) {
	plaintext, ok := in.Interface().(string)
	if !ok {
		return nil, &pongo2.Error{
			Sender:    "filter:encrypt",
			OrigError: errors.New("filter input argument must be of type 'string'"),
		}
	}

	password := ""
	if param == nil {
		password = e.opt.Password
	} else {
		password = param.String()
	}

	if password == "" {
		return nil, &pongo2.Error{
			Sender:    "filter:encrypt",
			OrigError: errors.New("password is required"),
		}
	}

	text, err := e.encrypt(plaintext, password)
	if err != nil {
		return nil, &pongo2.Error{
			Sender:    "filter:encrypt",
			OrigError: err,
		}
	}
	return pongo2.AsValue(text), nil
}
