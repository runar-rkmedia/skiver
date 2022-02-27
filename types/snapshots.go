package types

import (
	"github.com/mitchellh/hashstructure/v2"
)

// swagger:model ProjectSnapshot
type ProjectSnapshot struct {
	Entity      `json:"entity"`
	Project     ExtendedProject `json:"project"`
	ProjectHash uint64          `json:"project_hash"`
}

func (p ExtendedProject) CreateSnapshot(createdBy string) (s ProjectSnapshot, err error) {
	s.Project = p
	s.Project.Snapshots = nil
	pHash, err := hashstructure.Hash(p, hashstructure.FormatV2, nil)
	if err != nil {
		return s, err
	}
	s.ProjectHash = pHash

	s.Entity.CreatedBy = createdBy
	s.OrganizationID = p.OrganizationID

	return s, nil
}
func (e ProjectSnapshot) Namespace() string {
	return e.Kind()
}
func (e ProjectSnapshot) Kind() string {
	return string(PubTypeSnapshot)
}
