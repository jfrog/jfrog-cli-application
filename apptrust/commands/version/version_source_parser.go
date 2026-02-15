package version

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands/utils"
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/jfrog/jfrog-cli-core/v2/utils/coreutils"
	"github.com/jfrog/jfrog-client-go/utils/errorutils"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
)

type versionSpec struct {
	Artifacts      []model.CreateVersionArtifact      `json:"artifacts,omitempty"`
	Packages       []model.CreateVersionPackage       `json:"packages,omitempty"`
	Builds         []model.CreateVersionBuild         `json:"builds,omitempty"`
	ReleaseBundles []model.CreateVersionReleaseBundle `json:"release_bundles,omitempty"`
	Versions       []model.CreateVersionReference     `json:"versions,omitempty"`
	Filters        *model.CreateVersionFilters        `json:"filters,omitempty"`
}

// validateNoSpecAndFlagsTogether returns error if both --spec and any other source flag or filter flag are set.
func validateNoSpecAndFlagsTogether(ctx *components.Context) error {
	if ctx.IsFlagSet(commands.SpecFlag) {
		otherSourceFlags := []string{
			commands.SourceTypeBuildsFlag,
			commands.SourceTypeReleaseBundlesFlag,
			commands.SourceTypeApplicationVersionsFlag,
			commands.SourceTypePackagesFlag,
			commands.SourceTypeArtifactsFlag,
		}
		for _, flag := range otherSourceFlags {
			if ctx.IsFlagSet(flag) {
				return errorutils.CheckErrorf("--spec provided: all other source flags (e.g., --%s) are not allowed.", flag)
			}
		}
		if ctx.IsFlagSet(commands.IncludeFilterFlag) {
			return errorutils.CheckErrorf("--spec provided: filter flags (e.g., --%s) are not allowed.", commands.IncludeFilterFlag)
		}
		if ctx.IsFlagSet(commands.ExcludeFilterFlag) {
			return errorutils.CheckErrorf("--spec provided: filter flags (e.g., --%s) are not allowed.", commands.ExcludeFilterFlag)
		}
	}
	return nil
}

// validateAtLeastOneSourceFlag returns error if no source flags or --spec is set.
func validateAtLeastOneSourceFlag(ctx *components.Context) error {
	if !hasSourceFlags(ctx) {
		return errorutils.CheckErrorf(
			"At least one source flag is required. Please provide --%s or at least one of the following: --%s, --%s, --%s, --%s, --%s.",
			commands.SpecFlag, commands.SourceTypeBuildsFlag, commands.SourceTypeReleaseBundlesFlag, commands.SourceTypeApplicationVersionsFlag, commands.SourceTypePackagesFlag, commands.SourceTypeArtifactsFlag)
	}
	return nil
}

func validateRequiredFieldsInMap(m map[string]string, requiredFields ...string) error {
	if m == nil {
		return errorutils.CheckErrorf("missing required fields: %v", strings.Join(requiredFields, ", "))
	}
	for _, field := range requiredFields {
		if _, exists := m[field]; !exists {
			return errorutils.CheckErrorf("missing required field: %s", field)
		}
	}
	return nil
}

// hasSourceFlags returns true if any source flag or --spec is set in the context.
func hasSourceFlags(ctx *components.Context) bool {
	return ctx.IsFlagSet(commands.SpecFlag) ||
		ctx.IsFlagSet(commands.SourceTypeBuildsFlag) ||
		ctx.IsFlagSet(commands.SourceTypeReleaseBundlesFlag) ||
		ctx.IsFlagSet(commands.SourceTypeApplicationVersionsFlag) ||
		ctx.IsFlagSet(commands.SourceTypePackagesFlag) ||
		ctx.IsFlagSet(commands.SourceTypeArtifactsFlag)
}

// buildSourcesAndFiltersFromContext parses sources and filters from either a spec file or CLI flags.
func buildSourcesAndFiltersFromContext(ctx *components.Context) (*model.CreateVersionSources, *model.CreateVersionFilters, error) {
	if ctx.IsFlagSet(commands.SpecFlag) {
		return loadSourcesFromSpec(ctx)
	}
	sources, err := buildSourcesFromFlags(ctx)
	if err != nil {
		return nil, nil, err
	}
	filters, err := buildFiltersFromFlags(ctx)
	if err != nil {
		return nil, nil, err
	}
	return sources, filters, nil
}

func buildSourcesFromFlags(ctx *components.Context) (*model.CreateVersionSources, error) {
	sources := &model.CreateVersionSources{}
	if buildsStr := ctx.GetStringFlagValue(commands.SourceTypeBuildsFlag); buildsStr != "" {
		builds, err := parseBuilds(buildsStr)
		if err != nil {
			return nil, err
		}
		sources.Builds = builds
	}
	if rbStr := ctx.GetStringFlagValue(commands.SourceTypeReleaseBundlesFlag); rbStr != "" {
		releaseBundles, err := parseReleaseBundles(rbStr)
		if err != nil {
			return nil, err
		}
		sources.ReleaseBundles = releaseBundles
	}
	if srcVersionsStr := ctx.GetStringFlagValue(commands.SourceTypeApplicationVersionsFlag); srcVersionsStr != "" {
		sourceVersions, err := parseSourceVersions(srcVersionsStr)
		if err != nil {
			return nil, err
		}
		sources.Versions = sourceVersions
	}
	if packagesStr := ctx.GetStringFlagValue(commands.SourceTypePackagesFlag); packagesStr != "" {
		packages, err := parsePackages(packagesStr)
		if err != nil {
			return nil, err
		}
		sources.Packages = packages
	}
	if artifactsStr := ctx.GetStringFlagValue(commands.SourceTypeArtifactsFlag); artifactsStr != "" {
		artifacts, err := parseArtifacts(artifactsStr)
		if err != nil {
			return nil, err
		}
		sources.Artifacts = artifacts
	}
	return sources, nil
}

func loadSourcesFromSpec(ctx *components.Context) (*model.CreateVersionSources, *model.CreateVersionFilters, error) {
	specFilePath := ctx.GetStringFlagValue(commands.SpecFlag)
	spec := new(versionSpec)
	specVars := coreutils.SpecVarsStringToMap(ctx.GetStringFlagValue(commands.SpecVarsFlag))
	content, err := fileutils.ReadFile(specFilePath)
	if errorutils.CheckError(err) != nil {
		return nil, nil, err
	}

	if len(specVars) > 0 {
		content = coreutils.ReplaceVars(content, specVars)
	}

	err = json.Unmarshal(content, spec)
	if errorutils.CheckError(err) != nil {
		return nil, nil, err
	}

	// Validation: if all sources are empty, return error
	if (len(spec.Packages) == 0) && (len(spec.Builds) == 0) && (len(spec.ReleaseBundles) == 0) && (len(spec.Versions) == 0) && (len(spec.Artifacts) == 0) {
		return nil, nil, errorutils.CheckErrorf("Spec file is empty: must provide at least one source (artifacts, packages, builds, release_bundles, or versions)")
	}

	sources := &model.CreateVersionSources{
		Artifacts:      spec.Artifacts,
		Packages:       spec.Packages,
		Builds:         spec.Builds,
		ReleaseBundles: spec.ReleaseBundles,
		Versions:       spec.Versions,
	}

	return sources, spec.Filters, nil
}

func parseBuilds(buildsStr string) ([]model.CreateVersionBuild, error) {
	const (
		nameField       = "name"
		idField         = "id"
		includeDepField = "include-deps"
		repoKeyField    = "repo-key"
		startedField    = "started"
	)

	var builds []model.CreateVersionBuild
	buildEntries := utils.ParseSliceFlag(buildsStr)
	for _, entry := range buildEntries {
		buildEntryMap, err := utils.ParseKeyValueString(entry, ",")
		if err != nil {
			return nil, errorutils.CheckErrorf("invalid build format: %v", err)
		}
		err = validateRequiredFieldsInMap(buildEntryMap, nameField, idField)
		if err != nil {
			return nil, errorutils.CheckErrorf("invalid build format: %v", err)
		}
		build := model.CreateVersionBuild{
			Name:          buildEntryMap[nameField],
			Number:        buildEntryMap[idField],
			RepositoryKey: buildEntryMap[repoKeyField],
			Started:       buildEntryMap[startedField],
		}
		if _, ok := buildEntryMap[includeDepField]; ok {
			includeDep, err := strconv.ParseBool(buildEntryMap[includeDepField])
			if err != nil {
				return nil, errorutils.CheckErrorf("invalid build format: %v", err)
			}
			build.IncludeDependencies = includeDep
		}
		builds = append(builds, build)
	}
	return builds, nil
}

func parseReleaseBundles(rbStr string) ([]model.CreateVersionReleaseBundle, error) {
	const (
		projectKeyField = "project-key"
		repoKeyField    = "repo-key"
		nameField       = "name"
		versionField    = "version"
	)

	var bundles []model.CreateVersionReleaseBundle
	releaseBundleEntries := utils.ParseSliceFlag(rbStr)
	for _, entry := range releaseBundleEntries {
		releaseBundleEntryMap, err := utils.ParseKeyValueString(entry, ",")
		if err != nil {
			return nil, errorutils.CheckErrorf("invalid release bundle format: %v", err)
		}
		err = validateRequiredFieldsInMap(releaseBundleEntryMap, nameField, versionField)
		if err != nil {
			return nil, errorutils.CheckErrorf("invalid release bundle format: %v", err)
		}
		bundles = append(bundles, model.CreateVersionReleaseBundle{
			ProjectKey:    releaseBundleEntryMap[projectKeyField],
			RepositoryKey: releaseBundleEntryMap[repoKeyField],
			Name:          releaseBundleEntryMap[nameField],
			Version:       releaseBundleEntryMap[versionField],
		})
	}
	return bundles, nil
}

func parseSourceVersions(applicationVersionsStr string) ([]model.CreateVersionReference, error) {
	const (
		applicationKeyField = "application-key"
		versionField        = "version"
	)

	var refs []model.CreateVersionReference
	applicationVersionEntries := utils.ParseSliceFlag(applicationVersionsStr)
	for _, entry := range applicationVersionEntries {
		applicationVersionEntryMap, err := utils.ParseKeyValueString(entry, ",")
		if err != nil {
			return nil, errorutils.CheckErrorf("invalid application version format: %v", err)
		}
		err = validateRequiredFieldsInMap(applicationVersionEntryMap, applicationKeyField, versionField)
		if err != nil {
			return nil, errorutils.CheckErrorf("invalid application version format: %v", err)
		}
		refs = append(refs, model.CreateVersionReference{
			ApplicationKey: applicationVersionEntryMap[applicationKeyField],
			Version:        applicationVersionEntryMap[versionField],
		})
	}
	return refs, nil
}

func parsePackages(packagesStr string) ([]model.CreateVersionPackage, error) {
	const (
		typeField       = "type"
		nameField       = "name"
		versionField    = "version"
		repositoryField = "repo-key"
	)

	var packages []model.CreateVersionPackage
	packageEntries := utils.ParseSliceFlag(packagesStr)
	for _, entry := range packageEntries {
		packageEntryMap, err := utils.ParseKeyValueString(entry, ",")
		if err != nil {
			return nil, errorutils.CheckErrorf("invalid package format: %v", err)
		}
		err = validateRequiredFieldsInMap(packageEntryMap, typeField, nameField, versionField, repositoryField)
		if err != nil {
			return nil, errorutils.CheckErrorf("invalid package format: %v", err)
		}
		packages = append(packages, model.CreateVersionPackage{
			Type:       packageEntryMap[typeField],
			Name:       packageEntryMap[nameField],
			Version:    packageEntryMap[versionField],
			Repository: packageEntryMap[repositoryField],
		})
	}
	return packages, nil
}

func parseArtifacts(artifactsStr string) ([]model.CreateVersionArtifact, error) {
	const (
		pathField   = "path"
		sha256Field = "sha256"
	)

	var artifacts []model.CreateVersionArtifact
	artifactEntries := utils.ParseSliceFlag(artifactsStr)
	for _, entry := range artifactEntries {
		artifactEntryMap, err := utils.ParseKeyValueString(entry, ",")
		if err != nil {
			return nil, errorutils.CheckErrorf("invalid artifact format: %v", err)
		}
		err = validateRequiredFieldsInMap(artifactEntryMap, pathField)
		if err != nil {
			return nil, errorutils.CheckErrorf("invalid artifact format: %v", err)
		}
		artifact := model.CreateVersionArtifact{
			Path:   artifactEntryMap[pathField],
			SHA256: artifactEntryMap[sha256Field],
		}
		artifacts = append(artifacts, artifact)
	}
	return artifacts, nil
}

func buildFiltersFromFlags(ctx *components.Context) (*model.CreateVersionFilters, error) {
	includeFilterValues := ctx.GetStringsArrFlagValue(commands.IncludeFilterFlag)
	excludeFilterValues := ctx.GetStringsArrFlagValue(commands.ExcludeFilterFlag)

	if len(includeFilterValues) == 0 && len(excludeFilterValues) == 0 {
		return nil, nil
	}
	filters := &model.CreateVersionFilters{}
	if includedFilters, err := parseFilterValues(includeFilterValues); err != nil {
		return nil, err
	} else if len(includedFilters) > 0 {
		filters.Included = includedFilters
	}
	if excludedFilters, err := parseFilterValues(excludeFilterValues); err != nil {
		return nil, err
	} else if len(excludedFilters) > 0 {
		filters.Excluded = excludedFilters
	}

	return filters, nil
}

func parseFilterValues(filterValues []string) ([]*model.CreateVersionSourceFilter, error) {
	if len(filterValues) == 0 {
		return nil, nil
	}
	return parseFilters(filterValues)
}

func parseFilters(filterStrings []string) ([]*model.CreateVersionSourceFilter, error) {
	const (
		filterTypeField     = "filter_type"
		packageTypeField    = "type"
		packageNameField    = "name"
		packageVersionField = "version"
		artifactPathField   = "path"
		artifactShaField    = "sha256"
	)

	var filters []*model.CreateVersionSourceFilter

	for i, filterStr := range filterStrings {
		filterMap, err := utils.ParseKeyValueString(filterStr, ",")
		if err != nil {
			return nil, errorutils.CheckErrorf("invalid filter format at index %d: %v", i, err)
		}
		filterType, ok := filterMap[filterTypeField]
		if !ok {
			return nil, errorutils.CheckErrorf("invalid filter format at index %d: missing 'filter_type' field", i)
		}
		filter := &model.CreateVersionSourceFilter{}

		switch filterType {
		case "package":
			if val, ok := filterMap[packageTypeField]; ok {
				filter.PackageType = val
			}
			if val, ok := filterMap[packageNameField]; ok {
				filter.PackageName = val
			}
			if val, ok := filterMap[packageVersionField]; ok {
				filter.PackageVersion = val
			}
			if filter.PackageType == "" && filter.PackageName == "" && filter.PackageVersion == "" {
				return nil, errorutils.CheckErrorf("invalid package filter at index %d: at least one of 'type', 'name', or 'version' must be specified", i)
			}
		case "artifact":
			if val, ok := filterMap[artifactPathField]; ok {
				filter.Path = val
			}
			if val, ok := filterMap[artifactShaField]; ok {
				filter.SHA256 = val
			}
			if filter.Path == "" && filter.SHA256 == "" {
				return nil, errorutils.CheckErrorf("invalid artifact filter at index %d: at least one of 'path' or 'sha256' must be specified", i)
			}
		default:
			return nil, errorutils.CheckErrorf("invalid filter_type '%s' at index %d: must be 'package' or 'artifact'", filterType, i)
		}

		filters = append(filters, filter)
	}

	return filters, nil
}
