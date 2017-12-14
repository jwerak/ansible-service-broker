package validation

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/validation/path"
	"k8s.io/apimachinery/pkg/util/validation/field"
	kvalidation "k8s.io/kubernetes/pkg/api/validation"

	userapi "github.com/openshift/ansible-service-broker/pkg/origin/copy/user"
)

// ValidateUserName -
func ValidateUserName(name string, _ bool) []string {
	if reasons := path.ValidatePathSegmentName(name, false); len(reasons) != 0 {
		return reasons
	}

	if strings.Contains(name, ":") {
		return []string{`may not contain ":"`}
	}
	if name == "~" {
		return []string{`may not equal "~"`}
	}
	return nil
}

// ValidateIdentityName -
func ValidateIdentityName(name string, _ bool) []string {
	if reasons := path.ValidatePathSegmentName(name, false); len(reasons) != 0 {
		return reasons
	}

	parts := strings.Split(name, ":")
	if len(parts) != 2 {
		return []string{`must be in the format <providerName>:<providerUserName>`}
	}
	if len(parts[0]) == 0 {
		return []string{`must be in the format <providerName>:<providerUserName> with a non-empty providerName`}
	}
	if len(parts[1]) == 0 {
		return []string{`must be in the format <providerName>:<providerUserName> with a non-empty providerUserName`}
	}
	return nil
}

// ValidateGroupName -
func ValidateGroupName(name string, _ bool) []string {
	if reasons := path.ValidatePathSegmentName(name, false); len(reasons) != 0 {
		return reasons
	}

	if strings.Contains(name, ":") {
		return []string{`may not contain ":"`}
	}
	if name == "~" {
		return []string{`may not equal "~"`}
	}
	return nil
}

// ValidateIdentityProviderName -
func ValidateIdentityProviderName(name string) []string {
	if reasons := path.ValidatePathSegmentName(name, false); len(reasons) != 0 {
		return reasons
	}

	if strings.Contains(name, ":") {
		return []string{`may not contain ":"`}
	}
	return nil
}

// ValidateIdentityProviderUserName -
func ValidateIdentityProviderUserName(name string) []string {
	// Any provider user name must be a valid user name
	return ValidateUserName(name, false)
}

// ValidateGroup -
func ValidateGroup(group *userapi.Group) field.ErrorList {
	allErrs := kvalidation.ValidateObjectMeta(&group.ObjectMeta, false, ValidateGroupName, field.NewPath("metadata"))

	userPath := field.NewPath("user")
	for index, user := range group.Users {
		idxPath := userPath.Index(index)
		if len(user) == 0 {
			allErrs = append(allErrs, field.Invalid(idxPath, user, "may not be empty"))
			continue
		}
		if reasons := ValidateUserName(user, false); len(reasons) != 0 {
			allErrs = append(allErrs, field.Invalid(idxPath, user, strings.Join(reasons, ", ")))
		}
	}

	return allErrs
}

// ValidateGroupUpdate -
func ValidateGroupUpdate(group *userapi.Group, old *userapi.Group) field.ErrorList {
	allErrs := kvalidation.ValidateObjectMetaUpdate(&group.ObjectMeta, &old.ObjectMeta, field.NewPath("metadata"))
	allErrs = append(allErrs, ValidateGroup(group)...)
	return allErrs
}

// ValidateUser -
func ValidateUser(user *userapi.User) field.ErrorList {
	allErrs := kvalidation.ValidateObjectMeta(&user.ObjectMeta, false, ValidateUserName, field.NewPath("metadata"))
	identitiesPath := field.NewPath("identities")
	for index, identity := range user.Identities {
		idxPath := identitiesPath.Index(index)
		if reasons := ValidateIdentityName(identity, false); len(reasons) != 0 {
			allErrs = append(allErrs, field.Invalid(idxPath, identity, strings.Join(reasons, ", ")))
		}
	}

	groupsPath := field.NewPath("groups")
	for index, group := range user.Groups {
		idxPath := groupsPath.Index(index)
		if len(group) == 0 {
			allErrs = append(allErrs, field.Invalid(idxPath, group, "may not be empty"))
			continue
		}
		if reasons := ValidateGroupName(group, false); len(reasons) != 0 {
			allErrs = append(allErrs, field.Invalid(idxPath, group, strings.Join(reasons, ", ")))
		}
	}

	return allErrs
}

// ValidateUserUpdate -
func ValidateUserUpdate(user *userapi.User, old *userapi.User) field.ErrorList {
	allErrs := kvalidation.ValidateObjectMetaUpdate(&user.ObjectMeta, &old.ObjectMeta, field.NewPath("metadata"))
	allErrs = append(allErrs, ValidateUser(user)...)
	return allErrs
}

// ValidateIdentity -
func ValidateIdentity(identity *userapi.Identity) field.ErrorList {
	allErrs := kvalidation.ValidateObjectMeta(&identity.ObjectMeta, false, ValidateIdentityName, field.NewPath("metadata"))

	if len(identity.ProviderName) == 0 {
		allErrs = append(allErrs, field.Required(field.NewPath("providerName"), ""))
	} else if reasons := ValidateIdentityProviderName(identity.ProviderName); len(reasons) != 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("providerName"), identity.ProviderName, strings.Join(reasons, ", ")))
	}

	if len(identity.ProviderUserName) == 0 {
		allErrs = append(allErrs, field.Required(field.NewPath("providerUserName"), ""))
	} else if reasons := ValidateIdentityProviderUserName(identity.ProviderUserName); len(reasons) != 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("providerUserName"), identity.ProviderUserName, strings.Join(reasons, ", ")))
	}

	userPath := field.NewPath("user")
	if len(identity.ProviderName) > 0 && len(identity.ProviderUserName) > 0 {
		expectedIdentityName := identity.ProviderName + ":" + identity.ProviderUserName
		if identity.Name != expectedIdentityName {
			allErrs = append(allErrs, field.Invalid(userPath.Child("name"), identity.User.Name, fmt.Sprintf("must be %s", expectedIdentityName)))
		}
	}

	if reasons := ValidateUserName(identity.User.Name, false); len(reasons) != 0 {
		allErrs = append(allErrs, field.Invalid(userPath.Child("name"), identity.User.Name, strings.Join(reasons, ", ")))
	}
	if len(identity.User.Name) == 0 && len(identity.User.UID) != 0 {
		allErrs = append(allErrs, field.Required(userPath.Child("username"), "username is required when uid is provided"))
	}
	if len(identity.User.Name) != 0 && len(identity.User.UID) == 0 {
		allErrs = append(allErrs, field.Required(userPath.Child("uid"), "uid is required when username is provided"))
	}
	return allErrs
}

// ValidateIdentityUpdate -
func ValidateIdentityUpdate(identity *userapi.Identity, old *userapi.Identity) field.ErrorList {
	allErrs := kvalidation.ValidateObjectMetaUpdate(&identity.ObjectMeta, &old.ObjectMeta, field.NewPath("metadata"))
	allErrs = append(allErrs, ValidateIdentity(identity)...)

	if identity.ProviderName != old.ProviderName {
		allErrs = append(allErrs, field.Invalid(field.NewPath("providerName"), identity.ProviderName, "may not change providerName"))
	}
	if identity.ProviderUserName != old.ProviderUserName {
		allErrs = append(allErrs, field.Invalid(field.NewPath("providerUserName"), identity.ProviderUserName, "may not change providerUserName"))
	}

	return allErrs
}

// ValidateIdentityMapping -
func ValidateIdentityMapping(mapping *userapi.IdentityMapping) field.ErrorList {
	allErrs := kvalidation.ValidateObjectMeta(&mapping.ObjectMeta, false, ValidateIdentityName, field.NewPath("metadata"))

	identityPath := field.NewPath("identity")
	if len(mapping.Identity.Name) == 0 {
		allErrs = append(allErrs, field.Required(identityPath.Child("name"), ""))
	}
	if mapping.Identity.Name != mapping.Name {
		allErrs = append(allErrs, field.Invalid(identityPath.Child("name"), mapping.Identity.Name, "must match metadata.name"))
	}
	if len(mapping.User.Name) == 0 {
		allErrs = append(allErrs, field.Required(field.NewPath("user", "name"), ""))
	}
	return allErrs
}

// ValidateIdentityMappingUpdate -
func ValidateIdentityMappingUpdate(mapping *userapi.IdentityMapping, old *userapi.IdentityMapping) field.ErrorList {
	allErrs := kvalidation.ValidateObjectMetaUpdate(&mapping.ObjectMeta, &old.ObjectMeta, field.NewPath("metadata"))
	allErrs = append(allErrs, ValidateIdentityMapping(mapping)...)
	return allErrs
}
