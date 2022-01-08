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

// Updates an existing entity-struct with changes.
// NOTE: This does NOT update the db-value itself., but is meant as a helper-func
func updateEntity(existing, changes types.Entity) (types.Entity, error) {
	if existing.ID == "" {
		return existing, fmt.Errorf("The existing-value does not have an ID.")
	}
	updatedBy := changes.UpdatedBy
	if updatedBy == "" {
		updatedBy = changes.CreatedBy
	}
	if updatedBy == "" {
		return existing, fmt.Errorf("UpdatedBy is not set")
	}

	if changes.UpdatedAt == nil {
		changes.UpdatedAt = nowPointer()
	}
	existing.UpdatedAt = changes.UpdatedAt
	existing.UpdatedBy = changes.UpdatedBy
	return existing, nil
}

func (bb *BBolter) UpdateTranslationValue(tv types.TranslationValue) (types.TranslationValue, error) {
	if tv.Value == "" {
		return tv, fmt.Errorf("empty value")
	}
	var ex types.TranslationValue
	err := bb.updater(tv.ID, BucketTranslationValue, func(b []byte) ([]byte, error) {
		err := bb.Unmarshal(b, &ex)
		if err != nil {
			return nil, err
		}
		entity, err := updateEntity(ex.Entity, tv.Entity)
		if err != nil {
			return nil, err
		}
		ex.Entity = entity
		ex.Value = tv.Value
		return bb.Marshal(ex)
	})

	return ex, err
}
func (b *BBolter) CreateTranslationValue(tv types.TranslationValue) (types.TranslationValue, error) {
	if tv.Value == "" {
		return tv, fmt.Errorf("empty value")
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
	tv.Entity, err = b.NewEntity(tv.CreatedBy)
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
