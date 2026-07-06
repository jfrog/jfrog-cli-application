package version

import (
	"strings"

	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands/utils"
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-core/v2/common/spec"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/jfrog/jfrog-client-go/utils/distribution"
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

	// Convert to uppercase for API request
	return strings.ToUpper(validatedStrategy), nil
}

// ParseConflictResolution extracts and validates the conflict resolution strategy from command context
func ParseConflictResolution(ctx *components.Context) (string, error) {
	conflictResolution := ctx.GetStringFlagValue(commands.ConflictResolutionFlag)
	if conflictResolution == "" {
		return "", nil
	}

	return utils.ValidateEnumFlag(commands.ConflictResolutionFlag, conflictResolution, "", model.ConflictResolutionValues)
}

// ParsePathMappings extracts path mapping rules from the --path-mapping flag.
// Format: "input=(.*), output=stable-release/$1[, package-type=.*]; input=(...), output=..."
// Returns nil if flag is not provided.
func ParsePathMappings(ctx *components.Context) (*model.PromotionModifications, error) {
	const (
		inputField       = "input"
		outputField      = "output"
		packageTypeField = "package-type"
	)

	flagValue := ctx.GetStringFlagValue(commands.PathMappingFlag)
	if flagValue == "" {
		return nil, nil
	}

	entries := utils.ParseSliceFlag(flagValue)
	var mappings []model.PromotionPathMapping

	for i, entry := range entries {
		if entry == "" {
			return nil, errorutils.CheckErrorf("--%s entry %d is empty", commands.PathMappingFlag, i+1)
		}

		entryMap, err := utils.ParseKeyValueString(entry, ",")
		if err != nil {
			return nil, errorutils.CheckErrorf("--%s entry %d: %s", commands.PathMappingFlag, i+1, err.Error())
		}

		input, hasInput := entryMap[inputField]
		output, hasOutput := entryMap[outputField]

		if !hasInput || input == "" {
			return nil, errorutils.CheckErrorf("--%s entry %d: '%s' is required", commands.PathMappingFlag, i+1, inputField)
		}
		if !hasOutput || output == "" {
			return nil, errorutils.CheckErrorf("--%s entry %d: '%s' is required", commands.PathMappingFlag, i+1, outputField)
		}

		mapping := model.PromotionPathMapping{
			Input:  input,
			Output: output,
		}
		if pt, ok := entryMap[packageTypeField]; ok {
			mapping.PackageType = pt
		}

		mappings = append(mappings, mapping)
	}

	return &model.PromotionModifications{Mappings: mappings}, nil
}

func ValidateDistributionFlags(ctx *components.Context) error {
	if ctx.IsFlagSet(commands.DistRulesFlag) &&
		(ctx.IsFlagSet(commands.SiteFlag) || ctx.IsFlagSet(commands.CityFlag) || ctx.IsFlagSet(commands.CountryCodesFlag)) {
		return errorutils.CheckErrorf("the --%s option can't be used with --%s, --%s or --%s",
			commands.DistRulesFlag, commands.SiteFlag, commands.CityFlag, commands.CountryCodesFlag)
	}

	return nil
}

const allSiteName = "*"

func ParseDistributionRules(ctx *components.Context) ([]model.DistributionRule, error) {
	if !ctx.IsFlagSet(commands.DistRulesFlag) {
		rule := model.DistributionRule{
			SiteName:     ctx.GetStringFlagValue(commands.SiteFlag),
			CityName:     ctx.GetStringFlagValue(commands.CityFlag),
			CountryCodes: ctx.GetStringsArrFlagValue(commands.CountryCodesFlag),
		}

		// If no site, city or country codes were provided, default to distributing to all targets.
		if rule.SiteName == "" && rule.CityName == "" && len(rule.CountryCodes) == 0 {
			rule.SiteName = allSiteName
		}
		return []model.DistributionRule{rule}, nil
	}

	distributionRules, err := spec.CreateDistributionRulesFromFile(ctx.GetStringFlagValue(commands.DistRulesFlag))
	if err != nil {
		return nil, err
	}

	modelRules := make([]model.DistributionRule, 0, len(distributionRules.DistributionRules))
	for i := range distributionRules.DistributionRules {
		modelRules = append(modelRules, model.DistributionRule{
			SiteName:     distributionRules.DistributionRules[i].SiteName,
			CityName:     distributionRules.DistributionRules[i].CityName,
			CountryCodes: distributionRules.DistributionRules[i].CountryCodes,
		})
	}
	return modelRules, nil
}

func ParseDistributionModifications(ctx *components.Context) (*model.DistributionModifications, error) {
	pattern := ctx.GetStringFlagValue(commands.MappingPatternFlag)
	target := ctx.GetStringFlagValue(commands.MappingTargetFlag)

	if pattern == "" && target == "" {
		return nil, nil
	}
	if pattern == "" || target == "" {
		return nil, errorutils.CheckErrorf(
			"the --%s and --%s options must be provided together",
			commands.MappingPatternFlag, commands.MappingTargetFlag)
	}

	pathMappings := distribution.CreatePathMappingsFromPatternAndTarget(pattern, target)
	modelMappings := make([]model.DistributionPathMapping, 0, len(pathMappings))
	for _, mapping := range pathMappings {
		modelMappings = append(modelMappings, model.DistributionPathMapping{
			Input:  mapping.Input,
			Output: mapping.Output,
		})
	}
	return &model.DistributionModifications{PathMappings: modelMappings}, nil
}
