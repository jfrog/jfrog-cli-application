package applications

//go:generate ${PROJECT_DIR}/scripts/mockgen.sh ${GOFILE}

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/log"

	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-application/apptrust/service"
)

type ApplicationService interface {
	CreateApplication(ctx service.Context, requestBody *model.AppDescriptor) ([]byte, error)
	UpdateApplication(ctx service.Context, requestBody *model.AppDescriptor) ([]byte, error)
	DeleteApplication(ctx service.Context, applicationKey string) error
	ExportApplication(ctx service.Context, applicationKey string) ([]byte, error)
	ImportApplication(ctx service.Context, applicationEnvelope []byte) error
}

type applicationService struct{}

func NewApplicationService() ApplicationService {
	return &applicationService{}
}

func (as *applicationService) CreateApplication(ctx service.Context, requestBody *model.AppDescriptor) ([]byte, error) {
	response, responseBody, err := ctx.GetHttpClient().Post("/v1/applications", requestBody, nil)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusCreated {
		return nil, errorutils.CheckErrorf("failed to create an application. Status code: %d.\n%s",
			response.StatusCode, responseBody)
	}

	log.Info(fmt.Sprintf("Application \"%s\" created successfully.", requestBody.ApplicationKey))
	return responseBody, nil
}

func (as *applicationService) UpdateApplication(ctx service.Context, requestBody *model.AppDescriptor) ([]byte, error) {
	endpoint := fmt.Sprintf("/v1/applications/%s", requestBody.ApplicationKey)
	response, responseBody, err := ctx.GetHttpClient().Patch(endpoint, requestBody, nil)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, errorutils.CheckErrorf("failed to update application. Status code: %d.\n%s",
			response.StatusCode, responseBody)
	}

	log.Info(fmt.Sprintf("Application \"%s\" updated successfully.", requestBody.ApplicationKey))
	return responseBody, nil
}

func (as *applicationService) DeleteApplication(ctx service.Context, applicationKey string) error {
	endpoint := fmt.Sprintf("/v1/applications/%s", applicationKey)
	response, responseBody, err := ctx.GetHttpClient().Delete(endpoint, nil)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusNoContent {
		return errorutils.CheckErrorf("failed to delete application. Status code: %d.\n%s",
			response.StatusCode, responseBody)
	}

	log.Info(fmt.Sprintf("Application \"%s\" deleted successfully.", applicationKey))
	return nil
}

func (as *applicationService) ExportApplication(ctx service.Context, applicationKey string) ([]byte, error) {
	endpoint := fmt.Sprintf("/v1/applications/%s/export", applicationKey)
	response, responseBody, err := ctx.GetHttpClient().Post(endpoint, map[string]any{}, nil)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, errorutils.CheckErrorf("failed to export application. Status code: %d.\n%s",
			response.StatusCode, responseBody)
	}

	log.Info(fmt.Sprintf("Application \"%s\" exported successfully.", applicationKey))
	return responseBody, nil
}

func (as *applicationService) ImportApplication(ctx service.Context, applicationEnvelope []byte) error {
	response, responseBody, err := ctx.GetHttpClient().Post("/v1/applications/import", json.RawMessage(applicationEnvelope), nil)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return errorutils.CheckErrorf("failed to import application. Status code: %d.\n%s",
			response.StatusCode, responseBody)
	}

	log.Info("Application imported successfully.")
	return nil
}
