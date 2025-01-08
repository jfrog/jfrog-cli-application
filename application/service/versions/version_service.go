package versions

//go:generate ${PROJECT_DIR}/scripts/mockgen.sh ${GOFILE}

import (
	"fmt"

	"github.com/jfrog/jfrog-cli-application/application/service"

	"github.com/jfrog/jfrog-cli-application/application/model"
)

type VersionService interface {
	CreateAppVersion(ctx service.Context, request *model.CreateAppVersionRequest) error
	PromoteAppVersion(ctx service.Context, payload *model.PromoteAppVersionRequest) error
}

type versionService struct{}

func NewVersionService() VersionService {
	return &versionService{}
}

func (vs *versionService) CreateAppVersion(ctx service.Context, request *model.CreateAppVersionRequest) error {
	response, responseBody, err := ctx.GetHttpClient().Post("/v1/version", request)
	if err != nil {
		return err
	}

	if response.StatusCode != 201 {
		return fmt.Errorf("failed to create app version. Status code: %d. \n%s",
			response.StatusCode, responseBody)
	}

	return nil
}

func (vs *versionService) PromoteAppVersion(ctx service.Context, payload *model.PromoteAppVersionRequest) error {
	response, responseBody, err := ctx.GetHttpClient().Post("/v1/version/promote", payload)
	if err != nil {
		return err
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("failed to promote app version. Status code: %d. \n%s",
			response.StatusCode, responseBody)
	}

	return nil
}
