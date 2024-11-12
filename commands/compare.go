package commands

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/bwmarrin/discordgo"
)

func equalCommand(left, right *discordgo.ApplicationCommand) (bool, error) {
	switch true {
	case left.Type != right.Type:
		return false, fmt.Errorf("Type is not equal. Left: %#v Right: %#v", left.Type, right.Type)

	case left.Name != right.Name:
		return false, fmt.Errorf("Name is not equal. Left: %#v Right: %#v", left.Name, right.Name)

	case (left.NameLocalizations != nil || right.NameLocalizations != nil) &&
		!reflect.DeepEqual(left.NameLocalizations, right.NameLocalizations):
		return false, fmt.Errorf(
			"NameLocalizations is not equal. Left: %#v Right: %#v",
			left.NameLocalizations,
			right.NameLocalizations,
		)

	case (left.DefaultMemberPermissions != nil || right.DefaultMemberPermissions != nil) &&
		!reflect.DeepEqual(left.DefaultMemberPermissions, right.DefaultMemberPermissions):
		return false, fmt.Errorf(
			"DefaultMemberPermissions is not equal. Left: %#v Right: %#v",
			left.DefaultMemberPermissions,
			right.DefaultMemberPermissions,
		)

	case (left.DMPermission != nil || right.DMPermission != nil) &&
		!reflect.DeepEqual(left.DMPermission, right.DMPermission):
		return false, fmt.Errorf(
			"DMPermission is not equal. Left: %#v Right: %#v",
			left.DMPermission,
			right.DMPermission,
		)

	case (left.NSFW != nil || right.NSFW != nil) && !reflect.DeepEqual(left.NSFW, right.NSFW):
		return false, fmt.Errorf("VALUE is not equal. Left: %#v Right: %#v", left.NSFW, right.NSFW)

	case (left.Description != "" || right.Description != "") && left.Description != right.Description:
		return false, fmt.Errorf(
			"Description is not equal. Left: %#v Right: %#v",
			left.Description,
			right.Description,
		)

	case (left.DescriptionLocalizations != nil || right.DescriptionLocalizations != nil) &&
		!reflect.DeepEqual(left.DescriptionLocalizations, right.DescriptionLocalizations):
		return false, fmt.Errorf(
			"DescriptionLocalizations is not equal. Left: %#v Right: %#v",
			left.DescriptionLocalizations,
			right.DescriptionLocalizations,
		)

	case len(left.Options) != len(right.Options):
		return false, fmt.Errorf(
			"Options is not equal. Left: %#v Right: %#v",
			left.Options,
			right.Options,
		)

	case len(left.Options) > 0 && len(right.Options) > 0:
		for i, o := range left.Options {
			if ok, err := equalCommandOption(o, right.Options[i]); !ok {
				return ok, errors.Join(fmt.Errorf("Option element of index %v has difference", err))
			}
		}
	}

	return true, nil
}

func equalCommandOption(left, right *discordgo.ApplicationCommandOption) (bool, error) {
	switch true {
	case left.Type != right.Type:
		return false, fmt.Errorf("Type is not equal. Left: %#v Right: %#v", left.Type, right.Type)

	case left.Name != right.Name:
		return false, fmt.Errorf("Name is not equal. Left: %#v Right: %#v", left.Name, right.Name)

	case left.Description != right.Description:
		return false, fmt.Errorf(
			"Description is not equal. Left: %#v Right: %#v",
			left.Description,
			right.Description,
		)

	case (left.DescriptionLocalizations != nil || right.DescriptionLocalizations != nil) &&
		!reflect.DeepEqual(left.DescriptionLocalizations, right.DescriptionLocalizations):
		return false, fmt.Errorf(
			"DescriptionLocalizations is not equal. Left: %#v Right: %#v",
			left.DescriptionLocalizations,
			right.DescriptionLocalizations,
		)

	case !reflect.DeepEqual(left.ChannelTypes, right.ChannelTypes):
		return false, fmt.Errorf(
			"ChannelTypes is not equal. Left: %#v Right: %#v",
			left.ChannelTypes,
			right.ChannelTypes,
		)

	case left.Required != right.Required:
		return false, fmt.Errorf(
			"Required is not equal. Left: %#v Right: %#v",
			left.Required,
			right.Required,
		)

	case !reflect.DeepEqual(left.Choices, right.Choices):
		return false, fmt.Errorf(
			"Choices is not equal. Left: %#v Right: %#v",
			left.Choices,
			right.Choices,
		)

	case (left.MinValue != nil || right.MinValue != nil) &&
		!reflect.DeepEqual(left.MinValue, right.MinValue):
		return false, fmt.Errorf(
			"MinValue is not equal. Left: %#v Right: %#v",
			left.MinValue,
			right.MinValue,
		)

	case (left.MaxValue != 0 || right.MaxValue != 0) && left.MaxValue != right.MaxValue:
		return false, fmt.Errorf(
			"MaxValue is not equal. Left: %#v Right: %#v",
			left.MaxValue,
			right.MaxValue,
		)

	case (left.MinLength != nil || right.MinLength != nil) &&
		!reflect.DeepEqual(left.MinLength, right.MinLength):
		return false, fmt.Errorf(
			"MinLength is not equal. Left: %#v Right: %#v",
			left.MinLength,
			right.MinLength,
		)

	case (left.MaxLength != 0 || right.MaxLength != 0) && left.MaxLength != right.MaxLength:
		return false, fmt.Errorf(
			"MaxLength is not equal. Left: %#v Right: %#v",
			left.MaxLength,
			right.MaxLength,
		)
	}

	return true, nil
}
