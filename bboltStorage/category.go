package bboltStorage

import (
	"fmt"

	"github.com/runar-rkmedia/skiver/types"
	bolt "go.etcd.io/bbolt"
)

func (bb *BBolter) GetCategory(ID string) (*types.Category, error) {
	return Get[types.Category](bb, BucketCategory, ID)
}

func (bb *BBolter) FindOneCategory(filter ...types.CategoryFilter) (*types.Category, error) {
	return FindOne(bb, BucketCategory, func(t types.Category) bool {
		for _, f := range filter {
			if f.OrganizationID == "" {
				bb.l.Warn().Msg("Received a Category without organization-id")
			}
			if t.Filter(f) {
				return true
			}
		}
		return false
	})
}
func (bb *BBolter) FindCategories(max int, filter ...types.CategoryFilter) (map[string]types.Category, error) {
	return Find(bb, BucketCategory, max, func(cat types.Category) bool {
		if len(filter) == 0 {
			return true
		}
		for _, f := range filter {
			if cat.Filter(f) {
				return true
			}
		}
		return false
	})
}

// TODO: complete implementation
func (b *BBolter) UpdateCategory(id string, category types.Category) (types.Category, error) {
	if id == "" {
		return category, ErrMissingIdArg
	}
	existing, err := b.GetCategory(id)
	if err != nil {
		return *existing, err
	}
	var c types.Category
	err = b.Update(func(tx *bolt.Tx) error {

		bucket := tx.Bucket(BucketCategory)
		existing := bucket.Get([]byte(id))
		if existing == nil {
			return ErrNotFound
		}
		err := b.Unmarshal(existing, &c)
		if err != nil {
			return err
		}
		needsUpdate := false
		// TODO: ensure key-uniqueness
		if category.Key != c.Key {
			c.Key = category.Key
			needsUpdate = true
		}
		if category.Title != c.Title {
			c.Title = category.Title
			needsUpdate = true
		}
		if category.Description != c.Description {
			c.Description = category.Description
			needsUpdate = true
		}

		if !needsUpdate {
			return ErrNoFieldsChanged
		}
		c.UpdatedAt = nowPointer()

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

	b.PublishChange(PubTypeCategory, PubVerbUpdate, c)

	return c, err

}
func (b *BBolter) CreateCategory(category types.Category) (types.Category, error) {
	if category.ProjectID == "" {
		return category, ErrMissingProject
	}
	existing, err := b.FindOneCategory(category.AsUniqueFilter())
	if err != nil {
		return *existing, err
	}
	if existing != nil {
		b.l.Warn().Interface("existing", existing).Interface("input", category).Err(err).Msg("category already exists")
		return *existing, fmt.Errorf("Category-create-error: %w", ErrDuplicate)
	}
	category.Entity, err = b.NewEntity(category.Entity)
	if err != nil {
		return category, err
	}
	if category.Key == "" {
		category.Key = types.RootCategory
	}

	var p types.Project
	err = b.Update(func(tx *bolt.Tx) error {
		bucketCategory := tx.Bucket(BucketCategory)
		existing := bucketCategory.Get([]byte(category.ID))
		if existing != nil {
			return fmt.Errorf("there already exists a category with this ID")
		}
		{
			// Add the category to the project
			bucketProject := tx.Bucket(BucketProject)
			project := bucketProject.Get([]byte(category.ProjectID))
			if project == nil {
				return fmt.Errorf("did not find this project")
			}
			err := b.Unmarshal(project, &p)
			if err != nil {
				return err
			}
			p.CategoryIDs = append(p.CategoryIDs, category.ID)
			p.UpdatedAt = nowPointer()
			bytes, err := b.Marshal(p)
			if err != nil {
				return err
			}
			err = bucketProject.Put([]byte(p.ID), bytes)
			if err != nil {
				return err
			}
		}
		bytes, err := b.Marshal(category)
		if err != nil {
			return err
		}
		return bucketCategory.Put([]byte(category.ID), bytes)
	})
	if err != nil {
		return category, err
	}

	b.PublishChange(PubTypeCategory, PubVerbCreate, category)
	go b.UpdateMissingWithNewIds(types.MissingTranslation{Category: category.Key, CategoryID: category.ID})
	b.PublishChange(PubTypeProject, PubVerbUpdate, p)
	return category, err
}

func (bb *BBolter) GetCategories() (map[string]types.Category, error) {
	us := make(map[string]types.Category)
	err := bb.Iterate(BucketCategory, func(key, b []byte) bool {
		var u types.Category
		bb.Unmarshal(b, &u)
		us[string(key)] = u
		return false
	})
	if err == ErrNotFound {
		return us, nil
	}
	return us, err
}
