package versions

//go:generate ${PROJECT_DIR}/scripts/mockgen.sh ${GOFILE}

import (
	"fmt"

	"github.com/jfrog/jfrog-cli-application/apptrust/service"

	"github.com/jfrog/jfrog-cli-application/apptrust/model"
)

type VersionService interface {
	CreateAppVersion(ctx service.Context, request *model.CreateAppVersionRequest) error
	PromoteAppVersion(ctx service.Context, payload *model.PromoteAppVersionRequest) error
	DeleteAppVersion(ctx service.Context, applicationKey string, version string) error
}

type versionService struct{}

func NewVersionService() VersionService {
	return &versionService{}
}

func (vs *versionService) CreateAppVersion(ctx service.Context, request *model.CreateAppVersionRequest) error {
	response, responseBody, err := ctx.GetHttpClient().Post("/v1/applications/version", request)
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
	response, responseBody, err := ctx.GetHttpClient().Post("/v1/applications/version/promote", payload)
	if err != nil {
		return err
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("failed to promote app version. Status code: %d. \n%s",
			response.StatusCode, responseBody)
	}

	return nil
}

func (vs *versionService) DeleteAppVersion(ctx service.Context, applicationKey string, version string) error {
	url := fmt.Sprintf("/v1/applications/%s/versions/%s", applicationKey, version)
	response, responseBody, err := ctx.GetHttpClient().Delete(url, nil)
	if err != nil {
		return err
	}

	if response.StatusCode != 204 {
		return fmt.Errorf("failed to delete app version. Status code: %d.\n%s",
			response.StatusCode, responseBody)
	}

	return nil
}
