package utils

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

func ResolveAndStripSemver(v string) ([]string, error) {
	sv, err := semver.NewVersion(v)
	if err != nil {
		return nil, err
	}
	prerelease := sv.Prerelease()
	if prerelease != "" {
		return []string{sv.String()}, nil
	}
	patch := sv.Patch()
	minor := sv.Minor()
	major := sv.Major()

	slc := make([]string, 3)
	slc[0] = fmt.Sprintf("%d.%d.%d", major, minor, patch)
	slc[1] = fmt.Sprintf("%d.%d", major, minor)
	slc[2] = fmt.Sprintf("%d", major)

	return slc, nil

}

func unique(slice semver.Collection) semver.Collection {
	keys := make(map[string]bool)
	list := semver.Collection{}
	for _, entry := range slice {
		s := entry.String()
		if _, value := keys[s]; !value {
			keys[s] = true
			list = append(list, entry)
		}
	}
	return list
}
