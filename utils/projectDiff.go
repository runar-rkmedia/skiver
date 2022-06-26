package utils

import (
	"encoding/json"

	"github.com/r3labs/diff/v2"
	"github.com/runar-rkmedia/skiver/models"
)

func NewProjectDiff(a, b any, input models.DiffSnapshotInput) (*ProjectDiffResponse, error) {
	aJson, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	bJson, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}
	A := NewBinaryMetaFromBytes(aJson)
	B := NewBinaryMetaFromBytes(bJson)

	pd := &ProjectDiffResponse{
		A: ProjectStats{BinaryMeta: A},
		B: ProjectStats{BinaryMeta: B},
	}
	if !A.Equal(B) {
		c, err := diff.Diff(a, b, diff.DisableStructValues(), diff.AllowTypeMismatch(true))
		if err != nil {
			return nil, err
		}
		pd.Diff = c
	}
	if input.A != nil {
		pd.A.Tag = input.A.Tag
		if input.A.ProjectID != nil {
			pd.A.ProjectID = *input.A.ProjectID
		}
	}
	if input.B != nil {
		pd.B.Tag = input.B.Tag
		if input.B.ProjectID != nil {
			pd.B.ProjectID = *input.B.ProjectID
		}
	}
	return pd, nil
}

type ProjectStats struct {
	BinaryMeta
	ProjectID string `json:"project_id"`
	Tag       string `json:"tag"`
}
type ProjectDiffResponse struct {
	Diff diff.Changelog `json:"diff"`
	A    ProjectStats   `json:"a"`
	B    ProjectStats   `json:"b"`
}
