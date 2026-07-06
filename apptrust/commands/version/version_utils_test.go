package version

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOverwriteStrategy(t *testing.T) {
	tests := []struct {
		name          string
		flagValue     string
		expectError   bool
		expectedValue string
	}{
		{
			name:          "valid value - disabled",
			flagValue:     "disabled",
			expectError:   false,
			expectedValue: "DISABLED",
		},
		{
			name:          "valid value - latest",
			flagValue:     "latest",
			expectError:   false,
			expectedValue: "LATEST",
		},
		{
			name:          "valid value - all",
			flagValue:     "all",
			expectError:   false,
			expectedValue: "ALL",
		},
		{
			name:          "empty value",
			flagValue:     "",
			expectError:   false,
			expectedValue: "",
		},
		{
			name:        "invalid value",
			flagValue:   "INVALID",
			expectError: true,
		},
		{
			name:        "uppercase value",
			flagValue:   "DISABLED",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &components.Context{}
			if tt.flagValue != "" {
				ctx.AddStringFlag(commands.OverwriteStrategyFlag, tt.flagValue)
			}

			result, err := ParseOverwriteStrategy(ctx)

			if tt.expectError {
				assert.Error(t, err, "ParseOverwriteStrategy(%q) expected error, got nil", tt.flagValue)
				return
			}

			assert.NoError(t, err, "ParseOverwriteStrategy(%q) unexpected error: %v", tt.flagValue, err)
			assert.Equal(t, tt.expectedValue, result, "ParseOverwriteStrategy(%q) = %v, want %v", tt.flagValue, result, tt.expectedValue)
		})
	}
}

func TestParseConflictResolution(t *testing.T) {
	tests := []struct {
		name          string
		flagValue     string
		expectError   bool
		expectedValue string
	}{
		{
			name:          "valid value - automatic",
			flagValue:     "automatic",
			expectedValue: "automatic",
		},
		{
			name:          "valid value - manual",
			flagValue:     "manual",
			expectedValue: "manual",
		},
		{
			name:          "empty value (omitted)",
			flagValue:     "",
			expectedValue: "",
		},
		{
			name:        "invalid value",
			flagValue:   "invalid",
			expectError: true,
		},
		{
			name:        "uppercase value",
			flagValue:   "AUTOMATIC",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &components.Context{}
			if tt.flagValue != "" {
				ctx.AddStringFlag(commands.ConflictResolutionFlag, tt.flagValue)
			}

			result, err := ParseConflictResolution(ctx)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedValue, result)
		})
	}
}

func TestBuildPromotionParams(t *testing.T) {
	tests := []struct {
		name                  string
		promotionType         string
		dryRun                bool
		includeRepos          string
		excludeRepos          string
		expectedPromotionType string
		expectedIncludeRepos  []string
		expectedExcludeRepos  []string
		expectError           bool
	}{
		{
			name:                  "default promotion type (copy)",
			promotionType:         "",
			dryRun:                false,
			expectedPromotionType: model.PromotionTypeCopy,
			expectedIncludeRepos:  []string(nil),
			expectedExcludeRepos:  []string(nil),
			expectError:           false,
		},
		{
			name:                  "promotion type move",
			promotionType:         "move",
			dryRun:                false,
			expectedPromotionType: model.PromotionTypeMove,
			expectedIncludeRepos:  []string(nil),
			expectedExcludeRepos:  []string(nil),
			expectError:           false,
		},
		{
			name:                  "dry run overrides promotion type",
			promotionType:         "copy",
			dryRun:                true,
			expectedPromotionType: model.PromotionTypeDryRun,
			expectedIncludeRepos:  []string(nil),
			expectedExcludeRepos:  []string(nil),
			expectError:           false,
		},
		{
			name:                  "with include and exclude repos",
			promotionType:         "copy",
			dryRun:                false,
			includeRepos:          "repo1;repo2",
			excludeRepos:          "repo3;repo4",
			expectedPromotionType: model.PromotionTypeCopy,
			expectedIncludeRepos:  []string{"repo1", "repo2"},
			expectedExcludeRepos:  []string{"repo3", "repo4"},
			expectError:           false,
		},
		{
			name:          "invalid promotion type",
			promotionType: "invalid",
			dryRun:        false,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &components.Context{}

			// Set flag values using AddStringFlag
			if tt.promotionType != "" {
				ctx.AddStringFlag(commands.PromotionTypeFlag, tt.promotionType)
			}
			if tt.dryRun {
				ctx.AddBoolFlag(commands.DryRunFlag, tt.dryRun)
			}
			if tt.includeRepos != "" {
				ctx.AddStringFlag(commands.IncludeReposFlag, tt.includeRepos)
			}
			if tt.excludeRepos != "" {
				ctx.AddStringFlag(commands.ExcludeReposFlag, tt.excludeRepos)
			}

			promotionType, includeRepos, excludeRepos, err := BuildPromotionParams(ctx)

			if tt.expectError {
				assert.Error(t, err, "BuildPromotionParams expected error, got nil")
				return
			}

			assert.NoError(t, err, "BuildPromotionParams unexpected error: %v", err)
			assert.Equal(t, tt.expectedPromotionType, promotionType, "promotion type mismatch")
			assert.Equal(t, tt.expectedIncludeRepos, includeRepos, "include repos mismatch")
			assert.Equal(t, tt.expectedExcludeRepos, excludeRepos, "exclude repos mismatch")
		})
	}
}

func TestParsePathMappings(t *testing.T) {
	tests := []struct {
		name        string
		flagValue   string
		expected    *model.PromotionModifications
		expectError bool
		errContains string
	}{
		{
			name:     "no flag - returns nil",
			expected: nil,
		},
		{
			name:      "single mapping without package type",
			flagValue: "input=(.*), output=stable-release/$1",
			expected: &model.PromotionModifications{
				Mappings: []model.PromotionPathMapping{
					{Input: "(.*)", Output: "stable-release/$1"},
				},
			},
		},
		{
			name:      "single mapping with package type",
			flagValue: "input=(.*), output=stable-release/$1, package-type=.*",
			expected: &model.PromotionModifications{
				Mappings: []model.PromotionPathMapping{
					{PackageType: ".*", Input: "(.*)", Output: "stable-release/$1"},
				},
			},
		},
		{
			name:      "multiple mappings",
			flagValue: "input=(.*), output=release/$1, package-type=.*; input=(.*\\.jar), output=jars/$1, package-type=maven",
			expected: &model.PromotionModifications{
				Mappings: []model.PromotionPathMapping{
					{PackageType: ".*", Input: "(.*)", Output: "release/$1"},
					{PackageType: "maven", Input: "(.*\\.jar)", Output: "jars/$1"},
				},
			},
		},
		{
			name:      "mapping without package-type field",
			flagValue: "input=(.*), output=release/$1; input=(.*\\.jar), output=jars/$1",
			expected: &model.PromotionModifications{
				Mappings: []model.PromotionPathMapping{
					{Input: "(.*)", Output: "release/$1"},
					{Input: "(.*\\.jar)", Output: "jars/$1"},
				},
			},
		},
		{
			name:        "missing input field - error",
			flagValue:   "output=target/$1",
			expectError: true,
			errContains: "'input' is required",
		},
		{
			name:        "missing output field - error",
			flagValue:   "input=(.*)",
			expectError: true,
			errContains: "'output' is required",
		},
		{
			name:        "empty entry from trailing semicolon - error",
			flagValue:   "input=(.*), output=release/$1;",
			expectError: true,
			errContains: "entry 2 is empty",
		},
		{
			name:        "invalid key-value format - error",
			flagValue:   "not-a-valid-format",
			expectError: true,
			errContains: "entry 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &components.Context{}
			if tt.flagValue != "" {
				ctx.AddStringFlag(commands.PathMappingFlag, tt.flagValue)
			}

			result, err := ParsePathMappings(ctx)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateDistributionFlags(t *testing.T) {
	tests := []struct {
		name        string
		distRules   bool
		site        bool
		city        bool
		country     bool
		expectError bool
	}{
		{
			name: "no flags",
		},
		{
			name:      "dist-rules only",
			distRules: true,
		},
		{
			name:    "site/city/country only",
			site:    true,
			city:    true,
			country: true,
		},
		{
			name:        "dist-rules with site",
			distRules:   true,
			site:        true,
			expectError: true,
		},
		{
			name:        "dist-rules with city",
			distRules:   true,
			city:        true,
			expectError: true,
		},
		{
			name:        "dist-rules with country-codes",
			distRules:   true,
			country:     true,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &components.Context{}
			if tt.distRules {
				ctx.AddStringFlag(commands.DistRulesFlag, "rules.json")
			}
			if tt.site {
				ctx.AddStringFlag(commands.SiteFlag, "edge-*")
			}
			if tt.city {
				ctx.AddStringFlag(commands.CityFlag, "NYC")
			}
			if tt.country {
				ctx.AddStringFlag(commands.CountryCodesFlag, "US")
			}

			err := ValidateDistributionFlags(ctx)

			if tt.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestParseDistributionRules(t *testing.T) {
	t.Run("from site/city/country flags", func(t *testing.T) {
		ctx := &components.Context{}
		ctx.AddStringFlag(commands.SiteFlag, "edge-*")
		ctx.AddStringFlag(commands.CityFlag, "NYC")
		ctx.AddStringFlag(commands.CountryCodesFlag, "US;CA")

		rules, err := ParseDistributionRules(ctx)
		require.NoError(t, err)
		expected := []model.DistributionRule{
			{SiteName: "edge-*", CityName: "NYC", CountryCodes: []string{"US", "CA"}},
		}
		assert.Equal(t, expected, rules)
	})

	t.Run("no flags defaults to distributing to all targets", func(t *testing.T) {
		ctx := &components.Context{}

		rules, err := ParseDistributionRules(ctx)
		require.NoError(t, err)
		expected := []model.DistributionRule{
			{SiteName: "*"},
		}
		assert.Equal(t, expected, rules)
	})

	t.Run("from dist-rules file", func(t *testing.T) {
		content := `{"distribution_rules":[{"site_name":"site-1","city_name":"city-1","country_codes":["US"]},{"site_name":"site-2"}]}`
		filePath := filepath.Join(t.TempDir(), "dist-rules.json")
		require.NoError(t, os.WriteFile(filePath, []byte(content), 0o600))

		ctx := &components.Context{}
		ctx.AddStringFlag(commands.DistRulesFlag, filePath)

		rules, err := ParseDistributionRules(ctx)
		require.NoError(t, err)
		expected := []model.DistributionRule{
			{SiteName: "site-1", CityName: "city-1", CountryCodes: []string{"US"}},
			{SiteName: "site-2"},
		}
		assert.Equal(t, expected, rules)
	})

	t.Run("empty dist-rules file returns no rules", func(t *testing.T) {
		content := `{"distribution_rules":[]}`
		filePath := filepath.Join(t.TempDir(), "dist-rules.json")
		require.NoError(t, os.WriteFile(filePath, []byte(content), 0o600))

		ctx := &components.Context{}
		ctx.AddStringFlag(commands.DistRulesFlag, filePath)

		rules, err := ParseDistributionRules(ctx)
		require.NoError(t, err)
		assert.Empty(t, rules)
	})

	t.Run("missing dist-rules file returns error", func(t *testing.T) {
		ctx := &components.Context{}
		ctx.AddStringFlag(commands.DistRulesFlag, filepath.Join(t.TempDir(), "does-not-exist.json"))

		_, err := ParseDistributionRules(ctx)
		assert.Error(t, err)
	})
}

func TestParseDistributionModifications(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		target      string
		expected    []model.DistributionPathMapping
		expectError bool
	}{
		{
			name:     "no mapping flags",
			expected: nil,
		},
		{
			name:    "pattern and target provided",
			pattern: "my-repo/(*)",
			target:  "edge/{1}",
			expected: []model.DistributionPathMapping{
				{Input: "^my-repo/(.*)$", Output: "edge/$1"},
			},
		},
		{
			name:        "only pattern provided",
			pattern:     "my-repo/(*)",
			expectError: true,
		},
		{
			name:        "only target provided",
			target:      "edge/{1}",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &components.Context{}
			if tt.pattern != "" {
				ctx.AddStringFlag(commands.MappingPatternFlag, tt.pattern)
			}
			if tt.target != "" {
				ctx.AddStringFlag(commands.MappingTargetFlag, tt.target)
			}

			result, err := ParseDistributionModifications(ctx)

			if tt.expectError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}
			require.NotNil(t, result)
			assert.Equal(t, tt.expected, result.PathMappings)
		})
	}
}
