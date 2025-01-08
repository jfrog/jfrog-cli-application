package versions

import (
	"errors"
	"net/http"
	"testing"

	mockhttp "github.com/jfrog/jfrog-cli-application/application/http/mocks"
	mockservice "github.com/jfrog/jfrog-cli-application/application/service/mocks"
	"go.uber.org/mock/gomock"

	"github.com/jfrog/jfrog-cli-application/application/model"
	"github.com/stretchr/testify/assert"
)

func TestCreateAppVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := NewVersionService()

	tests := []struct {
		name             string
		request          *model.CreateAppVersionRequest
		mockResponse     *http.Response
		mockResponseBody string
		mockError        error
		expectedError    string
	}{
		{
			name:             "success",
			request:          &model.CreateAppVersionRequest{},
			mockResponse:     &http.Response{StatusCode: 201},
			mockResponseBody: "{}",
			mockError:        nil,
			expectedError:    "",
		},
		{
			name:             "failure",
			request:          &model.CreateAppVersionRequest{},
			mockResponse:     &http.Response{StatusCode: 400},
			mockResponseBody: "error",
			mockError:        nil,
			expectedError:    "failed to create app version",
		},
		{
			name:             "http client error",
			request:          &model.CreateAppVersionRequest{},
			mockResponse:     nil,
			mockResponseBody: "",
			mockError:        errors.New("http client error"),
			expectedError:    "http client error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHttpClient := mockhttp.NewMockAppHttpClient(ctrl)
			mockHttpClient.EXPECT().Post("/v1/version", tt.request).
				Return(tt.mockResponse, []byte(tt.mockResponseBody), tt.mockError).Times(1)

			mockCtx := mockservice.NewMockContext(ctrl)
			mockCtx.EXPECT().GetHttpClient().Return(mockHttpClient).Times(1)

			err := service.CreateAppVersion(mockCtx, tt.request)
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}

func TestPromoteAppVersion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := NewVersionService()

	tests := []struct {
		name             string
		payload          *model.PromoteAppVersionRequest
		mockResponse     *http.Response
		mockResponseBody string
		mockError        error
		expectedError    string
	}{
		{
			name:             "success",
			payload:          &model.PromoteAppVersionRequest{},
			mockResponse:     &http.Response{StatusCode: 200},
			mockResponseBody: "{}",
			mockError:        nil,
			expectedError:    "",
		},
		{
			name:             "failure",
			payload:          &model.PromoteAppVersionRequest{},
			mockResponse:     &http.Response{StatusCode: 400},
			mockResponseBody: "error",
			mockError:        nil,
			expectedError:    "failed to promote app version",
		},
		{
			name:             "http client error",
			payload:          &model.PromoteAppVersionRequest{},
			mockResponse:     nil,
			mockResponseBody: "",
			mockError:        errors.New("http client error"),
			expectedError:    "http client error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHttpClient := mockhttp.NewMockAppHttpClient(ctrl)
			mockHttpClient.EXPECT().Post("/v1/version/promote", tt.payload).
				Return(tt.mockResponse, []byte(tt.mockResponseBody), tt.mockError).Times(1)

			mockCtx := mockservice.NewMockContext(ctrl)
			mockCtx.EXPECT().GetHttpClient().Return(mockHttpClient).Times(1)

			err := service.PromoteAppVersion(mockCtx, tt.payload)
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}
