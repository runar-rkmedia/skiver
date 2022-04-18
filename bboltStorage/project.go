package bboltStorage

import (
	"errors"
	"fmt"
	"reflect"

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
	project.Entity, err = b.NewEntity(project.Entity)
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

func (b *BBolter) UpdateProject(id string, project types.Project) (types.Project, error) {
	if id == "" {
		return project, ErrMissingIdArg
	}
	existing, err := b.GetProject(id)
	if err != nil {
		return *existing, err
	}
	// TODO: Must only be unique within organization
	if project.ShortName != "" && existing.ShortName != project.ShortName {
		ex, err := b.GetProjectByIDOrShortName(project.ShortName)
		if err != nil {
			return *ex, err
		}
		if ex != nil {
			return *ex, errors.New("Shortname is already taken")
		}
	}
	var c types.Project
	err = b.Update(func(tx *bolt.Tx) error {

		bucket := tx.Bucket(BucketProject)
		existing := bucket.Get([]byte(id))
		if existing == nil {
			return ErrNotFound
		}
		err := b.Unmarshal(existing, &c)
		if err != nil {
			return err
		}
		needsUpdate := false
		if project.ShortName != c.ShortName {
			c.ShortName = project.ShortName
			needsUpdate = true
		}
		if project.Title != c.Title {
			c.Title = project.Title
			needsUpdate = true
		}
		if project.Description != c.Description {
			c.Description = project.Description
			needsUpdate = true
		}
		if len(project.LocaleIDs) != 0 && !reflect.DeepEqual(project.LocaleIDs, c.LocaleIDs) {
			c.LocaleIDs = project.LocaleIDs
			needsUpdate = true
		}
		if len(project.Snapshots) != 0 && !reflect.DeepEqual(project.Snapshots, c.Snapshots) {
			c.Snapshots = project.Snapshots
			needsUpdate = true
		}

		if !needsUpdate {
			return ErrNoFieldsChanged
		}
		err = c.Entity.Update(project.Entity)
		if err != nil {
			return err
		}

		bytes, err := b.Marshal(c)
		if err != nil {
			return err
		}
		err = bucket.Put([]byte(c.ID), bytes)

		return nil
	})
	if err != nil {
		return c, err
	}

	b.PublishChange(PubTypeProject, PubVerbUpdate, c)

	return c, err

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
	return FindOne(bb, BucketProject, func(t types.Project) bool {
		for _, f := range filter {
			if f.OrganizationID == "" {
				bb.l.Warn().Msg("Received a user-filter without organization-id")
			}
			if projectFilter(f, t) {
				return true
			}
		}
		return false
	})
}
func (bb *BBolter) FindProjects(max int, filter ...types.Project) (map[string]types.Project, error) {
	return Find(bb, BucketProject, max, func(uu types.Project) bool {
		if len(filter) == 0 {
			return true
		}
		for _, f := range filter {
			if projectFilter(f, uu) {
				return true
			}
		}
		return false
	})
}

func projectFilter(f, uu types.Project) bool {
	if f.OrganizationID != "" && f.OrganizationID != uu.OrganizationID {
		return false
	}
	if f.ID != "" && f.ID != uu.ID {
		return false
	}
	if f.ShortName != "" && f.ShortName != uu.ShortName {
		return false
	}
	if f.ID != "" && f.ID != uu.ID {
		return false
	}
	return true
}
