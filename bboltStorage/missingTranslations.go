package bboltStorage

import (
	"fmt"
	"strings"

	"github.com/runar-rkmedia/skiver/types"
	bolt "go.etcd.io/bbolt"
)

// Updates all MissingTranslations with Project/ProjectID
// TODO: do the same for Category, Translation and Localization
func (bb *BBolter) UpdateMissingWithNewIds(payload types.MissingTranslation) (map[string]types.MissingTranslation, error) {

	updated := map[string]types.MissingTranslation{}
	err := bb.DB.Batch(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(BucketMissing)
		c := bucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var m types.MissingTranslation
			err := bb.Unmarshal(v, &m)
			if err != nil {
				return err
			}
			shouldUpdate := false
			if payload.ProjectID != "" && payload.Project != "" {
				if payload.ProjectID == m.ProjectID && payload.Project != m.Project {
					m.Project = payload.Project
					shouldUpdate = true
				} else if payload.ProjectID != m.ProjectID && payload.Project == m.Project {
					m.ProjectID = payload.ProjectID
					shouldUpdate = true
				}
			}
			if payload.CategoryID != "" && payload.Category != "" {
				if payload.CategoryID == m.CategoryID && payload.Category != m.Category {
					m.Category = payload.Category
					shouldUpdate = true
				} else if payload.CategoryID != m.CategoryID && payload.Category == m.Category {
					m.CategoryID = payload.CategoryID
					shouldUpdate = true
				}
			}
			if payload.TranslationID != "" && payload.Translation != "" {
				if payload.TranslationID == m.TranslationID && payload.Translation != m.Translation {
					m.Translation = payload.Translation
					shouldUpdate = true
				} else if payload.TranslationID != m.TranslationID && payload.Translation == m.Translation {
					m.TranslationID = payload.TranslationID
					shouldUpdate = true
				}
			}
			if payload.LocaleID != "" && payload.Locale != "" {
				if payload.LocaleID == m.LocaleID && payload.Locale != m.Locale {
					m.Locale = payload.Locale
					shouldUpdate = true
				} else if payload.LocaleID != m.LocaleID && payload.Locale == m.Locale {
					m.LocaleID = payload.LocaleID
					shouldUpdate = true
				}
			}
			if !shouldUpdate {
				continue
			}
			mb, err := bb.Marshal(m)
			if err != nil {
				return err
			}
			err = bucket.Put(k, mb)
			if err != nil {
				return err
			}
			updated[string(k)] = m
		}
		return nil
	})
	for _, v := range updated {
		bb.PublishChange(PubTypeMissingTranslation, PubVerbUpdate, v)
	}
	return updated, err
}
func (b *BBolter) ReportMissing(key types.MissingTranslation) (*types.MissingTranslation, error) {
	if key.Project == "" {
		return &key, fmt.Errorf("Missing Project: %w", ErrMissingIdArg)
	}
	if key.Category == "" {
		return &key, fmt.Errorf("Missing Category: %w", ErrMissingIdArg)
	}
	if key.Translation == "" {
		return &key, fmt.Errorf("Missing Translation: %w", ErrMissingIdArg)
	}
	if key.Locale == "" {
		return &key, fmt.Errorf("Missing Locale: %w", ErrMissingIdArg)
	}
	entity, err := b.NewEntity(key.Entity)

	switch err {
	case nil:
		break
	case ErrMissingOrganizationID:
		// OrganizationID is filled later on, if possible
		break
	default:
		return nil, err
	}
	key.Entity = entity
	key.ID = strings.Join([]string{key.Project, key.Locale, key.Category, key.Translation}, " / ")
	// TODO: make sure this is isolated per OrganizationID

	if key.ProjectID == "" {
		project, err := b.GetProjectByShortName(key.Project)
		if err != nil {
			return &key, fmt.Errorf("failed to lookup project: %w", err)
		}
		if project != nil {
			key.ProjectID = project.ID
			if key.OrganizationID == "" {
				key.OrganizationID = project.OrganizationID
			}
		}
	}
	if key.OrganizationID == "" {
		return nil, ErrMissingOrganizationID
	}
	if key.LocaleID == "" {
		locale, err := b.GetLocaleByFirstMatch(key.Locale)
		if err != nil {
			return &key, fmt.Errorf("failed to lookup locale: %w", err)
		}
		if locale != nil {
			key.LocaleID = locale.ID
		}
	}
	if key.CategoryID == "" {
		// TODO: report missing on sub-categories!!!
		category, err := b.FindOneCategory(types.CategoryFilter{Key: key.Category})
		if err != nil {
			return &key, fmt.Errorf("failed to lookup category: %w", err)
		}
		if category != nil {
			key.CategoryID = category.ID
		}
	}
	if key.TranslationID == "" {
		translation, err := b.GetTranslationFilter(types.Translation{Key: key.Translation})
		if err != nil {
			return &key, fmt.Errorf("failed to lookup translation: %w", err)
		}
		if translation != nil {
			key.TranslationID = translation.ID
		}
	}
	verb := PubVerbCreate

	err = b.Update(func(tx *bolt.Tx) (err error) {
		bucket := tx.Bucket(BucketMissing)
		existing := bucket.Get([]byte(key.ID))
		var ex types.MissingTranslation
		if existing != nil {
			verb = PubVerbUpdate
			key.LatestUserAgent = key.FirstUserAgent
			key.FirstUserAgent = ex.FirstUserAgent
			err = b.Unmarshal(existing, &ex)
			if err != nil {
				return fmt.Errorf("failed to unmarshal existing key: %w", err)
			}
			key.Count = ex.Count + 1
			if key.ProjectID == "" {
				key.ProjectID = ex.ProjectID
			}
			if key.CategoryID == "" {
				key.CategoryID = ex.CategoryID
			}
			if key.TranslationID == "" {
				key.TranslationID = ex.TranslationID
			}

		}

		bytes, err := b.Marshal(key)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(key.ID), bytes)
	})
	if err != nil {
		return &key, err
	}

	b.PublishChange(PubTypeMissingTranslation, verb, key)
	return &key, err
}

// GetMissingKeys(filter ...MissingTranslation) (map[string]MissingTranslation, error)

func (bb *BBolter) GetMissingKeysFilter(max int, filter ...types.MissingTranslation) (map[string]types.MissingTranslation, error) {
	mts := make(map[string]types.MissingTranslation)
	length := 0
	err := bb.Iterate(BucketMissing, func(key, b []byte) bool {
		var mt types.MissingTranslation
		err := bb.Unmarshal(b, &mt)
		if err != nil {
			bb.l.Error().Err(err).Msg("failed to unmarshal missing-translation")
			return false
		}
		// if len(filter) == 0 {

		// }
		found := len(filter) == 0
		for _, f := range filter {
			if f.ProjectID != "" && f.ProjectID != mt.ProjectID {
				continue
			}
			if f.CategoryID != "" && f.CategoryID != mt.CategoryID {
				continue
			}
			if f.TranslationID != "" && f.TranslationID != mt.TranslationID {
				continue
			}
			if f.Project != "" && f.Project != mt.Project {
				continue
			}
			if f.Category != "" && f.Category != mt.Category {
				continue
			}
			if f.Translation != "" && f.Translation != mt.Translation {
				continue
			}
			found = true
			break
		}
		if found {
			mts[mt.ID] = mt
			length++
			if max == 0 {
				return false
			}
			return length >= max
		}
		return false
	})
	if err == ErrNotFound {
		err = nil
	}
	if err != nil {
		return mts, err
	}
	return mts, err
}
