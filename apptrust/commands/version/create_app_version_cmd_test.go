package version

import (
	"errors"
	"testing"

	mockversions "github.com/jfrog/jfrog-cli-application/apptrust/service/versions/mocks"
	"go.uber.org/mock/gomock"

	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/stretchr/testify/assert"
)

func TestCreateAppVersionCommand_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	requestPayload := &model.CreateAppVersionRequest{
		ApplicationKey: "app-key",
		Version:        "1.0.0",
		Packages: []model.CreateVersionPackage{
			{
				Type:       "type",
				Name:       "name",
				Version:    "1.0.0",
				Repository: "repo",
			},
		},
	}

	mockVersionService := mockversions.NewMockVersionService(ctrl)
	mockVersionService.EXPECT().CreateAppVersion(gomock.Any(), requestPayload).
		Return(nil).Times(1)

	cmd := &createAppVersionCommand{
		versionService: mockVersionService,
		serverDetails:  serverDetails,
		requestPayload: requestPayload,
	}

	err := cmd.Run()
	assert.NoError(t, err)
}

func TestCreateAppVersionCommand_Run_ContextError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	requestPayload := &model.CreateAppVersionRequest{
		ApplicationKey: "app-key",
		Version:        "1.0.0",
		Packages: []model.CreateVersionPackage{
			{
				Type:       "type",
				Name:       "name",
				Version:    "1.0.0",
				Repository: "repo",
			},
		},
	}

	mockVersionService := mockversions.NewMockVersionService(ctrl)
	mockVersionService.EXPECT().CreateAppVersion(gomock.Any(), requestPayload).
		Return(errors.New("context error")).Times(1)

	cmd := &createAppVersionCommand{
		versionService: mockVersionService,
		serverDetails:  serverDetails,
		requestPayload: requestPayload,
	}

	err := cmd.Run()
	assert.Error(t, err)
	assert.Equal(t, "context error", err.Error())
}
