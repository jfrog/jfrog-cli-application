package service

import (
	"fmt"
	"github.com/jfrog/jfrog-cli-application/application/http"
	"github.com/jfrog/jfrog-cli-application/application/model"
)

type VersionService interface {
	CreateAppVersion(ctx *Context, request *model.CreateAppVersionRequest) error
}

type versionService struct {
}

func NewVersionService() VersionService {
	return &versionService{}
}

func (vs *versionService) CreateAppVersion(ctx *Context, request *model.CreateAppVersionRequest) error {
	httpClient, err := http.NewAppHttpClient(ctx.ServerDetails)
	if err != nil {
		return err
	}

	response, responseBody, err := httpClient.Post("/v1/version", request)
	if err != nil {
		return err
	}

	if response.StatusCode != 201 {
		return fmt.Errorf("failed to create app version. Status code: %d. \n%s",
			response.StatusCode, responseBody)
	}

	return nil
}
