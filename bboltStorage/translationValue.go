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
	if err != nil {
		return types.TranslationValue{}, err
	}
	if existing != nil {
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

	tvs, err := bb.GetTranslationValuesFilter(1, filter...)
	if err != nil {
		return nil, err
	}
	for _, v := range tvs {
		return &v, err
	}
	return nil, err
}
func (bb *BBolter) GetTranslationValuesFilter(max int, filter ...types.TranslationValue) (map[string]types.TranslationValue, error) {
	tvs := make(map[string]types.TranslationValue)
	len := 0
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
			tvs[uu.ID] = uu
			len++
			if max == 0 {
				return false
			}
			return len >= max
		}
		return false
	})
	if err == ErrNotFound {
		err = nil
	}
	if err != nil {
		return tvs, err
	}
	return tvs, err
}
