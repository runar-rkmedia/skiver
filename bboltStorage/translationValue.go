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

func (bb *BBolter) UpdateTranslationValue(tv types.TranslationValue) (types.TranslationValue, error) {
	var ex types.TranslationValue
	err := bb.updater(tv.ID, BucketTranslationValue, func(b []byte) ([]byte, error) {
		err := bb.Unmarshal(b, &ex)
		if err != nil {
			return nil, err
		}
		err = ex.Entity.Update(tv.Entity)
		if err != nil {
			return nil, err
		}
		if tv.Value != "" {
			ex.Value = tv.Value
		}
		if tv.Source != "" {
			ex.Source = tv.Source
		}

		if len(tv.Context) > 0 {
			if ex.Context == nil {
				ex.Context = map[string]string{}
			}
			for k, v := range tv.Context {
				ex.Context[k] = v
			}
		}
		return bb.Marshal(ex)
	})
	bb.PublishChange(PubTypeTranslationValue, PubVerbUpdate, ex)

	return ex, err
}
func (b *BBolter) CreateTranslationValue(tv types.TranslationValue) (types.TranslationValue, error) {
	if tv.LocaleID == "" {
		return tv, fmt.Errorf("empty locale-id")
	}
	if tv.TranslationID == "" {
		return tv, fmt.Errorf("empty translation-id")
	}
	existing, err := b.GetTranslationValueFilter(tv)
	if err != nil {
		return types.TranslationValue{}, err
	}
	if existing != nil {
		if b.l.HasDebug() {
			b.l.Debug().Interface("existing", existing).Interface("input", tv).Msg("Translationvalue already exists")
		}
		return *existing, fmt.Errorf("Translationvalue already exists")
	}
	tv.Entity, err = b.NewEntity(tv.Entity)
	if err != nil {
		return tv, err
	}

	var t types.Translation
	err = b.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(BucketTranslationValue)
		existing := bucket.Get([]byte(tv.ID))
		if existing != nil {
			return fmt.Errorf("there already exists a project with this ID")
		}
		{
			// Add the Translation to the Category
			bucketTranslation := tx.Bucket(BucketTranslation)
			translation := bucketTranslation.Get([]byte(tv.TranslationID))
			if translation == nil {
				return fmt.Errorf("Failed to lookup translation-id: %w", err)
			}
			err := b.Unmarshal(translation, &t)
			if err != nil {
				return err
			}
			t.ValueIDs = append(t.ValueIDs, tv.ID)
			t.UpdatedAt = nowPointer()
			bytes, err := b.Marshal(t)
			if err != nil {
				return err
			}
			err = bucketTranslation.Put([]byte(t.ID), bytes)
			if err != nil {
				return err
			}
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
	b.PublishChange(PubTypeTranslation, PubVerbUpdate, t)
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
