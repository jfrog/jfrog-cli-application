package service

import (
	"fmt"

	"github.com/jfrog/jfrog-cli-application/application/http"
)

type SystemService interface {
	Ping(ctx *Context) error
}

type systemService struct{}

func NewSystemService() SystemService {
	return &systemService{}
}

func (ss *systemService) Ping(ctx *Context) error {
	httpClient, err := http.NewAppHttpClient(ctx.ServerDetails)
	if err != nil {
		return err
	}

	response, body, err := httpClient.Get("/v1/system/ping")
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("failed to create app version. Status code: %d", response.StatusCode)
	}

	fmt.Println(string(body))
	return nil
}
