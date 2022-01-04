package bboltStorage

import (
	"fmt"

	"github.com/runar-rkmedia/skiver/types"
	bolt "go.etcd.io/bbolt"
)

func (b *BBolter) GetLocale(ID string) (types.Locale, error) {
	var u types.Locale
	err := b.GetItem(BucketLocale, ID, &u)
	return u, err
}

func (b *BBolter) CreateLocale(locale types.Locale) (types.Locale, error) {
	existing, err := b.GetLocaleFilter(locale)
	if err != nil {
		return locale, err
	}
	if existing != nil {
		return *existing, fmt.Errorf("Locale already exists")
	}
	locale.Entity = b.NewEntity()

	err = b.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(BucketLocale)
		existing := bucket.Get([]byte(locale.ID))
		if existing != nil {
			return fmt.Errorf("there already exists a locale with this ID")
		}
		bytes, err := b.Marshal(locale)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(locale.ID), bytes)
	})
	if err != nil {
		return locale, err
	}

	b.PublishChange(PubTypeLocale, PubVerbCreate, locale)
	return locale, err
}
func (bb *BBolter) GetLocales() (map[string]types.Locale, error) {
	us := make(map[string]types.Locale)
	err := bb.Iterate(BucketLocale, func(key, b []byte) bool {
		var u types.Locale
		bb.Unmarshal(b, &u)
		us[string(key)] = u
		return false
	})
	if err == ErrNotFound {
		return us, nil
	}
	return us, err
}
func (bb *BBolter) GetLocaleByFirstMatch(name string) (*types.Locale, error) {
	// OR-like search:
	t, err := bb.GetLocaleFilter(
		types.Locale{Entity: types.Entity{ID: name}},
		types.Locale{Iso639_1: name},
		types.Locale{Iso639_2: name},
		types.Locale{Iso639_3: name},
		types.Locale{IETF: name},
	)
	return t, err
}
func (bb *BBolter) GetLocaleFilter(filter ...types.Locale) (*types.Locale, error) {
	var u *types.Locale
	err := bb.Iterate(BucketLocale, func(key, b []byte) bool {
		var uu types.Locale
		err := bb.Unmarshal(b, &uu)
		if err != nil {
			bb.l.Error().Err(err).Msg("failed to unmarshal user")
			return false
		}
		for _, f := range filter {
			if f.ID != "" && f.ID != uu.ID {
				continue
			}
			if f.Iso639_1 != "" && f.Iso639_1 != uu.Iso639_1 {
				continue
			}
			if f.Iso639_2 != "" && f.Iso639_2 != uu.Iso639_2 {
				continue
			}
			if f.Iso639_3 != "" && f.Iso639_3 != uu.Iso639_3 {
				continue
			}
			if f.IETF != "" && f.IETF != uu.IETF {
				continue
			}
			if f.Title != "" && f.Title != uu.Title {
				continue
			}
			u = &uu
			return true
		}
		return false
	})
	if err != nil {
		return u, err
	}
	return u, err
}
