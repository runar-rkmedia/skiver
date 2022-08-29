package types

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type ServerInfo struct {
	// When the server was started
	ServerStartedAt time.Time `json:"server_started_at"`
	// Short githash for current commit
	GitHash string `json:"git_hash"`
	// Version-number for commit
	Version string `json:"version"`
	// Date of build
	BuildDate time.Time `json:"build_date"`

	// Server-instance. This will change on every restart.
	Instance string `json:"instance"`
	// Hash of the current host. Should be semi-stable
	HostHash string `json:"host_hash"`

	// Size of database.
	DatabaseSize     int64        `json:"database_size"`
	DatabaseSizeStr  string       `json:"database_size_str"`
	LatestRelease    *ReleaseInfo `json:"latest_release"`
	LatestReleaseCLI *ReleaseInfo `json:"latest_cli_release"`
	// The minimum version of skiver-cli that can be used with this server.
	// The is [semver](https://semver.org/)-compatible, but has a leading `v`, like `v1.2.3`
	MinCliVersion string `json:"min_cli_version"`
}

// Server info
// swagger:response
type serverInfo struct {
	// in:body
	Body []ServerInfo
}

func GetLatestVersion(url string, c *http.Client) (*ReleaseInfo, error) {
	// Rate-limited to 60 calls per hour without a token
	req, err := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}
		return nil, fmt.Errorf("Failed to get ReleaseInfo from url '%s' : %d %s %#v", url, res.StatusCode, string(b), res.Header)
	}
	var j ReleaseInfo
	err = json.NewDecoder(res.Body).Decode(&j)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}
	if j.TagName == "" {
		return nil, fmt.Errorf("No tagname returned in response")
	}
	return &j, err
}

// The response includes more information, but we don't care about all that.

type ReleaseInfo struct {
	URL             string `json:"url"`
	AssetsURL       string `json:"assets_url"`
	UploadURL       string `json:"upload_url"`
	HTMLURL         string `json:"html_url"`
	TagName         string `json:"tag_name"`
	TargetCommitish string `json:"target_commitish"`
	Name            string `json:"name"`
	Draft           bool   `json:"draft"`
	Prerelease      bool   `json:"prerelease"`
	CreatedAt       string `json:"created_at"`
	PublishedAt     string `json:"published_at"`
	Body            string `json:"body"`
}
