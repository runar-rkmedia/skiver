package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Masterminds/semver/v3"
	"github.com/runar-rkmedia/skiver/utils"
)

type CallingClient struct {
	Name    string
	Version string
	GitHash string
	Semver  *semver.Version
}

func GetClientVersionFromRequest(r *http.Request) (client CallingClient, found bool) {
	client.Name = r.Header.Get("Client_App")
	found = client.Name != ""
	client.GitHash = r.Header.Get("Client_Hash")
	client.Version = r.Header.Get("Client_version")
	if client.Version != "" {
		if v, err := semver.NewVersion(client.Version); err == nil {
			client.Semver = v
		}
	}

	return
}

var mincliversion = utils.Must(semver.NewConstraint(">= 0.6.0"))

// temporary check for clients before they upgrade to a version of
func ValidateClientVersion(rw http.ResponseWriter, r *http.Request) error {
	client, found := GetClientVersionFromRequest(r)
	if !found {
		// a client which is not reporting their versioning.
		// we have not control over them, and do not wish to interfere
		return nil
	}
	// This is temporary
	var err error
	var details []any
	if client.Version == "" {
		err = fmt.Errorf("client %s detected, missing version-number. Please upgrade.", client.Name)
	} else if client.Semver == nil {
		err = fmt.Errorf("client %s with version '%s' is outdated. Please upgrade", client.Name, client.Version)
	} else {
		switch client.Name {
		case "skiver-cli":
			if ok, errs := mincliversion.Validate(client.Semver); !ok {
				err = fmt.Errorf("client '%s' with version '%s' (%s) is outdated. Please upgrade", client.Name, client.Version, client.Semver.String())
				details = append(details, errs)
			}

		}
	}
	if err == nil {
		return nil
	}
	json, jerr := json.Marshal(NewApiErr(err, http.StatusConflict, "ErrOutdatedClient", details...))
	if jerr != nil {
		panic(err)
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusConflict)
	rw.Write([]byte(json))
	return err
}
