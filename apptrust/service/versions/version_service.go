package versions

//go:generate ${PROJECT_DIR}/scripts/mockgen.sh ${GOFILE}

import (
	"fmt"
	"net/http"
	"strconv"

	apphttp "github.com/jfrog/jfrog-cli-application/apptrust/http"
	"github.com/jfrog/jfrog-cli-application/apptrust/service"
	"github.com/jfrog/jfrog-client-go/utils/log"

	"github.com/jfrog/jfrog-cli-application/apptrust/model"
)

type VersionService interface {
	CreateAppVersion(ctx service.Context, request *model.CreateAppVersionRequest, sync, dryRun bool, conflictResolution string) ([]byte, error)
	PromoteAppVersion(ctx service.Context, applicationKey string, version string, payload *model.PromoteAppVersionRequest, sync bool) ([]byte, error)
	ReleaseAppVersion(ctx service.Context, applicationKey string, version string, request *model.ReleaseAppVersionRequest, sync bool) ([]byte, error)
	RollbackAppVersion(ctx service.Context, applicationKey string, version string, request *model.RollbackAppVersionRequest, sync bool) ([]byte, error)
	DeleteAppVersion(ctx service.Context, applicationKey string, version string) error
	UpdateAppVersion(ctx service.Context, applicationKey string, version string, request *model.UpdateAppVersionRequest) ([]byte, error)
	UpdateAppVersionSources(ctx service.Context, applicationKey string, version string, request *model.UpdateVersionSourcesRequest, sync bool, dryRun bool, failFast bool) ([]byte, error)
	DistributeAppVersion(ctx service.Context, applicationKey string, version string, request *model.DistributeAppVersionRequest) error
	RemoteDeleteAppVersion(ctx service.Context, applicationKey string, version string, request *model.RemoteDeleteAppVersionRequest) error
}

type versionService struct{}

func NewVersionService() VersionService {
	return &versionService{}
}

func (vs *versionService) CreateAppVersion(ctx service.Context, request *model.CreateAppVersionRequest, sync, dryRun bool, conflictResolution string) ([]byte, error) {
	endpoint := fmt.Sprintf("/v1/applications/%s/versions/", request.ApplicationKey)
	params := map[string]string{
		"async":   strconv.FormatBool(!sync),
		"dry_run": strconv.FormatBool(dryRun),
	}
	if conflictResolution != "" {
		params["conflict_resolution"] = conflictResolution
	}
	response, responseBody, err := ctx.GetHttpClient().Post(endpoint, request, params)
	if err != nil {
		return nil, err
	}

	if !apphttp.IsSuccessStatusCode(response.StatusCode) {
		return nil, fmt.Errorf("failed to create app version. Status code: %d. \n%s",
			response.StatusCode, responseBody)
	}

	logSuccessMessage(sync, request, dryRun)
	return responseBody, nil
}

func (vs *versionService) PromoteAppVersion(ctx service.Context, applicationKey, version string, request *model.PromoteAppVersionRequest, sync bool) ([]byte, error) {
	endpoint := fmt.Sprintf("/v1/applications/%s/versions/%s/promote", applicationKey, version)
	response, responseBody, err := ctx.GetHttpClient().Post(endpoint, request, map[string]string{"async": strconv.FormatBool(!sync)})
	if err != nil {
		return nil, err
	}

	if !apphttp.IsSuccessStatusCode(response.StatusCode) {
		return nil, fmt.Errorf("failed to promote app version. Status code: %d. \n%s",
			response.StatusCode, responseBody)
	}

	return responseBody, nil
}

func (vs *versionService) ReleaseAppVersion(ctx service.Context, applicationKey, version string, request *model.ReleaseAppVersionRequest, sync bool) ([]byte, error) {
	endpoint := fmt.Sprintf("/v1/applications/%s/versions/%s/release", applicationKey, version)
	response, responseBody, err := ctx.GetHttpClient().Post(endpoint, request, map[string]string{"async": strconv.FormatBool(!sync)})
	if err != nil {
		return nil, err
	}

	if !apphttp.IsSuccessStatusCode(response.StatusCode) {
		return nil, fmt.Errorf("failed to release app version. Status code: %d. \n%s",
			response.StatusCode, responseBody)
	}

	return responseBody, nil
}

func (vs *versionService) RollbackAppVersion(ctx service.Context, applicationKey, version string, request *model.RollbackAppVersionRequest, sync bool) ([]byte, error) {
	endpoint := fmt.Sprintf("/v1/applications/%s/versions/%s/rollback", applicationKey, version)
	response, responseBody, err := ctx.GetHttpClient().Post(endpoint, request, map[string]string{"async": strconv.FormatBool(!sync)})
	if err != nil {
		return nil, err
	}

	if !apphttp.IsSuccessStatusCode(response.StatusCode) {
		return nil, fmt.Errorf("failed to rollback app version. Status code: %d. \n%s",
			response.StatusCode, responseBody)
	}

	return responseBody, nil
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

func (vs *versionService) UpdateAppVersion(ctx service.Context, applicationKey string, version string, request *model.UpdateAppVersionRequest) ([]byte, error) {
	endpoint := fmt.Sprintf("/v1/applications/%s/versions/%s", applicationKey, version)
	response, responseBody, err := ctx.GetHttpClient().Patch(endpoint, request, nil)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to update app version. Status code: %d. \n%s",
			response.StatusCode, responseBody)
	}

	log.Info("Application version updated successfully.")
	return responseBody, nil
}

func (vs *versionService) UpdateAppVersionSources(ctx service.Context, applicationKey string, version string, request *model.UpdateVersionSourcesRequest, sync bool, dryRun bool, failFast bool) ([]byte, error) {
	endpoint := fmt.Sprintf("/v1/applications/%s/versions/%s", applicationKey, version)

	params := map[string]string{
		"async":     strconv.FormatBool(!sync),
		"dry_run":   strconv.FormatBool(dryRun),
		"fail_fast": strconv.FormatBool(failFast),
	}

	response, responseBody, err := ctx.GetHttpClient().Patch(endpoint, request, params)
	if err != nil {
		return nil, err
	}

	expectedStatusCode := http.StatusOK
	if !sync {
		expectedStatusCode = http.StatusAccepted
	}

	if response.StatusCode != expectedStatusCode {
		return nil, fmt.Errorf("failed to update app version sources. Status code: %d. \n%s",
			response.StatusCode, responseBody)
	}

	log.Info("Application version sources updated successfully.")
	return responseBody, nil
}

func (vs *versionService) DistributeAppVersion(ctx service.Context, applicationKey, version string, request *model.DistributeAppVersionRequest) error {
	endpoint := fmt.Sprintf("/v1/applications/%s/versions/%s/distribute", applicationKey, version)
	response, responseBody, err := ctx.GetHttpClient().Post(endpoint, request, nil)
	if err != nil {
		return err
	}

	if !apphttp.IsSuccessStatusCode(response.StatusCode) {
		return fmt.Errorf("failed to distribute application version. Status code: %d. \n%s",
			response.StatusCode, responseBody)
	}

	log.Info(fmt.Sprintf("Distribution of application version '%s/%s' triggered successfully.", applicationKey, version))
	return nil
}

func (vs *versionService) RemoteDeleteAppVersion(ctx service.Context, applicationKey, version string, request *model.RemoteDeleteAppVersionRequest) error {
	endpoint := fmt.Sprintf("/v1/applications/%s/versions/%s/remote-delete", applicationKey, version)
	response, responseBody, err := ctx.GetHttpClient().Post(endpoint, request, nil)
	if err != nil {
		return err
	}

	if !apphttp.IsSuccessStatusCode(response.StatusCode) {
		return fmt.Errorf("failed to delete application version remotely. Status code: %d. \n%s",
			response.StatusCode, responseBody)
	}

	log.Info(fmt.Sprintf("Remote deletion of application version '%s/%s' triggered successfully.", applicationKey, version))
	return nil
}

func logSuccessMessage(sync bool, request *model.CreateAppVersionRequest, dryRun bool) {
	if !sync {
		log.Info(fmt.Sprintf("Application version creation initiated: %s:%s", request.ApplicationKey, request.Version))
	} else if dryRun {
		log.Info(fmt.Sprintf("Dry run successful for application version: %s:%s", request.ApplicationKey, request.Version))
	} else {
		log.Info(fmt.Sprintf("Application version created successfully: %s:%s", request.ApplicationKey, request.Version))
	}
}
