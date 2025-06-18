package version

import (
	"errors"
	"testing"

	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	mockversions "github.com/jfrog/jfrog-cli-application/apptrust/service/versions/mocks"
	"go.uber.org/mock/gomock"

	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/stretchr/testify/assert"
)

func TestReleaseAppVersionCommand_Run(t *testing.T) {
	tests := []struct {
		name string
		sync bool
	}{
		{
			name: "sync flag true",
			sync: true,
		},
		{
			name: "sync flag false",
			sync: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			serverDetails := &config.ServerDetails{Url: "https://example.com"}
			applicationKey := "app-key"
			version := "1.0.0"
			requestPayload := &model.ReleaseAppVersionRequest{
				PromotionType: model.PromotionTypeCopy,
			}

			mockVersionService := mockversions.NewMockVersionService(ctrl)
			mockVersionService.EXPECT().ReleaseAppVersion(gomock.Any(), applicationKey, version, requestPayload, tt.sync).
				Return(nil).Times(1)

			cmd := &ReleaseAppVersionCommand{
				versionService: mockVersionService,
				serverDetails:  serverDetails,
				applicationKey: applicationKey,
				version:        version,
				requestPayload: requestPayload,
				sync:           tt.sync,
			}

			err := cmd.Run()
			assert.NoError(t, err)
		})
	}
}

func TestReleaseAppVersionCommand_Run_Error(t *testing.T) {
	tests := []struct {
		name string
		sync bool
	}{
		{
			name: "sync flag true - error",
			sync: true,
		},
		{
			name: "sync flag false - error",
			sync: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			serverDetails := &config.ServerDetails{Url: "https://example.com"}
			applicationKey := "app-key"
			version := "1.0.0"
			requestPayload := &model.ReleaseAppVersionRequest{
				PromotionType: model.PromotionTypeCopy,
			}
			expectedError := errors.New("service error occurred")

			mockVersionService := mockversions.NewMockVersionService(ctrl)
			mockVersionService.EXPECT().ReleaseAppVersion(gomock.Any(), applicationKey, version, requestPayload, tt.sync).
				Return(expectedError).Times(1)

			cmd := &ReleaseAppVersionCommand{
				versionService: mockVersionService,
				serverDetails:  serverDetails,
				applicationKey: applicationKey,
				version:        version,
				requestPayload: requestPayload,
				sync:           tt.sync,
			}

			err := cmd.Run()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "service error occurred")
		})
	}
}

func TestReleaseAppVersionCommand_ServerDetails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{}
	cmd := &ReleaseAppVersionCommand{
		serverDetails: serverDetails,
	}

	details, err := cmd.ServerDetails()
	assert.NoError(t, err)
	assert.Equal(t, serverDetails, details)
}

func TestReleaseAppVersionCommand_CommandName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := &ReleaseAppVersionCommand{}
	assert.Equal(t, commands.ReleaseAppVersion, cmd.CommandName())
}
