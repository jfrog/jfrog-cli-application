package version

import (
	"errors"
	"testing"

	mockversions "github.com/jfrog/jfrog-cli-application/apptrust/service/versions/mocks"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDeleteAppVersionCommand_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	applicationKey := "app-key"
	version := "1.0.0"

	mockVersionService := mockversions.NewMockVersionService(ctrl)
	mockVersionService.EXPECT().DeleteAppVersion(gomock.Any(), applicationKey, version).
		Return(nil).Times(1)

	cmd := &deleteAppVersionCommand{
		versionService: mockVersionService,
		serverDetails:  serverDetails,
		applicationKey: applicationKey,
		version:        version,
	}

	err := cmd.Run()
	assert.NoError(t, err)
}

func TestDeleteAppVersionCommand_Run_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	applicationKey := "app-key"
	version := "1.0.0"

	mockVersionService := mockversions.NewMockVersionService(ctrl)
	mockVersionService.EXPECT().DeleteAppVersion(gomock.Any(), applicationKey, version).
		Return(errors.New("delete error")).Times(1)

	cmd := &deleteAppVersionCommand{
		versionService: mockVersionService,
		serverDetails:  serverDetails,
		applicationKey: applicationKey,
		version:        version,
	}

	err := cmd.Run()
	assert.Error(t, err)
	assert.Equal(t, "delete error", err.Error())
}
