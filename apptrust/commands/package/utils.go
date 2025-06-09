package packagecmds

import (
	"strings"

	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
)

// BuildPackageRequestPayload creates a BindPackageRequest from command arguments.
// This function is shared between bind and unbind package commands.
// It expects the following arguments in order: <application_key> <package_type> <package_name> [<package_versions>].
func BuildPackageRequestPayload(ctx *components.Context) (*model.BindPackageRequest, error) {
	applicationKey := ctx.Arguments[0]
	packageType := ctx.Arguments[1]
	packageName := ctx.Arguments[2]

	var versions []string
	if len(ctx.Arguments) > 3 {
		// Parse comma-separated versions
		versions = parseVersions(ctx.Arguments[3])
	}

	return &model.BindPackageRequest{
		ApplicationKey: applicationKey,
		Type:           packageType,
		Name:           packageName,
		Versions:       versions,
	}, nil
}

// parseVersions parses a comma-separated string of versions into a slice.
// It trims whitespaces from each version and filters out empty strings.
func parseVersions(versionsString string) []string {
	parts := strings.Split(versionsString, ",")
	var versions []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			versions = append(versions, trimmed)
		}
	}
	return versions
}
