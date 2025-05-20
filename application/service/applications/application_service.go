package applications

//go:generate ${PROJECT_DIR}/scripts/mockgen.sh ${GOFILE}

import (
	"fmt"
	"net/http"

	"github.com/jfrog/jfrog-client-go/utils/errorutils"

	"github.com/jfrog/jfrog-cli-application/application/model"
	"github.com/jfrog/jfrog-cli-application/application/service"
)

type ApplicationService interface {
	CreateApplication(ctx service.Context, requestBody *model.AppDescriptor) error
	UpdateApplication(ctx service.Context, requestBody *model.AppDescriptor) error
}

type applicationService struct{}

func NewApplicationService() ApplicationService {
	return &applicationService{}
}

func (as *applicationService) CreateApplication(ctx service.Context, requestBody *model.AppDescriptor) error {
	response, responseBody, err := ctx.GetHttpClient().Post("/v1/applications", requestBody)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusCreated {
		return errorutils.CheckErrorf("failed to create an application. Status code: %d.\n%s",
			response.StatusCode, responseBody)
	}

	fmt.Println(string(responseBody))
	return nil
}

func (as *applicationService) UpdateApplication(ctx service.Context, requestBody *model.AppDescriptor) error {
	endpoint := fmt.Sprintf("/v1/applications/%s", requestBody.ApplicationKey)
	response, responseBody, err := ctx.GetHttpClient().Patch(endpoint, requestBody)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return errorutils.CheckErrorf("failed to update application. Status code: %d.\n%s",
			response.StatusCode, responseBody)
	}

	fmt.Println(string(responseBody))
	return nil
}
