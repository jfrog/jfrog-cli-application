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

func TestCreateAppCommand_Run_Flags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := &components.Context{
		Arguments: []string{"app-key"},
	}
	ctx.AddStringFlag("application-name", "test-app")
	ctx.AddStringFlag("project", "test-project")
	ctx.AddStringFlag("desc", "Test application")
	ctx.AddStringFlag("business-criticality", "high")
	ctx.AddStringFlag("maturity-level", "production")
	ctx.AddStringFlag("labels", "env=prod;region=us-east")
	ctx.AddStringFlag("user-owners", "john.doe;jane.smith")
	ctx.AddStringFlag("group-owners", "devops;security")

	requestPayload := &model.AppDescriptor{
		ApplicationKey:      "app-key",
		ApplicationName:     "test-app",
		ProjectKey:          "test-project",
		Description:         "Test application",
		BusinessCriticality: "high",
		MaturityLevel:       "production",
		Labels: map[string]string{
			"env":    "prod",
			"region": "us-east",
		},
		UserOwners:  []string{"john.doe", "jane.smith"},
		GroupOwners: []string{"devops", "security"},
	}

	mockAppService := mockapps.NewMockApplicationService(ctrl)
	mockAppService.EXPECT().CreateApplication(gomock.Any(), requestPayload).Return(nil).Times(1)

	cmd := &createAppCommand{
		serverDetails:      &config.ServerDetails{Url: "https://example.com"},
		applicationService: mockAppService,
		requestBody:        requestPayload,
	}

	err := cmd.prepareAndRunCommand(ctx)
	assert.NoError(t, err)
}

func TestCreateAppCommand_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	requestPayload := &model.AppDescriptor{
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

func TestCreateAppCommand_WrongNumberOfArguments(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	app := cli.NewApp()
	set := flag.NewFlagSet("test", 0)
	ctx := cli.NewContext(app, set, nil)

	mockAppService := mockapps.NewMockApplicationService(ctrl)
	cmd := &createAppCommand{
		applicationService: mockAppService,
	}

	// Test with no arguments
	context, err := components.ConvertContext(ctx)
	assert.NoError(t, err)

	err = cmd.prepareAndRunCommand(context)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Wrong number of arguments")
}

func TestCreateAppCommand_MissingProjectFlag(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := &components.Context{
		Arguments: []string{"app-key"},
	}
	ctx.AddStringFlag("application-name", "test-app")
	mockAppService := mockapps.NewMockApplicationService(ctrl)

	cmd := &createAppCommand{
		serverDetails:      &config.ServerDetails{Url: "https://example.com"},
		applicationService: mockAppService,
	}

	err := cmd.prepareAndRunCommand(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "--project is mandatory")
}

func TestCreateAppCommand_Run_SpecFile(t *testing.T) {
	tests := []struct {
		name           string
		specPath       string
		args           []string
		expectsError   bool
		errorContains  string
		expectsPayload *model.AppDescriptor
	}{
		{
			name:     "minimal spec file",
			specPath: "./testfiles/minimal-spec.json",
			args:     []string{"app-min"},
			expectsPayload: &model.AppDescriptor{
				ApplicationKey:  "app-min",
				ApplicationName: "app-min",
				ProjectKey:      "test-project",
			},
		},
		{
			name:     "full spec file",
			specPath: "./testfiles/full-spec.json",
			args:     []string{"app-full"},
			expectsPayload: &model.AppDescriptor{
				ApplicationKey:      "app-full",
				ApplicationName:     "test-app-full",
				ProjectKey:          "test-project",
				Description:         "A comprehensive test application",
				MaturityLevel:       "production",
				BusinessCriticality: "high",
				Labels: map[string]string{
					"environment": "production",
					"region":      "us-east-1",
					"team":        "devops",
				},
				UserOwners:  []string{"john.doe", "jane.smith"},
				GroupOwners: []string{"devops-team", "security-team"},
			},
		},
		{
			name:          "invalid spec file",
			specPath:      "./testfiles/invalid-spec.json",
			args:          []string{"app-invalid"},
			expectsError:  true,
			errorContains: "unexpected end of JSON input",
		},
		{
			name:          "missing project key",
			specPath:      "./testfiles/missing-project-spec.json",
			args:          []string{"app-no-project"},
			expectsError:  true,
			errorContains: "project_key is mandatory in spec file",
		},
		{
			name:          "non-existent spec file",
			specPath:      "./testfiles/non-existent.json",
			args:          []string{"app-nonexistent"},
			expectsError:  true,
			errorContains: "no such file or directory",
		},
		{
			name:     "spec with application_key that should be ignored",
			specPath: "./testfiles/spec-with-app-key.json",
			args:     []string{"command-line-app-key"},
			expectsPayload: &model.AppDescriptor{
				ApplicationKey:  "command-line-app-key",
				ApplicationName: "test-app",
				ProjectKey:      "test-project",
				Description:     "A test application with application_key that should be ignored",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := &components.Context{
				Arguments: tt.args,
			}
			ctx.AddStringFlag("spec", tt.specPath)

			var actualPayload *model.AppDescriptor
			mockAppService := mockapps.NewMockApplicationService(ctrl)
			if !tt.expectsError {
				mockAppService.EXPECT().CreateApplication(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ interface{}, req *model.AppDescriptor) error {
						actualPayload = req
						return nil
					}).Times(1)
			}

			cmd := &createAppCommand{
				serverDetails:      &config.ServerDetails{Url: "https://example.com"},
				applicationService: mockAppService,
			}

			err := cmd.prepareAndRunCommand(ctx)
			if tt.expectsError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectsPayload, actualPayload)
			}
		})
	}
}

func TestCreateAppCommand_Run_SpecVars(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedPayload := &model.AppDescriptor{
		ApplicationKey:      "app-with-vars",
		ApplicationName:     "test-app",
		ProjectKey:          "test-project",
		Description:         "A test application for production",
		MaturityLevel:       "production",
		BusinessCriticality: "high",
		Labels: map[string]string{
			"environment": "production",
			"region":      "us-east-1",
		},
	}

	ctx := &components.Context{
		Arguments: []string{"app-with-vars"},
	}
	ctx.AddStringFlag("spec", "./testfiles/with-vars-spec.json")
	ctx.AddStringFlag("spec-vars", "PROJECT_KEY=test-project;APP_NAME=test-app;ENVIRONMENT=production;MATURITY_LEVEL=production;CRITICALITY=high;REGION=us-east-1")

	var actualPayload *model.AppDescriptor
	mockAppService := mockapps.NewMockApplicationService(ctrl)
	mockAppService.EXPECT().CreateApplication(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ interface{}, req *model.AppDescriptor) error {
			actualPayload = req
			return nil
		}).Times(1)

	cmd := &createAppCommand{
		serverDetails:      &config.ServerDetails{Url: "https://example.com"},
		applicationService: mockAppService,
	}

	err := cmd.prepareAndRunCommand(ctx)
	assert.NoError(t, err)
	assert.Equal(t, expectedPayload, actualPayload)
}

func TestCreateAppCommand_Error_SpecAndFlags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testSpecPath := "./testfiles/minimal-spec.json"
	ctx := &components.Context{
		Arguments: []string{"app-key"},
	}
	ctx.AddStringFlag("spec", testSpecPath)
	ctx.AddStringFlag("project", "test-project")
	ctx.AddStringFlag("url", "https://example.com")

	mockAppService := mockapps.NewMockApplicationService(ctrl)

	cmd := &createAppCommand{
		serverDetails:      &config.ServerDetails{Url: "https://example.com"},
		applicationService: mockAppService,
	}

	err := cmd.prepareAndRunCommand(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "the flag --project is not allowed when --spec is provided")
}
