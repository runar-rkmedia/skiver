package bboltStorage

import (
	"fmt"

	"github.com/runar-rkmedia/skiver/types"
	bolt "go.etcd.io/bbolt"
)

func (b *BBolter) GetTranslation(ID string) (*types.Translation, error) {
	var u types.Translation
	err := b.GetItem(BucketTranslation, ID, &u)
	return &u, err
}

func (b *BBolter) CreateTranslation(translation types.Translation) (types.Translation, error) {
	if translation.CategoryID == "" {
		return translation, fmt.Errorf("Missing CategoryID: %w", ErrMissingIdArg)
	}
	existing, err := b.GetTranslationFilter(translation)
	if err != nil {
		return translation, err
	}
	if existing != nil {
		return *existing, fmt.Errorf("Translation already exists")
	}
	translation.Entity, err = b.NewEntity(translation.Entity)
	if err != nil {
		return translation, err
	}

	var c types.Category
	err = b.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(BucketTranslation)
		existing := bucket.Get([]byte(translation.ID))
		if existing != nil {
			return fmt.Errorf("there already exists a translation with this ID")
		}
		{
			// Add the Translation to the Category
			bucketCategory := tx.Bucket(BucketCategory)
			category := bucketCategory.Get([]byte(translation.CategoryID))
			if category == nil {
				return fmt.Errorf("Failed to lookup category-id: %w", err)
			}
			err := b.Unmarshal(category, &c)
			if err != nil {
				return err
			}
			c.TranslationIDs = append(c.TranslationIDs, translation.ID)
			c.UpdatedAt = nowPointer()
			bytes, err := b.Marshal(c)
			if err != nil {
				return err
			}
			err = bucketCategory.Put([]byte(c.ID), bytes)
			if err != nil {
				return err
			}
		}
		bytes, err := b.Marshal(translation)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(translation.ID), bytes)
	})
	if err != nil {
		return translation, err
	}

	b.PublishChange(PubTypeTranslation, PubVerbCreate, translation)
	b.PublishChange(PubTypeCategory, PubVerbUpdate, c)
	go b.UpdateMissingWithNewIds(types.MissingTranslation{Translation: translation.Key, TranslationID: translation.ID})
	return translation, err
}

// TODO: complete implementation
func (b *BBolter) UpdateTranslation(id string, paylaod types.Translation) (types.Translation, error) {
	if id == "" {
		return paylaod, ErrMissingIdArg
	}
	existing, err := b.GetTranslation(id)
	if err != nil {
		return *existing, err
	}
	var c types.Translation
	err = b.Update(func(tx *bolt.Tx) error {

		bucket := tx.Bucket(BucketTranslation)
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
		if paylaod.Key != c.Key {
			c.Key = paylaod.Key
			needsUpdate = true
		}
		if paylaod.Title != "" && paylaod.Title != c.Title {
			c.Title = paylaod.Title
			needsUpdate = true
		}
		if paylaod.Description != "" && paylaod.Description != c.Description {
			c.Description = paylaod.Description
			needsUpdate = true
		}
		if len(paylaod.Variables) > 0 {
			c.Variables = paylaod.Variables
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

	b.PublishChange(PubTypeTranslation, PubVerbUpdate, c)
	return c, err

}

func (bb *BBolter) GetTranslations() (map[string]types.Translation, error) {
	us := make(map[string]types.Translation)
	err := bb.Iterate(BucketTranslation, func(key, b []byte) bool {
		var u types.Translation
		bb.Unmarshal(b, &u)
		us[string(key)] = u
		return false
	})
	if err == ErrNotFound {
		return us, nil
	}
	return us, err
}

func (bb *BBolter) GetTranslationFilter(filter ...types.Translation) (*types.Translation, error) {
	t, err := bb.GetTranslationsFilter(1, filter...)
	if err != nil {
		return nil, err
	}
	for _, v := range t {
		return &v, nil
	}
	return nil, nil
}
func (bb *BBolter) GetTranslationsFilter(max int, filter ...types.Translation) (map[string]types.Translation, error) {
	u := make(map[string]types.Translation)
	len := 0
	err := bb.Iterate(BucketTranslation, func(key, b []byte) bool {
		var uu types.Translation
		err := bb.Unmarshal(b, &uu)
		if err != nil {
			bb.l.Error().Err(err).Msg("failed to unmarshal user")
			return false
		}
		for _, f := range filter {
			if f.CategoryID != "" && f.CategoryID != uu.CategoryID {
				continue
			}
			if f.Key != "" && f.Key != uu.Key {
				continue
			}
			u[uu.ID] = uu
			len++
			if max == 0 {
				return false
			}
			return len >= max
		}
		return false
	})
	return u, err
}
