package commands

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/bwmarrin/discordgo"
)

// Helper function used inside equalToRegistered and equalToRegistered option,
// THIS ISN'T SUPPOSED TO BE USED outside of said functions.
func equal[T any](l, r T) bool {
	lv := reflect.ValueOf(l)

	if lv.Kind() == reflect.Pointer {
		// If the local value is a nil-pointer, we assume it as the
		// zero-value of the underling value for comparison
		var v any
		if lv.IsNil() {
			v = reflect.Zero(lv.Type().Elem()).Interface()
		} else {
			v = lv.Elem().Interface()
		}

		rv := reflect.ValueOf(r)

		var rvv any
		if rv.IsNil() {
			rvv = reflect.Zero(rv.Type().Elem()).Interface()
		} else {
			rvv = rv.Elem().Interface()
		}

		return equal(v, rvv)
	} else {
		return reflect.DeepEqual(l, r)
	}
}

// Helper function used inside equalToRegistered and equalToRegistered option,
// THIS ISN'T SUPPOSED TO BE USED outside of said functions.
func val[T any](v *T) T {
	if v != nil {
		return *v
	} else {
		return *new(T)
	}
}

func equalToRegistered(local, registered *discordgo.ApplicationCommand) (bool, error) {
	switch {
	case local.Type != registered.Type:
		return false, fmt.Errorf(
			"Type is not equal. Local: %#v Registered: %#v",
			local.Type,
			registered.Type,
		)

	case local.Name != registered.Name:
		return false, fmt.Errorf(
			"Name is not equal. Local: %#v Registered: %#v",
			local.Name,
			registered.Name,
		)

	case !equal(local.NameLocalizations, registered.NameLocalizations):
		return false, fmt.Errorf(
			"NameLocalizations is not equal. Local: *%#v Registered: *%#v",
			val(local.NameLocalizations),
			val(registered.NameLocalizations),
		)

	// DEPRECATED FIELDS https://discord.com/developers/docs/interactions/application-commands
	//
	// case !equal(local.DefaultMemberPermissions, registered.DefaultMemberPermissions):
	// 	return false, fmt.Errorf(
	// 		"DefaultMemberPermissions is not equal. Local: *%#v Registered: *%#v",
	// 		val(local.DefaultMemberPermissions),
	// 		val(registered.DefaultMemberPermissions),
	// 	)
	//
	// case
	// 	!equal(local.DMPermission, registered.DMPermission):
	// 	return false, fmt.Errorf(
	// 		"DMPermission is not equal. Local: *%#v Registered: *%#v",
	// 		val(local.DMPermission),
	// 		val(registered.DMPermission),
	// 	)

	case !equal(local.NSFW, registered.NSFW):
		return false, fmt.Errorf(
			"VALUE is not equal. Local: *%#v Registered: *%#v",
			val(local.NSFW),
			val(registered.NSFW),
		)

	case local.Description != registered.Description:
		return false, fmt.Errorf(
			"Description is not equal. Local: %#v Registered: %#v",
			local.Description,
			registered.Description,
		)

	case !equal(local.DescriptionLocalizations, registered.DescriptionLocalizations):
		return false, fmt.Errorf(
			"DescriptionLocalizations is not equal. Local: *%#v Registered: *%#v",
			val(local.DescriptionLocalizations),
			val(registered.DescriptionLocalizations),
		)

	case len(local.Options) != len(registered.Options):
		return false, fmt.Errorf(
			"Options is not equal. Local: %#v Registered: %#v",
			local.Options,
			registered.Options,
		)

	case len(local.Options) > 0 && len(registered.Options) > 0:
		for i, o := range local.Options {
			if ok, err := equalToRegisteredOption(o, registered.Options[i]); !ok {
				return ok, errors.Join(fmt.Errorf("Option element of index %v has difference", err))
			}
		}
	}

	return true, nil
}

func equalToRegisteredOption(local, registered *discordgo.ApplicationCommandOption) (bool, error) {
	switch {
	case local.Type != registered.Type:
		return false, fmt.Errorf(
			"Type is not equal. Local: %#v Registered: %#v",
			local.Type,
			registered.Type,
		)

	case local.Name != registered.Name:
		return false, fmt.Errorf(
			"Name is not equal. Local: %#v Registered: %#v",
			local.Name,
			registered.Name,
		)

	case local.Description != registered.Description:
		return false, fmt.Errorf(
			"Description is not equal. Local: %#v Registered: %#v",
			local.Description,
			registered.Description,
		)

	case !equal(local.DescriptionLocalizations, registered.DescriptionLocalizations):
		return false, fmt.Errorf(
			"DescriptionLocalizations is not equal. Local: %#v Registered: %#v",
			local.DescriptionLocalizations,
			registered.DescriptionLocalizations,
		)

	case !equal(local.ChannelTypes, registered.ChannelTypes):
		return false, fmt.Errorf(
			"ChannelTypes is not equal. Local: %#v Registered: %#v",
			local.ChannelTypes,
			registered.ChannelTypes,
		)

	case local.Required != registered.Required:
		return false, fmt.Errorf(
			"Required is not equal. Local: %#v Registered: %#v",
			local.Required,
			registered.Required,
		)

	case !equal(local.Choices, registered.Choices):
		return false, fmt.Errorf(
			"Choices is not equal. Local: %#v Registered: %#v",
			local.Choices,
			registered.Choices,
		)

	case !equal(local.MinValue, registered.MinValue):
		return false, fmt.Errorf(
			"MinValue is not equal. Local: *%#v Registered: *%#v",
			val(local.MinValue),
			val(registered.MinValue),
		)

	case local.MaxValue != registered.MaxValue:
		return false, fmt.Errorf(
			"MaxValue is not equal. Local: %#v Registered: %#v",
			local.MaxValue,
			registered.MaxValue,
		)

	case !equal(local.MinLength, registered.MinLength):
		return false, fmt.Errorf(
			"MinLength is not equal. Local: *%#v Registered: *%#v",
			val(local.MinLength),
			val(registered.MinLength),
		)

	case local.MaxLength != registered.MaxLength:
		return false, fmt.Errorf(
			"MaxLength is not equal. Local: %#v Registered: %#v",
			local.MaxLength,
			registered.MaxLength,
		)
	}

	return true, nil
}
