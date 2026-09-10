package version

import (
	"errors"
	"testing"

	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	mockversions "github.com/jfrog/jfrog-cli-application/apptrust/service/versions/mocks"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/jfrog/jfrog-cli-core/v2/utils/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestExportAppVersionCommand_WrongNumberOfArguments(t *testing.T) {
	tests := []struct {
		name      string
		arguments []string
	}{
		{name: "no arguments", arguments: []string{}},
		{name: "single argument", arguments: []string{"app-key"}},
		{name: "too many arguments", arguments: []string{"app-key", "1.0.0", "./out/", "extra"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx := &components.Context{
				PrintCommandHelp: func(string) error { return nil },
			}
			ctx.Arguments = tt.arguments

			cmd := &exportAppVersionCommand{
				versionService: mockversions.NewMockVersionService(ctrl),
			}

			err := cmd.prepareAndRunCommand(ctx)
			assert.Error(t, err)
		})
	}
}

func TestParseExportDownloadFlags(t *testing.T) {
	tests := []struct {
		name          string
		ctxSetup      func(*components.Context)
		wantMinSplit  int64
		wantSplitCnt  int
		errorContains string
	}{
		{
			name:         "defaults",
			ctxSetup:     func(*components.Context) {},
			wantMinSplit: 5120,
			wantSplitCnt: 3,
		},
		{
			name: "custom values",
			ctxSetup: func(ctx *components.Context) {
				ctx.AddStringFlag(commands.MinSplitFlag, "1024")
				ctx.AddStringFlag(commands.SplitCountFlag, "5")
			},
			wantMinSplit: 1024,
			wantSplitCnt: 5,
		},
		{
			name: "non-numeric min-split",
			ctxSetup: func(ctx *components.Context) {
				ctx.AddStringFlag(commands.MinSplitFlag, "abc")
			},
			errorContains: "the '--min-split' option should have a numeric value",
		},
		{
			name: "non-numeric split-count",
			ctxSetup: func(ctx *components.Context) {
				ctx.AddStringFlag(commands.SplitCountFlag, "abc")
			},
			errorContains: "the '--split-count' option should have a numeric value",
		},
		{
			name: "split-count above max",
			ctxSetup: func(ctx *components.Context) {
				ctx.AddStringFlag(commands.SplitCountFlag, "16")
			},
			errorContains: "maximum of 15",
		},
		{
			name: "negative split-count",
			ctxSetup: func(ctx *components.Context) {
				ctx.AddStringFlag(commands.SplitCountFlag, "-1")
			},
			errorContains: "cannot have a negative value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &components.Context{}
			tt.ctxSetup(ctx)
			minSplit, splitCount, err := parseExportDownloadFlags(ctx)
			if tt.errorContains != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantMinSplit, minSplit)
			assert.Equal(t, tt.wantSplitCnt, splitCount)
		})
	}
}

func TestShouldStopExportPolling(t *testing.T) {
	tests := []struct {
		name     string
		status   *model.AppVersionExportStatus
		wantStop bool
		wantErr  string
	}{
		{
			name:     "in progress keeps polling",
			status:   &model.AppVersionExportStatus{Status: "IN_PROGRESS"},
			wantStop: false,
		},
		{
			name: "completed stops",
			status: &model.AppVersionExportStatus{
				Status:              "COMPLETED",
				RelativeDownloadURL: "/repo/archive.zip",
			},
			wantStop: true,
		},
		{
			name: "failed stops",
			status: &model.AppVersionExportStatus{
				Status:  "FAILED",
				Message: "zip creation failed",
			},
			wantStop: true,
			wantErr:  "zip creation failed",
		},
		{
			name:     "unexpected status stops with error",
			status:   &model.AppVersionExportStatus{Status: "UNKNOWN"},
			wantStop: true,
			wantErr:  "unexpected export status",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStop, err := shouldStopExportPolling(tt.status)
			assert.Equal(t, tt.wantStop, gotStop)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestExportAppVersionCommand_Run_FailedStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	applicationKey := "app-key"
	version := "1.0.0"

	mockVersionService := mockversions.NewMockVersionService(ctrl)
	mockVersionService.EXPECT().TriggerExport(gomock.Any(), applicationKey, version).Return(nil).Times(1)
	mockVersionService.EXPECT().GetExportStatus(gomock.Any(), applicationKey, version).Return(&model.AppVersionExportStatus{
		Status:  "FAILED",
		Message: "zip creation failed",
	}, nil).Times(1)

	cmd := &exportAppVersionCommand{
		versionService: mockVersionService,
		serverDetails:  serverDetails,
		applicationKey: applicationKey,
		version:        version,
	}

	err := cmd.Run()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "zip creation failed")
}

func TestExportAppVersionCommand_Run_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	applicationKey := "app-key"
	version := "1.0.0"

	mockVersionService := mockversions.NewMockVersionService(ctrl)
	mockVersionService.EXPECT().TriggerExport(gomock.Any(), applicationKey, version).
		Return(errors.New("export error")).Times(1)

	cmd := &exportAppVersionCommand{
		versionService: mockVersionService,
		serverDetails:  serverDetails,
		applicationKey: applicationKey,
		version:        version,
	}

	err := cmd.Run()
	assert.Error(t, err)
	assert.Equal(t, "export error", err.Error())
}
