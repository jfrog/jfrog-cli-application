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
func ParseArtifactProps(ctx *components.Context) ([]model.ArtifactProperty, error) {
	if propsStr := ctx.GetStringFlagValue(commands.PropsFlag); propsStr != "" {
		props, err := utils.ParseListPropertiesFlag(propsStr)
		if err != nil {
			return nil, errorutils.CheckErrorf("failed to parse properties: %s", err.Error())
		}

		var artifactProps []model.ArtifactProperty
		for key, values := range props {
			artifactProps = append(artifactProps, model.ArtifactProperty{
				Key:    key,
				Values: values,
			})
		}
		return artifactProps, nil
	}
	return nil, nil
}

// ParseOverwriteStrategy extracts and validates the overwrite strategy from command context
func ParseOverwriteStrategy(ctx *components.Context) (string, error) {
	overwriteStrategy := ctx.GetStringFlagValue(commands.OverwriteStrategyFlag)
	if overwriteStrategy == "" {
		return "", nil
	}

	validatedStrategy, err := utils.ValidateEnumFlag(commands.OverwriteStrategyFlag, overwriteStrategy, "", model.OverwriteStrategyValues)
	if err != nil {
		return "", err
	}

	return validatedStrategy, nil
}
