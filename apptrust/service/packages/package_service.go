package packages

//go:generate ${PROJECT_DIR}/scripts/mockgen.sh ${GOFILE}

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-application/apptrust/service"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type PackageService interface {
	BindPackage(ctx service.Context, applicationKey string, request *model.BindPackageRequest) error
	UnbindPackage(ctx service.Context, applicationKey, pkgType, pkgName, pkgVersion string) error
}

type packageService struct{}

func NewPackageService() PackageService {
	return &packageService{}
}

func (ps *packageService) BindPackage(ctx service.Context, applicationKey string, request *model.BindPackageRequest) error {
	endpoint := fmt.Sprintf("/v1/applications/%s/packages", applicationKey)
	response, responseBody, err := ctx.GetHttpClient().Post(endpoint, request, nil)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to bind package. Status code: %d.\n%s",
			response.StatusCode, responseBody)
	}

	log.Output(string(responseBody))
	return nil
}

func (ps *packageService) UnbindPackage(ctx service.Context, applicationKey, pkgType, pkgName, pkgVersion string) error {
	endpoint := fmt.Sprintf("/v1/applications/%s/packages/%s/%s/%s", applicationKey, pkgType, url.PathEscape(pkgName), pkgVersion)
	response, responseBody, err := ctx.GetHttpClient().Delete(endpoint)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to unbind package. Status code: %d.\n%s",
			response.StatusCode, responseBody)
	}

	log.Output("Package unbound successfully")
	return nil
}
