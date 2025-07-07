package versions

//go:generate ${PROJECT_DIR}/scripts/mockgen.sh ${GOFILE}

import (
	"fmt"
	"strconv"

	"github.com/jfrog/jfrog-cli-application/apptrust/service"

	"github.com/jfrog/jfrog-cli-application/apptrust/model"
)

type VersionService interface {
	CreateAppVersion(ctx service.Context, request *model.CreateAppVersionRequest) error
	PromoteAppVersion(ctx service.Context, applicationKey string, version string, payload *model.PromoteAppVersionRequest, sync bool) error
	ReleaseAppVersion(ctx service.Context, applicationKey string, version string, request *model.ReleaseAppVersionRequest, sync bool) error
	DeleteAppVersion(ctx service.Context, applicationKey string, version string) error
	UpdateAppVersion(ctx service.Context, applicationKey string, version string, request *model.UpdateAppVersionRequest) error
}

type versionService struct{}

func NewVersionService() VersionService {
	return &versionService{}
}

func (vs *versionService) CreateAppVersion(ctx service.Context, request *model.CreateAppVersionRequest) error {
	endpoint := fmt.Sprintf("/v1/applications/%s/versions/", request.ApplicationKey)
	response, responseBody, err := ctx.GetHttpClient().Post(endpoint, request, nil)
	if err != nil {
		return err
	}

	if response.StatusCode != 201 {
		return fmt.Errorf("failed to create app version. Status code: %d. \n%s",
			response.StatusCode, responseBody)
	}

	return nil
}

func (vs *versionService) PromoteAppVersion(ctx service.Context, applicationKey, version string, request *model.PromoteAppVersionRequest, sync bool) error {
	endpoint := fmt.Sprintf("/v1/applications/%s/versions/%s/promote", applicationKey, version)
	response, responseBody, err := ctx.GetHttpClient().Post(endpoint, request, map[string]string{"async": strconv.FormatBool(!sync)})
	if err != nil {
		return err
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("failed to promote app version. Status code: %d. \n%s",
			response.StatusCode, responseBody)
	}

	return nil
}

func (vs *versionService) ReleaseAppVersion(ctx service.Context, applicationKey, version string, request *model.ReleaseAppVersionRequest, sync bool) error {
	endpoint := fmt.Sprintf("/v1/applications/%s/versions/%s/release", applicationKey, version)
	response, responseBody, err := ctx.GetHttpClient().Post(endpoint, request, map[string]string{"async": strconv.FormatBool(!sync)})
	if err != nil {
		return err
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("failed to release app version. Status code: %d. \n%s",
			response.StatusCode, responseBody)
	}

	return nil
}

func (vs *versionService) DeleteAppVersion(ctx service.Context, applicationKey, version string) error {
	url := fmt.Sprintf("/v1/applications/%s/versions/%s", applicationKey, version)
	response, responseBody, err := ctx.GetHttpClient().Delete(url)
	if err != nil {
		return err
	}

	if response.StatusCode != 204 {
		return fmt.Errorf("failed to delete app version. Status code: %d.\n%s",
			response.StatusCode, responseBody)
	}

	return nil
}

func (vs *versionService) UpdateAppVersion(ctx service.Context, applicationKey, version string, request *model.UpdateAppVersionRequest) error {
	endpoint := fmt.Sprintf("/v1/applications/%s/versions/%s", applicationKey, version)
	response, responseBody, err := ctx.GetHttpClient().Patch(endpoint, request)
	if err != nil {
		return err
	}

	if response.StatusCode != 202 {
		return fmt.Errorf("failed to update app version. Status code: %d. \n%s",
			response.StatusCode, responseBody)
	}

	return nil
}
