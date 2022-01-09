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
	if project.ShortName == "" {
		return project, fmt.Errorf("Missing short-name: %w", ErrMissingIdArg)
	}
	if project.Title == "" {
		return project, fmt.Errorf("Missing title: %w", ErrMissingIdArg)
	}
	existing, err := b.GetProjectFilter(types.Project{ShortName: project.ShortName})
	if err != nil {
		return *existing, err
	}
	if existing != nil {

		return *existing, fmt.Errorf("Already exists")
	}
	project.Entity, err = b.NewEntity(project.CreatedBy)
	if err != nil {
		return project, err
	}

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

	go b.UpdateMissingWithNewIds(types.MissingTranslation{Project: project.ShortName, ProjectID: project.ID})
	b.PublishChange(PubTypeProject, PubVerbCreate, project)
	return project, err
}
func (b *BBolter) UpdateProject(project types.Project) (types.Project, error) {
	err := b.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(BucketProject)
		existing := bucket.Get([]byte(project.ID))
		if existing == nil {
			return fmt.Errorf("Project was not found")
		}
		var ex types.Project
		err := b.Unmarshal(existing, &ex)
		if err != nil {
			return err
		}
		if project.ShortName != "" {
			ex.ShortName = project.ShortName
		}
		if project.Title != "" {
			ex.Title = project.Title
		}
		if project.Description != "" {
			ex.Description = project.Description
		}
		if project.IncludedTags != nil {
			ex.IncludedTags = project.IncludedTags
		}
		if project.CategoryIDs != nil {
			ex.CategoryIDs = project.CategoryIDs
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

	go b.UpdateMissingWithNewIds(types.MissingTranslation{Project: project.ShortName, ProjectID: project.ID})
	b.PublishChange(PubTypeProject, PubVerbUpdate, project)
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

func (bb *BBolter) GetProjectByIDOrShortName(shortNameOrId string) (*types.Project, error) {
	p, err := bb.GetProject(shortNameOrId)
	if err == nil {
		return p, err
	} else if err != ErrNotFound {
		return nil, err
	}
	return bb.GetProjectByShortName(shortNameOrId)
}

func (bb *BBolter) GetProjectByShortName(shortName string) (*types.Project, error) {
	return bb.GetProjectFilter(
		types.Project{Entity: types.Entity{ID: shortName}},
		types.Project{ShortName: shortName},
	)
}

func (bb *BBolter) GetProjectFilter(filter ...types.Project) (*types.Project, error) {
	var u *types.Project
	err := bb.Iterate(BucketProject, func(key, b []byte) bool {
		var uu types.Project
		err := bb.Unmarshal(b, &uu)
		if err != nil {
			bb.l.Error().Err(err).Msg("failed to unmarshal user")
			return false
		}
		for _, f := range filter {
			if f.ID != "" && f.ID != uu.ID {
				continue
			}
			if f.ShortName != "" && f.ShortName != uu.ShortName {
				continue
			}
			if f.Description != "" && f.Description != uu.Description {
				continue
			}
			u = &uu
			return true
		}
		return false
	})
	if err != nil {
		if err == ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return u, err
}
