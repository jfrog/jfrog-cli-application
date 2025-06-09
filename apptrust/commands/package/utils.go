package packagecmds

import (
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
)

// BuildPackageRequestPayload creates a BindPackageRequest from command arguments.
// This function is shared between bind and unbind package commands.
// It expects the following arguments in order: <application_key> <package_type> <package_name> <package_version>.
func BuildPackageRequestPayload(ctx *components.Context) (*model.BindPackageRequest, error) {
	applicationKey := ctx.Arguments[0]
	packageType := ctx.Arguments[1]
	packageName := ctx.Arguments[2]
	packageVersion := ctx.Arguments[3]

	return &model.BindPackageRequest{
		ApplicationKey: applicationKey,
		Type:           packageType,
		Name:           packageName,
		Version:        packageVersion,
	}, nil
}
