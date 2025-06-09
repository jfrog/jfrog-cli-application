package packages

import (
	"errors"
	"net/http"
	"testing"

	mockhttp "github.com/jfrog/jfrog-cli-application/apptrust/http/mocks"
	mockservice "github.com/jfrog/jfrog-cli-application/apptrust/service/mocks"
	"go.uber.org/mock/gomock"

	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/stretchr/testify/assert"
)

func TestBindPackage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := NewPackageService()

	tests := []struct {
		name          string
		request       *model.BindPackageRequest
		mockResponse  *http.Response
		mockError     error
		expectedError string
	}{
		{
			name: "success",
			request: &model.BindPackageRequest{
				ApplicationKey: "test-app",
				Type:           "npm",
				Name:           "test-package",
				Version:        "1.0.0",
			},
			mockResponse:  &http.Response{StatusCode: 201},
			mockError:     nil,
			expectedError: "",
		},
		{
			name: "failed with non-200 status code",
			request: &model.BindPackageRequest{
				ApplicationKey: "test-app",
				Type:           "npm",
				Name:           "test-package",
				Version:        "1.0.0",
			},
			mockResponse:  &http.Response{StatusCode: 400},
			mockError:     nil,
			expectedError: "failed to bind package. Status code: 400",
		},
		{
			name: "http client error",
			request: &model.BindPackageRequest{
				ApplicationKey: "test-app",
				Type:           "npm",
				Name:           "test-package",
				Version:        "1.0.0",
			},
			mockResponse:  nil,
			mockError:     errors.New("http client error"),
			expectedError: "http client error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHttpClient := mockhttp.NewMockApptrustHttpClient(ctrl)
			mockHttpClient.EXPECT().Post("/v1/package", tt.request).
				Return(tt.mockResponse, []byte(""), tt.mockError).Times(1)

			mockCtx := mockservice.NewMockContext(ctrl)
			mockCtx.EXPECT().GetHttpClient().Return(mockHttpClient).Times(1)

			err := service.BindPackage(mockCtx, tt.request)
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}

func TestUnbindPackage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	service := NewPackageService()

	tests := []struct {
		name          string
		request       *model.BindPackageRequest
		mockResponse  *http.Response
		mockError     error
		expectedError string
	}{
		{
			name: "success",
			request: &model.BindPackageRequest{
				ApplicationKey: "test-app",
				Type:           "npm",
				Name:           "test-package",
				Version:        "1.0.0",
			},
			mockResponse:  &http.Response{StatusCode: 204},
			mockError:     nil,
			expectedError: "",
		},
		{
			name: "failed with non-200 status code",
			request: &model.BindPackageRequest{
				ApplicationKey: "test-app",
				Type:           "npm",
				Name:           "test-package",
				Version:        "1.0.0",
			},
			mockResponse:  &http.Response{StatusCode: 400},
			mockError:     nil,
			expectedError: "failed to unbind package. Status code: 400",
		},
		{
			name: "http client error",
			request: &model.BindPackageRequest{
				ApplicationKey: "test-app",
				Type:           "npm",
				Name:           "test-package",
				Version:        "1.0.0",
			},
			mockResponse:  nil,
			mockError:     errors.New("http client error"),
			expectedError: "http client error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHttpClient := mockhttp.NewMockApptrustHttpClient(ctrl)
			mockHttpClient.EXPECT().Delete("/v1/package", tt.request).
				Return(tt.mockResponse, []byte(""), tt.mockError).Times(1)

			mockCtx := mockservice.NewMockContext(ctrl)
			mockCtx.EXPECT().GetHttpClient().Return(mockHttpClient).Times(1)

			err := service.UnbindPackage(mockCtx, tt.request)
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}
