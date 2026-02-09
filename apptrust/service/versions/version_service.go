package versions

//go:generate ${PROJECT_DIR}/scripts/mockgen.sh ${GOFILE}

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/jfrog/jfrog-cli-application/apptrust/service"
	"github.com/jfrog/jfrog-client-go/utils/log"

	"github.com/jfrog/jfrog-cli-application/apptrust/model"
)

type VersionService interface {
	CreateAppVersion(ctx service.Context, request *model.CreateAppVersionRequest, sync bool) error
	PromoteAppVersion(ctx service.Context, applicationKey string, version string, payload *model.PromoteAppVersionRequest, sync bool) error
	ReleaseAppVersion(ctx service.Context, applicationKey string, version string, request *model.ReleaseAppVersionRequest, sync bool) error
	RollbackAppVersion(ctx service.Context, applicationKey string, version string, request *model.RollbackAppVersionRequest, sync bool) error
	DeleteAppVersion(ctx service.Context, applicationKey string, version string) error
	UpdateAppVersion(ctx service.Context, applicationKey string, version string, request *model.UpdateAppVersionRequest) error
	UpdateAppVersionSources(ctx service.Context, applicationKey string, version string, request *model.UpdateVersionSourcesRequest, sync bool, dryRun bool, failFast bool) error
}

type versionService struct{}

func NewVersionService() VersionService {
	return &versionService{}
}

func (vs *versionService) CreateAppVersion(ctx service.Context, request *model.CreateAppVersionRequest, sync bool) error {
	endpoint := fmt.Sprintf("/v1/applications/%s/versions/", request.ApplicationKey)
	response, responseBody, err := ctx.GetHttpClient().Post(endpoint, request, map[string]string{"async": strconv.FormatBool(!sync)})
	if err != nil {
		return err
	}

	expectedStatusCode := http.StatusCreated
	if !sync {
		expectedStatusCode = http.StatusAccepted
	}

	if response.StatusCode != expectedStatusCode {
		return fmt.Errorf("failed to create app version. Status code: %d. \n%s",
			response.StatusCode, responseBody)
	}

	log.Info("Application version created successfully.")
	log.Output(string(responseBody))
	return nil
}

func (vs *versionService) PromoteAppVersion(ctx service.Context, applicationKey, version string, request *model.PromoteAppVersionRequest, sync bool) error {
	endpoint := fmt.Sprintf("/v1/applications/%s/versions/%s/promote", applicationKey, version)
	response, responseBody, err := ctx.GetHttpClient().Post(endpoint, request, map[string]string{"async": strconv.FormatBool(!sync)})
	if err != nil {
		return err
	}

	if response.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("failed to promote app version. Status code: %d. \n%s",
			response.StatusCode, responseBody)
	}

	log.Output(string(responseBody))
	return nil
}

func (vs *versionService) ReleaseAppVersion(ctx service.Context, applicationKey, version string, request *model.ReleaseAppVersionRequest, sync bool) error {
	endpoint := fmt.Sprintf("/v1/applications/%s/versions/%s/release", applicationKey, version)
	response, responseBody, err := ctx.GetHttpClient().Post(endpoint, request, map[string]string{"async": strconv.FormatBool(!sync)})
	if err != nil {
		return err
	}

	if response.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("failed to release app version. Status code: %d. \n%s",
			response.StatusCode, responseBody)
	}

	log.Output(string(responseBody))
	return nil
}

func (vs *versionService) RollbackAppVersion(ctx service.Context, applicationKey, version string, request *model.RollbackAppVersionRequest, sync bool) error {
	endpoint := fmt.Sprintf("/v1/applications/%s/versions/%s/rollback", applicationKey, version)
	response, responseBody, err := ctx.GetHttpClient().Post(endpoint, request, map[string]string{"async": strconv.FormatBool(!sync)})
	if err != nil {
		return err
	}

	// Validate status code based on sync mode
	expectedStatusCode := http.StatusAccepted
	if sync {
		expectedStatusCode = http.StatusOK
	}

	if response.StatusCode != expectedStatusCode {
		return fmt.Errorf("failed to rollback app version. Status code: %d. \n%s",
			response.StatusCode, responseBody)
	}

	log.Output(string(responseBody))
	return nil
}

func (vs *versionService) DeleteAppVersion(ctx service.Context, applicationKey, version string) error {
	url := fmt.Sprintf("/v1/applications/%s/versions/%s", applicationKey, version)
	response, responseBody, err := ctx.GetHttpClient().Delete(url, map[string]string{"async": "false"})
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to delete app version. Status code: %d.\n%s",
			response.StatusCode, responseBody)
	}

	log.Info("Application version deleted successfully.")
	return nil
}

func (vs *versionService) UpdateAppVersion(ctx service.Context, applicationKey string, version string, request *model.UpdateAppVersionRequest) error {
	endpoint := fmt.Sprintf("/v1/applications/%s/versions/%s", applicationKey, version)
	response, responseBody, err := ctx.GetHttpClient().Patch(endpoint, request, nil)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update app version. Status code: %d. \n%s",
			response.StatusCode, responseBody)
	}

	log.Info("Application version updated successfully.")
	return nil
}

func (vs *versionService) UpdateAppVersionSources(ctx service.Context, applicationKey string, version string, request *model.UpdateVersionSourcesRequest, sync bool, dryRun bool, failFast bool) error {
	endpoint := fmt.Sprintf("/v1/applications/%s/versions/%s", applicationKey, version)

	params := map[string]string{
		"async":     strconv.FormatBool(!sync),
		"dry_run":   strconv.FormatBool(dryRun),
		"fail_fast": strconv.FormatBool(failFast),
	}

	response, responseBody, err := ctx.GetHttpClient().Patch(endpoint, request, params)
	if err != nil {
		return err
	}

	expectedStatusCode := http.StatusOK
	if !sync {
		expectedStatusCode = http.StatusAccepted
	}

	if response.StatusCode != expectedStatusCode {
		return fmt.Errorf("failed to update app version sources. Status code: %d. \n%s",
			response.StatusCode, responseBody)
	}

	log.Info("Application version sources updated successfully.")
	log.Output(string(responseBody))
	return nil
}
