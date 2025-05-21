package application

import (
	"errors"
	"flag"
	"testing"

	"github.com/urfave/cli"

	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	mockapps "github.com/jfrog/jfrog-cli-application/apptrust/service/applications/mocks"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUpdateAppCommand_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	appKey := "app-key"
	requestPayload := &model.AppDescriptor{
		ApplicationKey:      appKey,
		ApplicationName:     "app-name",
		Description:         "Updated description",
		MaturityLevel:       "production",
		BusinessCriticality: "high",
		Labels: map[string]string{
			"environment": "production",
			"region":      "us-east",
		},
		UserOwners:  []string{"JohnD", "Dave Rice"},
		GroupOwners: []string{"DevOps"},
	}

	mockAppService := mockapps.NewMockApplicationService(ctrl)
	mockAppService.EXPECT().UpdateApplication(gomock.Any(), requestPayload).Return(nil).Times(1)

	cmd := &updateAppCommand{
		applicationService: mockAppService,
		serverDetails:      serverDetails,
		requestBody:        requestPayload,
	}

	err := cmd.Run()
	assert.NoError(t, err)
}

func TestUpdateAppCommand_Run_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	appKey := "app-key"
	requestPayload := &model.AppDescriptor{
		ApplicationKey:      appKey,
		ApplicationName:     "app-name",
		Description:         "Updated description",
		MaturityLevel:       "production",
		BusinessCriticality: "high",
		Labels: map[string]string{
			"environment": "production",
			"region":      "us-east",
		},
		UserOwners:  []string{"JohnD", "Dave Rice"},
		GroupOwners: []string{"DevOps"},
	}

	mockAppService := mockapps.NewMockApplicationService(ctrl)
	mockAppService.EXPECT().UpdateApplication(gomock.Any(), requestPayload).Return(errors.New("failed to update application. Status code: 500")).Times(1)

	cmd := &updateAppCommand{
		applicationService: mockAppService,
		serverDetails:      serverDetails,
		requestBody:        requestPayload,
	}

	err := cmd.Run()
	assert.Error(t, err)
	assert.Equal(t, "failed to update application. Status code: 500", err.Error())
}

func TestUpdateAppCommand_WrongNumberOfArguments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	app := cli.NewApp()
	set := flag.NewFlagSet("test", 0)
	ctx := cli.NewContext(app, set, nil)

	mockAppService := mockapps.NewMockApplicationService(ctrl)
	cmd := &updateAppCommand{
		applicationService: mockAppService,
	}

	// Test with no arguments
	context, err := components.ConvertContext(ctx)
	assert.NoError(t, err)

	err = cmd.prepareAndRunCommand(context)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Wrong number of arguments")
}
