package bboltStorage

import (
	"fmt"

	"github.com/runar-rkmedia/skiver/types"
	bolt "go.etcd.io/bbolt"
)

func (b *BBolter) GetTranslation(ID string) (*types.Translation, error) {
	var u types.Translation
	err := b.GetItem(BucketTranslations, ID, &u)
	return &u, err
}

func (b *BBolter) CreateTranslation(translation types.Translation) (types.Translation, error) {
	if translation.ProjectID == "" {
		return translation, fmt.Errorf("Missing ProjectID: %w", ErrMissingIdArg)
	}
	if translation.LocaleID == "" {
		return translation, fmt.Errorf("Missing LocaleID: %w", ErrMissingIdArg)
	}
	if p, err := b.GetProject(translation.ProjectID); err != nil {
		return translation, fmt.Errorf("Failed to lookup project-id: %w", err)
	} else if p == nil {
		return translation, fmt.Errorf("Did not find project with id %s: %w", translation.ProjectID, ErrNotFound)
	}
	if p, err := b.GetLocale(translation.LocaleID); err != nil {
		return translation, fmt.Errorf("Failed to lookup locale-id: %w", err)
	} else if p.ID == "" {
		return translation, fmt.Errorf("Did not find locale with id %s: %w", translation.LocaleID, ErrNotFound)
	}
	existing, err := b.GetTranslationFilter(translation)
	if err != ErrNotFound {
		return *existing, fmt.Errorf("Already exists")
	}
	translation.Entity = b.NewEntity()

	err = b.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(BucketTranslations)
		existing := bucket.Get([]byte(translation.ID))
		if existing != nil {
			return fmt.Errorf("there already exists a translation with this ID")
		}
		fmt.Println(translation)
		bytes, err := b.Marshal(translation)
		if err != nil {
			fmt.Println("tttttt", err)
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
	err := bb.Iterate(BucketTranslations, func(key, b []byte) bool {
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
	err := bb.Iterate(BucketTranslations, func(key, b []byte) bool {
		var uu types.Translation
		err := bb.Unmarshal(b, &uu)
		if err != nil {
			bb.l.Error().Err(err).Msg("failed to unmarshal user")
			return false
		}
		for _, f := range filter {
			if f.ProjectID != "" && f.ProjectID != uu.ProjectID {
				continue
			}
			if f.Prefix != "" && f.Prefix != uu.Prefix {
				continue
			}
			if f.Key != "" && f.Key != uu.Key {
				continue
			}
			if f.LocaleID != "" && f.LocaleID != uu.LocaleID {
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
