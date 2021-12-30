package bboltStorage

import (
	"fmt"

	"github.com/runar-rkmedia/skiver/types"
	bolt "go.etcd.io/bbolt"
)

func (b *BBolter) GetTranslationValue(ID string) (*types.TranslationValue, error) {
	var u types.TranslationValue
	err := b.GetItem(BucketTranslationValue, ID, &u)
	return &u, err
}

func (b *BBolter) CreateTranslationValue(tv types.TranslationValue) (types.TranslationValue, error) {
	existing, err := b.GetTranslationValueFilter(tv)
	if err != ErrNotFound {
		return *existing, fmt.Errorf("Already exists")
	}
	tv.Entity = b.NewEntity()

	err = b.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(BucketTranslationValue)
		existing := bucket.Get([]byte(tv.ID))
		if existing != nil {
			return fmt.Errorf("there already exists a project with this ID")
		}
		bytes, err := b.Marshal(tv)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(tv.ID), bytes)
	})
	if err != nil {
		return tv, err
	}

	b.PublishChange(PubTypeTranslationValue, PubVerbCreate, tv)
	return tv, err
}

func (bb *BBolter) GetTranslationValues() (map[string]types.TranslationValue, error) {
	us := make(map[string]types.TranslationValue)
	err := bb.Iterate(BucketTranslationValue, func(key, b []byte) bool {
		var u types.TranslationValue
		bb.Unmarshal(b, &u)
		us[string(key)] = u
		return false
	})
	if err == ErrNotFound {
		return us, nil
	}
	return us, err
}

func (bb *BBolter) GetTranslationValueFilter(filter ...types.TranslationValue) (*types.TranslationValue, error) {
	var u types.TranslationValue
	err := bb.Iterate(BucketTranslationValue, func(key, b []byte) bool {
		var uu types.TranslationValue
		err := bb.Unmarshal(b, &uu)
		if err != nil {
			bb.l.Error().Err(err).Msg("failed to unmarshal user")
			return false
		}
		for _, f := range filter {
			if f.LocaleID != "" && f.LocaleID != uu.LocaleID {
				continue
			}
			if f.TranslationID != "" && f.TranslationID != uu.TranslationID {
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
