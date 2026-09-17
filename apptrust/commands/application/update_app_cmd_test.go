package application

import (
	"errors"
	"flag"
	"testing"

	"github.com/urfave/cli"

	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
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
	description := "Updated description"
	maturityLevel := "production"
	businessCriticality := "high"
	requestPayload := &model.AppDescriptor{
		ApplicationKey:      appKey,
		ApplicationName:     "app-name",
		Description:         &description,
		MaturityLevel:       &maturityLevel,
		BusinessCriticality: &businessCriticality,
		Labels: &[]model.LabelEntry{
			{Key: "environment", Value: "production"},
			{Key: "region", Value: "us-east"},
		},
		UserOwners:  &[]string{"JohnD", "Dave Rice"},
		GroupOwners: &[]string{"DevOps"},
	}

	mockAppService := mockapps.NewMockApplicationService(ctrl)
	mockAppService.EXPECT().UpdateApplication(gomock.Any(), requestPayload).Return(nil, nil).Times(1)

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
	description := "Updated description"
	maturityLevel := "production"
	businessCriticality := "high"
	requestPayload := &model.AppDescriptor{
		ApplicationKey:      appKey,
		ApplicationName:     "app-name",
		Description:         &description,
		MaturityLevel:       &maturityLevel,
		BusinessCriticality: &businessCriticality,
		Labels: &[]model.LabelEntry{
			{Key: "environment", Value: "production"},
			{Key: "region", Value: "us-east"},
		},
		UserOwners:  &[]string{"JohnD", "Dave Rice"},
		GroupOwners: &[]string{"DevOps"},
	}

	mockAppService := mockapps.NewMockApplicationService(ctrl)
	mockAppService.EXPECT().UpdateApplication(gomock.Any(), requestPayload).Return(nil, errors.New("failed to update application. Status code: 500")).Times(1)

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

func TestUpdateAppCommand_FlagsSuite(t *testing.T) {
	monitorPolicyValue := 5

	tests := []struct {
		name           string
		ctxSetup       func(*components.Context)
		expectsError   bool
		errorContains  string
		expectsPayload *model.AppDescriptor
	}{
		{
			name: "add-labels only - single label",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key"}
				ctx.AddStringFlag(commands.AddLabelsFlag, "environment=production")
			},
			expectsPayload: &model.AppDescriptor{
				ApplicationKey: "app-key",
				LabelUpdates: &model.LabelUpdates{
					Add: []model.LabelEntry{
						{Key: "environment", Value: "production"},
					},
				},
			},
		},
		{
			name: "add-labels only - multiple labels",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key"}
				ctx.AddStringFlag(commands.AddLabelsFlag, "environment=production;region=us-east")
			},
			expectsPayload: &model.AppDescriptor{
				ApplicationKey: "app-key",
				LabelUpdates: &model.LabelUpdates{
					Add: []model.LabelEntry{
						{Key: "environment", Value: "production"},
						{Key: "region", Value: "us-east"},
					},
				},
			},
		},
		{
			name: "add-labels only - same key multiple values",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key"}
				ctx.AddStringFlag(commands.AddLabelsFlag, "environment=production;environment=staging;region=us-east")
			},
			expectsPayload: &model.AppDescriptor{
				ApplicationKey: "app-key",
				LabelUpdates: &model.LabelUpdates{
					Add: []model.LabelEntry{
						{Key: "environment", Value: "production"},
						{Key: "environment", Value: "staging"},
						{Key: "region", Value: "us-east"},
					},
				},
			},
		},
		{
			name: "remove-labels only - single label",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key"}
				ctx.AddStringFlag(commands.RemoveLabelsFlag, "infra-version=v1.0")
			},
			expectsPayload: &model.AppDescriptor{
				ApplicationKey: "app-key",
				LabelUpdates: &model.LabelUpdates{
					Remove: []model.LabelEntry{
						{Key: "infra-version", Value: "v1.0"},
					},
				},
			},
		},
		{
			name: "remove-labels only - multiple labels",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key"}
				ctx.AddStringFlag(commands.RemoveLabelsFlag, "infra-version=v1.0;region=us-west")
			},
			expectsPayload: &model.AppDescriptor{
				ApplicationKey: "app-key",
				LabelUpdates: &model.LabelUpdates{
					Remove: []model.LabelEntry{
						{Key: "infra-version", Value: "v1.0"},
						{Key: "region", Value: "us-west"},
					},
				},
			},
		},
		{
			name: "both add-labels and remove-labels",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key"}
				ctx.AddStringFlag(commands.AddLabelsFlag, "environment=production;environment=staging;region=us-east")
				ctx.AddStringFlag(commands.RemoveLabelsFlag, "infra-version=v1.0;region=us-west")
			},
			expectsPayload: &model.AppDescriptor{
				ApplicationKey: "app-key",
				LabelUpdates: &model.LabelUpdates{
					Add: []model.LabelEntry{
						{Key: "environment", Value: "production"},
						{Key: "environment", Value: "staging"},
						{Key: "region", Value: "us-east"},
					},
					Remove: []model.LabelEntry{
						{Key: "infra-version", Value: "v1.0"},
						{Key: "region", Value: "us-west"},
					},
				},
			},
		},
		{
			name: "no label flags - label_updates should not be in request",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key"}
				ctx.AddStringFlag(commands.DescriptionFlag, "Updated description")
			},
			expectsPayload: &model.AppDescriptor{
				ApplicationKey: "app-key",
				Description:    stringPtr("Updated description"),
			},
		},
		{
			name: "combined with other flags",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key"}
				ctx.AddStringFlag(commands.ApplicationNameFlag, "updated-app-name")
				ctx.AddStringFlag(commands.DescriptionFlag, "Updated description")
				ctx.AddStringFlag(commands.MaturityLevelFlag, "production")
				ctx.AddStringFlag(commands.AddLabelsFlag, "environment=production")
				ctx.AddStringFlag(commands.RemoveLabelsFlag, "old-label=old-value")
			},
			expectsPayload: &model.AppDescriptor{
				ApplicationKey:  "app-key",
				ApplicationName: "updated-app-name",
				Description:     stringPtr("Updated description"),
				MaturityLevel:   stringPtr("production"),
				LabelUpdates: &model.LabelUpdates{
					Add: []model.LabelEntry{
						{Key: "environment", Value: "production"},
					},
					Remove: []model.LabelEntry{
						{Key: "old-label", Value: "old-value"},
					},
				},
			},
		},
		{
			name: "monitor-policy",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key"}
				ctx.AddStringFlag(commands.MonitorPolicyFlag, "type=version_count, value=5")
			},
			expectsPayload: &model.AppDescriptor{
				ApplicationKey: "app-key",
				MonitorPolicy:  &model.MonitorPolicy{Type: model.MonitorPolicyTypeVersionCount, Value: &monitorPolicyValue},
			},
		},
		{
			name: "auto-promote-stages",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key"}
				ctx.AddStringFlag("auto-promote-stages", "DEV;PROD")
			},
			expectsPayload: &model.AppDescriptor{
				ApplicationKey:    "app-key",
				AutoPromoteStages: &[]string{"DEV", "PROD"},
			},
		},
		{
			name: "empty auto-promote-stages clears configuration",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key"}
				ctx.AddStringFlag("auto-promote-stages", "")
			},
			expectsPayload: &model.AppDescriptor{
				ApplicationKey:    "app-key",
				AutoPromoteStages: &[]string{},
			},
		},
		{
			name: "invalid add-labels format - missing equals",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key"}
				ctx.AddStringFlag(commands.AddLabelsFlag, "invalid-format")
			},
			expectsError:  true,
			errorContains: "invalid key-value pair",
		},
		{
			name: "invalid remove-labels format - missing equals",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key"}
				ctx.AddStringFlag(commands.RemoveLabelsFlag, "invalid-format")
			},
			expectsError:  true,
			errorContains: "invalid key-value pair",
		},
		{
			name: "whitespace handling",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key"}
				ctx.AddStringFlag(commands.AddLabelsFlag, " key1 = value1 ; key2 = value2 ")
			},
			expectsPayload: &model.AppDescriptor{
				ApplicationKey: "app-key",
				LabelUpdates: &model.LabelUpdates{
					Add: []model.LabelEntry{
						{Key: "key1", Value: "value1"},
						{Key: "key2", Value: "value2"},
					},
				},
			},
		},
		{
			name: "empty add-labels flag",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key"}
				ctx.AddStringFlag(commands.AddLabelsFlag, "")
				ctx.AddStringFlag(commands.RemoveLabelsFlag, "key=value")
			},
			expectsPayload: &model.AppDescriptor{
				ApplicationKey: "app-key",
				LabelUpdates: &model.LabelUpdates{
					Add: []model.LabelEntry{},
					Remove: []model.LabelEntry{
						{Key: "key", Value: "value"},
					},
				},
			},
		},
		{
			name: "empty remove-labels flag",
			ctxSetup: func(ctx *components.Context) {
				ctx.Arguments = []string{"app-key"}
				ctx.AddStringFlag(commands.AddLabelsFlag, "key=value")
				ctx.AddStringFlag(commands.RemoveLabelsFlag, "")
			},
			expectsPayload: &model.AppDescriptor{
				ApplicationKey: "app-key",
				LabelUpdates: &model.LabelUpdates{
					Add: []model.LabelEntry{
						{Key: "key", Value: "value"},
					},
					Remove: []model.LabelEntry{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := &components.Context{}
			tt.ctxSetup(ctx)
			ctx.AddStringFlag("url", "https://example.com")

			var actualPayload *model.AppDescriptor
			mockAppService := mockapps.NewMockApplicationService(ctrl)
			if !tt.expectsError {
				mockAppService.EXPECT().UpdateApplication(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ interface{}, req *model.AppDescriptor) ([]byte, error) {
						actualPayload = req
						return nil, nil
					}).Times(1)
			}

			cmd := &updateAppCommand{
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

func stringPtr(s string) *string {
	return &s
}
