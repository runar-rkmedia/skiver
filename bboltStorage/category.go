package bboltStorage

import (
	"fmt"

	"github.com/runar-rkmedia/skiver/internal"
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
func (bb *BBolter) CreateSubCategory(rootCategoryID, targetParentCategoryId string, category types.Category) (types.Category, error) {
	var err error
	if rootCategoryID == "" {
		return category, fmt.Errorf("Missing root-category-id %w", ErrMissingIdArg)
	}
	if targetParentCategoryId == "" {
		return category, fmt.Errorf("Missing target-parent-category-id %w", ErrMissingIdArg)
	}
	if category.Key == "" || category.Key == types.RootCategory {
		return category, fmt.Errorf("A sub-category cannot be a root-category (empty category.key) (%w)", ErrDuplicate)
	}
	category.Entity, err = bb.NewEntity(category.Entity)
	if err != nil {
		return category, err
	}

	// Reads are less *expensive* than opening a write, so we first make sure the category exists
	filter := types.CategoryFilter{ID: rootCategoryID}
	if rootCategoryID != targetParentCategoryId {
		filter.SubCategory = []types.CategoryFilter{{ID: targetParentCategoryId}}
	}
	c, err := bb.FindOneCategory(filter)
	if err != nil {
		return category, fmt.Errorf("error during CreateSubCategory: %w", err)
	}
	if c == nil {
		return category, fmt.Errorf("failed to find category during CreateSubCategory: %w", err)
	}

	return Update(bb, BucketCategory, rootCategoryID, func(t types.Category) (types.Category, error) {

		e, err := updateEntity(t.Entity, category.Entity)
		if err != nil {
			return t, err
		}
		if t.ID == targetParentCategoryId {
			if t.SubCategories == nil {
				t.SubCategories = []types.Category{}
			}
			// panic(t.SubCategories)
			t.SubCategories = append(t.SubCategories, t)
		}
		fmt.Println(internal.MustYaml(t), "\n", targetParentCategoryId)
		panic(internal.MustYaml(t))
		t.Entity = e

		return t, nil
	})

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
		return *existing, fmt.Errorf("Category already exists")
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
