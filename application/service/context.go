package service

//go:generate ${PROJECT_DIR}/scripts/mockgen.sh ${GOFILE}

import (
	"github.com/jfrog/jfrog-cli-application/application/http"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
)

type Context interface {
	GetServerDetails() coreConfig.ServerDetails
	GetHttpClient() http.AppHttpClient
}

type context struct {
	ServerDetails coreConfig.ServerDetails
	HttpClient    http.AppHttpClient
}

func (c *context) GetServerDetails() coreConfig.ServerDetails {
	return c.ServerDetails
}

func (c *context) GetHttpClient() http.AppHttpClient {
	return c.HttpClient
}

func NewContext(serverDetails coreConfig.ServerDetails) (Context, error) {
	httpClient, err := http.NewAppHttpClient(&serverDetails)
	if err != nil {
		return nil, err
	}

	return &context{ServerDetails: serverDetails, HttpClient: httpClient}, nil
}
