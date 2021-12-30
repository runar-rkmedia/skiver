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
	if p, err := b.GetCategory(translation.CategoryID); err != nil {
		return translation, fmt.Errorf("Failed to lookup category-id: %w", err)
	} else if p.ID == "" {
		return translation, fmt.Errorf("Did not find category with id %s: %w", translation.CategoryID, ErrNotFound)
	}
	existing, err := b.GetTranslationFilter(translation)
	if err != ErrNotFound {
		return *existing, fmt.Errorf("Already exists")
	}
	translation.Entity = b.NewEntity()

	err = b.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(BucketTranslation)
		existing := bucket.Get([]byte(translation.ID))
		if existing != nil {
			return fmt.Errorf("there already exists a translation with this ID")
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
	return translation, err
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
	var u types.Translation
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
