package version

import (
	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands/utils"
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
)

// BuildPromotionParams extracts common promotion parameters from command context
// Used by both promote and release commands
func BuildPromotionParams(ctx *components.Context) (string, []string, []string, error) {
	var includedRepos []string
	var excludedRepos []string

	if includeReposStr := ctx.GetStringFlagValue(commands.IncludeReposFlag); includeReposStr != "" {
		includedRepos = utils.ParseSliceFlag(includeReposStr)
	}

	if excludeReposStr := ctx.GetStringFlagValue(commands.ExcludeReposFlag); excludeReposStr != "" {
		excludedRepos = utils.ParseSliceFlag(excludeReposStr)
	}

	promotionType := ctx.GetStringFlagValue(commands.PromotionTypeFlag)

	validatedPromotionType, err := utils.ValidateEnumFlag(commands.PromotionTypeFlag, promotionType, model.PromotionTypeCopy, model.PromotionTypeValues)
	if err != nil {
		return "", nil, nil, err
	}

	// If dry-run is true, override with dry_run
	dryRun := ctx.GetBoolFlagValue(commands.DryRunFlag)
	if dryRun {
		validatedPromotionType = model.PromotionTypeDryRun
	}

	return validatedPromotionType, includedRepos, excludedRepos, nil
}

// ParseArtifactProps extracts artifact properties from command context
func ParseArtifactProps(ctx *components.Context) (map[string]string, error) {
	if propsStr := ctx.GetStringFlagValue(commands.PropsFlag); propsStr != "" {
		props, err := utils.ParseMapFlag(propsStr)
		if err != nil {
			return nil, errorutils.CheckErrorf("failed to parse properties: %s", err.Error())
		}
		return props, nil
	}
	return nil, nil
}
