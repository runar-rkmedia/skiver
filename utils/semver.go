package utils

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

// Returns an array of strings that resolves to the given version
func ResolveSemver(v string) (semver.Collection, error) {
	var resolved semver.Collection
	sv, err := semver.NewVersion(v)
	if err != nil {
		return resolved, err
	}
	prerelease := sv.Prerelease()
	patch := sv.Patch()
	minor := sv.Minor()
	major := sv.Major()

	if prerelease != "" {
		resolved = append(resolved, sv)
		// On prereleases, we dont return the rest of the versions.
		// since they should not resolve to this version
		return resolved, nil
	}
	if patch != 0 {
		s := fmt.Sprintf("v%d.%d.%d", major, minor, patch)
		resolved = append(resolved, semver.MustParse(s))
	}
	if minor != 0 {
		s := fmt.Sprintf("v%d.%d", major, minor)
		resolved = append(resolved, semver.MustParse(s))
	}
	if major != 0 {
		s := fmt.Sprintf("v%d", major)
		resolved = append(resolved, semver.MustParse(s))
	}

	return resolved, err
}
