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
			requestPayload := model.NewReleaseAppVersionRequest(
				model.PromotionTypeCopy,
				nil, // includedRepos
				nil, // excludedRepos
				nil, // artifactProps
			)

			mockVersionService := mockversions.NewMockVersionService(ctrl)
			mockVersionService.EXPECT().ReleaseAppVersion(gomock.Any(), applicationKey, version, requestPayload, tt.sync).
				Return(nil).Times(1)

			cmd := &releaseAppVersionCommand{
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	applicationKey := "app-key"
	version := "1.0.0"
	requestPayload := model.NewReleaseAppVersionRequest(
		model.PromotionTypeCopy,
		nil, // includedRepos
		nil, // excludedRepos
		nil, // artifactProps
	)
	expectedError := errors.New("service error occurred")

	mockVersionService := mockversions.NewMockVersionService(ctrl)
	mockVersionService.EXPECT().ReleaseAppVersion(gomock.Any(), applicationKey, version, requestPayload, false).
		Return(expectedError).Times(1)

	cmd := &releaseAppVersionCommand{
		versionService: mockVersionService,
		serverDetails:  serverDetails,
		applicationKey: applicationKey,
		version:        version,
		requestPayload: requestPayload,
		sync:           false,
	}

	err := cmd.Run()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service error occurred")
}
