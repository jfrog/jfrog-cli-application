package systems

//go:generate ${PROJECT_DIR}/scripts/mockgen.sh ${GOFILE}

import (
	"fmt"

	"github.com/jfrog/jfrog-cli-application/apptrust/service"
	"github.com/jfrog/jfrog-client-go/utils/log"
)

type SystemService interface {
	Ping(ctx service.Context) error
}

type systemService struct{}

func NewSystemService() SystemService {
	return &systemService{}
}

func (ss *systemService) Ping(ctx service.Context) error {
	response, body, err := ctx.GetHttpClient().Get("/v1/system/ping")
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("failed pinging application service. Status code: %d", response.StatusCode)
	}

	log.Output(string(body))
	return nil
}
