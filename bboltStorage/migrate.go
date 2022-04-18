package bboltStorage

import (
	"fmt"
	"strings"

	"github.com/runar-rkmedia/skiver/types"
	bolt "go.etcd.io/bbolt"
)

func versionForMigrationPoint(point int) string {
	switch point {
	case 0:
		return `pre v0.5.0`
	case 1:
		return `v0.5.0`
	}

	return ""
}

type migrationHook func(state types.State, wantedMigrationPoint int) error

func (bb *BBolter) Migrate(hooks ...func(state types.State, wantedMigrationPoint int) error) (types.State, error) {
	wantedMigrationPoint := 1
	debug := bb.l.HasDebug()
	state, err := bb.GetState()
	if err != nil {
		return types.State{}, err
	}
	if state == nil {
		state = &types.State{}
	}
	for {
		l := bb.l.With().
			Interface("state", state).
			Int("wanted-migration-point", wantedMigrationPoint).
			Logger()
		if wantedMigrationPoint < state.MigrationPoint {
			return *state, fmt.Errorf("The migration-point is higher than expected. Please update the application to above %s", versionForMigrationPoint(wantedMigrationPoint))
		}

		if wantedMigrationPoint == state.MigrationPoint {
			if debug {
				l.Debug().Msg("Migration-check ok")
			}
			return *state, nil
		}
		if debug {
			l.Debug().Msg("Migrating")
		}
		switch state.MigrationPoint {
		// pre v0.5.0
		// In earlier versions, the Translation had the `References` stored in `Variables`, with a key-prefix of `_refs:`.
		// This was simply done because of priorities/laziness.
		// This moves those keys into the `References`.
		case 0:
			tvs, err := bb.GetTranslations()
			if err != nil {
				return *state, err
			}
			for _, tv := range tvs {
				if len(tv.Variables) == 0 {
					continue
				}
				needsUpdate := false
				for key := range tv.Variables {
					if !strings.HasPrefix(key, "_refs:") {
						continue
					}
					k := strings.TrimPrefix(key, "_refs:")
					if k != "" {
						tv.References = append(tv.References, k)
					}
					delete(tv.Variables, key)
					needsUpdate = true
				}
				if needsUpdate {
					if debug {
						l.Debug().Interface("tv", tv).Msg("Updating translation: Moving refs from Variables to References")
					}
					_, err := Update(bb, BucketTranslation, tv.ID, func(t types.Translation) (types.Translation, error) {
						return tv, nil
					})
					if err != nil {
						return *state, err
					}
				}
			}
		}
		if err != nil {
			return *state, err
		}

		for i, hook := range hooks {
			l.Debug().Int("hook_no", i).Int("total", len(hooks)).Msg("Running hook")
			err := hook(*state, wantedMigrationPoint)
			if err != nil {
				return *state, err
			}
		}
		l.Info().Msg("Migration for point was successful")
		state.MigrationPoint += 1
		s, err := bb.SetState(*state)
		if err != nil {
			return s, err
		}
		state = &s
	}
}

func (bb *BBolter) GetState() (*types.State, error) {
	var j types.State
	err := bb.GetItem(BucketSys, "state", &j)
	// TODO: remove the ErrNotFound from bb.GetItem
	if err == ErrNotFound {
		return nil, nil
	}
	return &j, err
}

func (bb *BBolter) SetState(newState types.State) (types.State, error) {
	err := bb.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(BucketSys)

		bytes, err := bb.Marshal(newState)
		if err != nil {
			return err
		}
		return bucket.Put([]byte("state"), bytes)
	})
	return newState, err
}
