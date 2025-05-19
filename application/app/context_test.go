package app

import (
	"testing"

	mockapplications "github.com/jfrog/jfrog-cli-application/application/service/applications/mocks"
	mocksystems "github.com/jfrog/jfrog-cli-application/application/service/systems/mocks"
	mockversions "github.com/jfrog/jfrog-cli-application/application/service/versions/mocks"

	"github.com/stretchr/testify/assert"
)

func TestNewAppContext(t *testing.T) {
	ctx := NewAppContext()
	assert.NotNil(t, ctx)
	assert.NotNil(t, ctx.GetApplicationService())
	assert.NotNil(t, ctx.GetVersionService())
	assert.NotNil(t, ctx.GetSystemService())
}

func TestGetApplicationService(t *testing.T) {
	mockApplicationService := &mockapplications.MockApplicationService{}
	ctx := &context{
		applicationService: mockApplicationService,
	}
	assert.Equal(t, mockApplicationService, ctx.GetApplicationService())
}

func TestGetVersionService(t *testing.T) {
	mockVersionService := &mockversions.MockVersionService{}
	ctx := &context{
		versionService: mockVersionService,
	}
	assert.Equal(t, mockVersionService, ctx.GetVersionService())
}

func TestGetSystemService(t *testing.T) {
	mockSystemService := &mocksystems.MockSystemService{}
	ctx := &context{
		systemService: mockSystemService,
	}
	assert.Equal(t, mockSystemService, ctx.GetSystemService())
}

func TestGetConfig(t *testing.T) {
	ctx := &context{}
	assert.Nil(t, ctx.GetConfig())
}
