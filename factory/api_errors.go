package factory

import "github.com/hublabs/common/api"

func NewError(errTemplate api.ErrorTemplate, message string) api.Error {
	err := errTemplate.New(nil)
	err.Message = message
	return err
}
