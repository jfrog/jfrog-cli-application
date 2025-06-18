package version

import (
	"errors"
	"testing"

	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands/utils"
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

			cmd := &releaseAppVersionCommand{
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
	cmd := &releaseAppVersionCommand{
		serverDetails: serverDetails,
	}

	details, err := cmd.ServerDetails()
	assert.NoError(t, err)
	assert.Equal(t, serverDetails, details)
}

func TestReleaseAppVersionCommand_CommandName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cmd := &releaseAppVersionCommand{}
	assert.Equal(t, commands.VersionRelease, cmd.CommandName())
}

func TestReleaseAppVersionCommand_BuildRequestPayload(t *testing.T) {
	// Test the key validation logic used in buildRequestPayload
	t.Run("validation test", func(t *testing.T) {
		// Let's define the actual allowed values for promotion type
		allowedValues := []string{model.PromotionTypeCopy, model.PromotionTypeMove}

		// Case 1: Empty string should return the default value
		result1, err1 := utils.ValidateEnumFlag(
			"test-flag",
			"",
			model.PromotionTypeCopy,
			allowedValues)

		assert.NoError(t, err1)
		assert.Equal(t, model.PromotionTypeCopy, result1)

		// Case 2: Valid value should be returned as is
		result2, err2 := utils.ValidateEnumFlag(
			"test-flag",
			model.PromotionTypeMove,
			model.PromotionTypeCopy,
			allowedValues)

		assert.NoError(t, err2)
		assert.Equal(t, model.PromotionTypeMove, result2)

		// Case 3: Invalid value should return an error
		result3, err3 := utils.ValidateEnumFlag(
			"test-flag",
			"invalid-type",
			model.PromotionTypeCopy,
			allowedValues)

		assert.Error(t, err3)
		assert.Equal(t, "", result3) // On error, the result is empty
		assert.Contains(t, err3.Error(), "invalid value")
	})

	t.Run("property parsing test", func(t *testing.T) {
		// Test property parsing behavior
		props, err := utils.ParseMapFlag("key1=value1;key2=value2")
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{"key1": "value1", "key2": "value2"}, props)

		_, err = utils.ParseMapFlag("invalid-format")
		assert.Error(t, err)
	})
}
