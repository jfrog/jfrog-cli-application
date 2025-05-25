package service

import (
	"testing"

	"github.com/jfrog/jfrog-cli-application/apptrust/http"
	coreConfig "github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/stretchr/testify/assert"
)

func TestGetServerDetails(t *testing.T) {
	serverDetails := coreConfig.ServerDetails{
		Url: "https://example.com",
	}
	ctx := &context{ServerDetails: serverDetails}

	assert.Equal(t, serverDetails, ctx.GetServerDetails())
}

func TestGetHttpClient(t *testing.T) {
	serverDetails := coreConfig.ServerDetails{
		Url: "https://example.com",
	}
	httpClient, err := http.NewAppHttpClient(&serverDetails)
	assert.NoError(t, err)

	ctx := &context{ServerDetails: serverDetails, HttpClient: httpClient}

	assert.Equal(t, httpClient, ctx.GetHttpClient())
}

func TestNewContext(t *testing.T) {
	serverDetails := coreConfig.ServerDetails{
		Url: "https://example.com",
	}
	ctx, err := NewContext(serverDetails)
	assert.NoError(t, err)
	assert.NotNil(t, ctx)

	assert.Equal(t, serverDetails, ctx.GetServerDetails())
	assert.NotNil(t, ctx.GetHttpClient())
}
