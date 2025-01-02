package auth

import "github.com/jfrog/jfrog-client-go/auth"

type appDetails struct {
	auth.CommonConfigFields
}

func NewAppDetails() auth.ServiceDetails {
	return &appDetails{}
}

func (rt *appDetails) GetVersion() (string, error) {
	panic("Failed: Method is not implemented")
}
