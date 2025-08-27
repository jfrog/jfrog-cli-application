package application

import (
	"github.com/jfrog/jfrog-cli-application/apptrust/commands"
	"github.com/jfrog/jfrog-cli-application/apptrust/commands/utils"
	"github.com/jfrog/jfrog-cli-application/apptrust/model"
	"github.com/jfrog/jfrog-cli-core/v2/plugins/components"
)

func populateApplicationFromFlags(ctx *components.Context, descriptor *model.AppDescriptor) error {
	descriptor.ApplicationName = ctx.GetStringFlagValue(commands.ApplicationNameFlag)

	if ctx.IsFlagSet(commands.DescriptionFlag) {
		description := ctx.GetStringFlagValue(commands.DescriptionFlag)
		descriptor.Description = &description
	}

	if ctx.IsFlagSet(commands.BusinessCriticalityFlag) {
		businessCriticalityStr := ctx.GetStringFlagValue(commands.BusinessCriticalityFlag)
		businessCriticality, err := utils.ValidateEnumFlag(
			commands.BusinessCriticalityFlag,
			businessCriticalityStr,
			model.BusinessCriticalityUnspecified,
			model.BusinessCriticalityValues)
		if err != nil {
			return err
		}
		descriptor.BusinessCriticality = &businessCriticality
	}

	if ctx.IsFlagSet(commands.MaturityLevelFlag) {
		maturityLevelStr := ctx.GetStringFlagValue(commands.MaturityLevelFlag)
		maturityLevel, err := utils.ValidateEnumFlag(
			commands.MaturityLevelFlag,
			maturityLevelStr,
			model.MaturityLevelUnspecified,
			model.MaturityLevelValues)
		if err != nil {
			return err
		}
		descriptor.MaturityLevel = &maturityLevel
	}

	if ctx.IsFlagSet(commands.LabelsFlag) {
		labelsMap, err := utils.ParseMapFlag(ctx.GetStringFlagValue(commands.LabelsFlag))
		if err != nil {
			return err
		}
		descriptor.Labels = &labelsMap
	}

	if ctx.IsFlagSet(commands.UserOwnersFlag) {
		userOwners := utils.ParseSliceFlag(ctx.GetStringFlagValue(commands.UserOwnersFlag))
		descriptor.UserOwners = &userOwners
	}

	if ctx.IsFlagSet(commands.GroupOwnersFlag) {
		groupOwners := utils.ParseSliceFlag(ctx.GetStringFlagValue(commands.GroupOwnersFlag))
		descriptor.GroupOwners = &groupOwners
	}

	return nil
}
