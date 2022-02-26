package bboltStorage

import (
	"fmt"

	"github.com/runar-rkmedia/skiver/types"
)

func (bb *BBolter) GetSnapshot(snapshotId string) (*types.ProjectSnapshot, error) {
	return Get[types.ProjectSnapshot](bb, BucketSnapshot, snapshotId)
}
func (bb *BBolter) FindSnapshots(max int, filter ...types.ProjectSnapshot) (map[string]types.ProjectSnapshot, error) {
	return Find(bb, BucketSnapshot, max, func(uu types.ProjectSnapshot) bool {
		if len(filter) == 0 {
			return true
		}
		for _, f := range filter {
			if snapshotFilter(f, uu) {
				return true
			}
		}
		return false
	})
}

func snapshotFilter(f, s types.ProjectSnapshot) bool {
	if f.OrganizationID != "" && f.OrganizationID != s.OrganizationID {
		return false
	}
	if f.ID != "" && f.ID != s.ID {
		return false
	}
	if f.ProjectHash != 0 && f.ProjectHash != s.ProjectHash {
		return false
	}
	return true
}

func (b *BBolter) CreateSnapshot(snapshot types.ProjectSnapshot) (types.ProjectSnapshot, error) {
	if snapshot.OrganizationID == "" {
		return snapshot, ErrMissingOrganizationID
	}
	if snapshot.ProjectHash == 0 {
		return snapshot, fmt.Errorf("Hash must be set")
	}
	if snapshot.Project.ID == "" {
		return snapshot, ErrMissingProjectID
	}
	if v, err := b.FindOneSnapshot(snapshot); err != nil {
		return snapshot, err
	} else if v != nil {
		return *v, fmt.Errorf("There already exists a snapshot with these exact contents")
	}
	entity, err := b.NewEntity(snapshot.Entity)
	if err != nil {
		return snapshot, err
	}
	snapshot.Entity = entity

	err = Create(b, BucketSnapshot, snapshot)
	return snapshot, err
}

func (bb *BBolter) FindOneSnapshot(filter ...types.ProjectSnapshot) (*types.ProjectSnapshot, error) {
	return FindOne(bb, BucketSnapshot, func(s types.ProjectSnapshot) bool {
		for _, f := range filter {
			if f.OrganizationID == "" {
				bb.l.Warn().Msg("Received a snapshot-filter without organization-id")
			}
			if snapshotFilter(f, s) {
				return true
			}
		}
		return false
	})
}

func (bb *BBolter) UpdateSnapshot(id string, payload types.ProjectSnapshot) (types.ProjectSnapshot, error) {
	return Update(bb, BucketSnapshot, id, func(t types.ProjectSnapshot) (types.ProjectSnapshot, error) {
		if payload.Project.ID != "" && t.Project.ID != payload.Project.ID {
			return t, fmt.Errorf("Cannot replace a projects snappshot with a different project")
		}
		if payload.ProjectHash == t.ProjectHash {
			return t, ErrNoFieldsChanged
		}
		if payload.ProjectHash != 0 && t.ProjectHash != payload.ProjectHash {
			t.ProjectHash = payload.ProjectHash
			t.Project = payload.Project
		}
		err := t.Entity.Update(payload.Entity)
		if err != nil {
			return t, err
		}

		return t, nil
	})
}
