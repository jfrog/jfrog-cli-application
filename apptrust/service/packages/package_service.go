package packages

//go:generate ${PROJECT_DIR}/scripts/mockgen.sh ${GOFILE}

import (
	"fmt"
	"net/http"

	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-application/apptrust/service"
)

type PackageService interface {
	BindPackage(ctx service.Context, request *model.BindPackageRequest) error
}

type packageService struct{}

func NewPackageService() PackageService {
	return &packageService{}
}

func (ps *packageService) BindPackage(ctx service.Context, request *model.BindPackageRequest) error {
	endpoint := "/v1/package"
	response, responseBody, err := ctx.GetHttpClient().Post(endpoint, request)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to bind package. Status code: %d. \n%s",
			response.StatusCode, responseBody)
	}

	return nil
}
