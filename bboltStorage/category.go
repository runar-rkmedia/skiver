package bboltStorage

import (
	"fmt"

	"github.com/runar-rkmedia/skiver/types"
	bolt "go.etcd.io/bbolt"
)

func (b *BBolter) GetCategory(ID string) (*types.Category, error) {
	var u types.Category
	err := b.GetItem(BucketCategory, ID, &u)
	return &u, err
}

func (b *BBolter) CreateCategory(category types.Category) (types.Category, error) {
	existing, err := b.GetCategoryFilter(category)
	if err != nil {
		return *existing, err
	}
	if existing != nil {
		return *existing, fmt.Errorf("Category already exists")
	}
	category.Entity, err = b.NewEntity(category.CreatedBy)
	if err != nil {
		return category, err
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

func (bb *BBolter) GetCategoryFilter(filter ...types.Category) (*types.Category, error) {
	var u *types.Category
	err := bb.Iterate(BucketCategory, func(key, b []byte) bool {
		var uu types.Category
		err := bb.Unmarshal(b, &uu)
		if err != nil {
			bb.l.Error().Err(err).Msg("failed to unmarshal user")
			return false
		}
		for _, f := range filter {
			if f.ProjectID != "" && f.ProjectID != uu.ProjectID {
				continue
			}
			if f.Key != "" && f.Key != uu.Key {
				continue
			}
			u = &uu
			return true
		}
		return false
	})
	return u, err
}
