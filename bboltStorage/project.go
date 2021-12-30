package bboltStorage

import (
	"fmt"

	"github.com/runar-rkmedia/skiver/types"
	bolt "go.etcd.io/bbolt"
)

func (b *BBolter) GetProject(ID string) (*types.Project, error) {
	var u types.Project
	err := b.GetItem(BucketProject, ID, &u)
	return &u, err
}

func (b *BBolter) CreateProject(project types.Project) (types.Project, error) {
	existing, err := b.GetProjectFilter(project)
	if err != ErrNotFound {
		return *existing, fmt.Errorf("Already exists")
	}
	project.Entity = b.NewEntity()

	err = b.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(BucketProject)
		existing := bucket.Get([]byte(project.ID))
		if existing != nil {
			return fmt.Errorf("there already exists a project with this ID")
		}
		bytes, err := b.Marshal(project)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(project.ID), bytes)
	})
	if err != nil {
		return project, err
	}

	b.PublishChange(PubTypeProject, PubVerbCreate, project)
	return project, err
}

func (bb *BBolter) GetProjects() (map[string]types.Project, error) {
	us := make(map[string]types.Project)
	err := bb.Iterate(BucketProject, func(key, b []byte) bool {
		var u types.Project
		bb.Unmarshal(b, &u)
		us[string(key)] = u
		return false
	})
	if err == ErrNotFound {
		return us, nil
	}
	return us, err
}

func (bb *BBolter) GetProjectFilter(filter ...types.Project) (*types.Project, error) {
	var u types.Project
	err := bb.Iterate(BucketProject, func(key, b []byte) bool {
		var uu types.Project
		err := bb.Unmarshal(b, &uu)
		if err != nil {
			bb.l.Error().Err(err).Msg("failed to unmarshal user")
			return false
		}
		for _, f := range filter {
			if f.Title != "" && f.Title != uu.Title {
				continue
			}
			if f.Description != "" && f.Description != uu.Description {
				continue
			}
			u = uu
			return true
		}
		return false
	})
	if err != nil {
		return nil, err
	}
	return &u, err
}
