package version

import (
	"encoding/json"
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
		Sources: &model.CreateVersionSources{
			Packages: []model.CreateVersionPackage{
				{
					Type:       "type",
					Name:       "name",
					Version:    "1.0.0",
					Repository: "repo",
				},
			},
		},
	}

	mockVersionService := mockversions.NewMockVersionService(ctrl)
	mockVersionService.EXPECT().CreateAppVersion(gomock.Any(), requestPayload, "", true).
		Return(nil).Times(1)

	cmd := &createAppVersionCommand{
		versionService: mockVersionService,
		serverDetails:  serverDetails,
		requestPayload: requestPayload,
		signingKey:     "",
		sync:           true,
	}

	err := cmd.Run()
	assert.NoError(t, err)
}

func TestCreateAppVersionCommand_Run_WithSigningKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	requestPayload := &model.CreateAppVersionRequest{
		ApplicationKey: "app-key",
		Version:        "1.0.0",
		Sources: &model.CreateVersionSources{
			Packages: []model.CreateVersionPackage{
				{
					Type:       "type",
					Name:       "name",
					Version:    "1.0.0",
					Repository: "repo",
				},
			},
		},
	}

	mockVersionService := mockversions.NewMockVersionService(ctrl)
	mockVersionService.EXPECT().CreateAppVersion(gomock.Any(), requestPayload, "key1", true).
		Return(nil).Times(1)

	cmd := &createAppVersionCommand{
		versionService: mockVersionService,
		serverDetails:  serverDetails,
		requestPayload: requestPayload,
		signingKey:     "key1",
		sync:           true,
	}

	err := cmd.Run()
	assert.NoError(t, err)
}

func TestCreateAppVersionCommand_Run_Async(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	serverDetails := &config.ServerDetails{Url: "https://example.com"}
	requestPayload := &model.CreateAppVersionRequest{
		ApplicationKey: "app-key",
		Version:        "1.0.0",
		Sources: &model.CreateVersionSources{
			Packages: []model.CreateVersionPackage{
				{
					Type:       "type",
					Name:       "name",
					Version:    "1.0.0",
					Repository: "repo",
				},
			},
		},
	}

	mockVersionService := mockversions.NewMockVersionService(ctrl)
	mockVersionService.EXPECT().CreateAppVersion(gomock.Any(), requestPayload, "", false).
		Return(nil).Times(1)

	cmd := &createAppVersionCommand{
		versionService: mockVersionService,
		serverDetails:  serverDetails,
		requestPayload: requestPayload,
		signingKey:     "",
		sync:           false,
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
		Sources: &model.CreateVersionSources{
			Packages: []model.CreateVersionPackage{
				{
					Type:       "type",
					Name:       "name",
					Version:    "1.0.0",
					Repository: "repo",
				},
			},
		},
	}

	mockVersionService := mockversions.NewMockVersionService(ctrl)
	mockVersionService.EXPECT().CreateAppVersion(gomock.Any(), requestPayload, "", true).
		Return(errors.New("context error")).Times(1)

	cmd := &createAppVersionCommand{
		versionService: mockVersionService,
		serverDetails:  serverDetails,
		requestPayload: requestPayload,
		signingKey:     "",
		sync:           true,
	}

	err := cmd.Run()
	assert.Error(t, err)
	assert.Equal(t, "context error", err.Error())
}

func TestParseBuilds(t *testing.T) {
	cmd := &createAppVersionCommand{}

	// Test basic build parsing
	builds, err := cmd.parseBuilds("build1:1.0.0;build2:2.0.0")
	assert.NoError(t, err)
	assert.Len(t, builds, 2)
	assert.Equal(t, "build1", builds[0].Name)
	assert.Equal(t, "1.0.0", builds[0].Number)
	assert.Equal(t, "build2", builds[1].Name)
	assert.Equal(t, "2.0.0", builds[1].Number)

	// Test invalid format
	_, err = cmd.parseBuilds("build1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid build format")

	// Test too many parts
	_, err = cmd.parseBuilds("build1:1.0.0:timestamp:extra")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid build format")
}

func TestParseReleaseBundles(t *testing.T) {
	cmd := &createAppVersionCommand{}

	// Test basic release bundle parsing
	rbs, err := cmd.parseReleaseBundles("rb1:1.0.0;rb2:2.0.0")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(rbs))
	assert.Equal(t, "rb1", rbs[0].Name)
	assert.Equal(t, "1.0.0", rbs[0].Version)
	assert.Equal(t, "rb2", rbs[1].Name)
	assert.Equal(t, "2.0.0", rbs[1].Version)

	// Test invalid format
	_, err = cmd.parseReleaseBundles("rb1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid release bundle format")

	// Test invalid format with too many parts
	_, err = cmd.parseReleaseBundles("rb1:1.0.0:extra")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid release bundle format")
}

func TestParseSourceVersions(t *testing.T) {
	cmd := &createAppVersionCommand{}

	// Test basic source versions parsing
	svs, err := cmd.parseSourceVersions("app1:1.0.0;app2:2.0.0")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(svs))
	assert.Equal(t, "app1", svs[0].ApplicationKey)
	assert.Equal(t, "1.0.0", svs[0].Version)
	assert.Equal(t, "app2", svs[1].ApplicationKey)
	assert.Equal(t, "2.0.0", svs[1].Version)

	// Test invalid format
	_, err = cmd.parseSourceVersions("app1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid source version format")

	// Test invalid format with too many parts
	_, err = cmd.parseSourceVersions("app1:1.0.0:extra")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid source version format")
}

func TestParseExcludedPackages(t *testing.T) {
	cmd := &createAppVersionCommand{}

	// Test basic excluded packages parsing
	excludes, err := cmd.parseExcludedPackages("pkg1:1.0.0;pkg2:2.0.0")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(excludes))
	assert.Equal(t, "pkg1", excludes[0].Name)
	assert.Equal(t, "1.0.0", excludes[0].Version)
	assert.Equal(t, "pkg2", excludes[1].Name)
	assert.Equal(t, "2.0.0", excludes[1].Version)

	// Test invalid format
	_, err = cmd.parseExcludedPackages("pkg1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid exclude package format")

	// Test invalid format with too many parts
	_, err = cmd.parseExcludedPackages("pkg1:1.0.0:extra")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid exclude package format")
}

func TestCreateVersionSpec(t *testing.T) {
	// Test that the struct can be properly unmarshalled
	specData := `{
		"packages": [{"name": "pkg1", "version": "1.0.0", "repository": "repo1", "type": "npm"}],
		"builds": [{"name": "build1", "number": "5"}],
		"release_bundles": [{"name": "rb1", "version": "1.0.0"}],
		"versions": [{"application_key": "app1", "version": "1.0.0"}],
		"exclude": [{"name": "excluded-pkg", "version": "2.0.0"}]
	}`

	spec := new(createVersionSpec)
	err := json.Unmarshal([]byte(specData), spec)
	assert.NoError(t, err)

	assert.Len(t, spec.Packages, 1)
	assert.Equal(t, "pkg1", spec.Packages[0].Name)

	assert.Len(t, spec.Builds, 1)
	assert.Equal(t, "build1", spec.Builds[0].Name)

	assert.Len(t, spec.ReleaseBundles, 1)
	assert.Equal(t, "rb1", spec.ReleaseBundles[0].Name)

	assert.Len(t, spec.Versions, 1)
	assert.Equal(t, "app1", spec.Versions[0].ApplicationKey)

	assert.Len(t, spec.Exclude, 1)
	assert.Equal(t, "excluded-pkg", spec.Exclude[0].Name)
}
