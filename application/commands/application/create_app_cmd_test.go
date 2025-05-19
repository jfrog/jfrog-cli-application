package application

import (
	"errors"
	"testing"

	"github.com/jfrog/jfrog-cli-application/application/model"
	mockapps "github.com/jfrog/jfrog-cli-application/application/service/applications/mocks"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCreateAppCommand_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	requestPayload := &model.CreateAppRequest{
		ApplicationKey:  "app-key",
		ApplicationName: "app-name",
		ProjectKey:      "proj-key",
	}

	mockAppService := mockapps.NewMockApplicationService(ctrl)
	mockAppService.EXPECT().CreateApplication(gomock.Any(), requestPayload).Return(nil).Times(1)

	cmd := &createAppCommand{
		applicationService: mockAppService,
		serverDetails:      serverDetails,
		requestBody:        requestPayload,
	}

	err := cmd.Run()
	assert.NoError(t, err)
}

func TestCreateAppCommand_Run_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	requestPayload := &model.CreateAppRequest{
		ApplicationKey:  "app-key",
		ApplicationName: "app-name",
		ProjectKey:      "proj-key",
	}

	mockAppService := mockapps.NewMockApplicationService(ctrl)
	mockAppService.EXPECT().CreateApplication(gomock.Any(), requestPayload).Return(errors.New("failed to create an application. Status code: 500")).Times(1)

	cmd := &createAppCommand{
		applicationService: mockAppService,
		serverDetails:      serverDetails,
		requestBody:        requestPayload,
	}

	err := cmd.Run()
	assert.Error(t, err)
	assert.Equal(t, "failed to create an application. Status code: 500", err.Error())
}
