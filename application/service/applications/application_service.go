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
	CreateApplication(ctx service.Context, requestBody *model.CreateAppRequest) error
}

type applicationService struct{}

func NewApplicationService() ApplicationService {
	return &applicationService{}
}

func (as *applicationService) CreateApplication(ctx service.Context, requestBody *model.CreateAppRequest) error {
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
