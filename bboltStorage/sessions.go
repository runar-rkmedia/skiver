package bboltStorage

import (
	"fmt"
	"time"

	"github.com/runar-rkmedia/skiver/types"
	"go.etcd.io/bbolt"
)

func (bb *BBolter) GetSessions() (sess map[string]types.Session, err error) {
	sess = make(map[string]types.Session)
	var toEvict [][]byte
	now := time.Now()
	err = bb.Iterate(BucketSession, func(key, b []byte) bool {
		var j types.Session
		err := bb.Unmarshal(b, &j)
		if err != nil {
			bb.l.Error().Err(err).Msg("failed to unmarshal session")
			return false
		}
		if j.Expires.Before(now) {
			toEvict = append(toEvict, key)
			return false
		}
		sess[string(key)] = j
		return false
	})
	if err != nil {
		if err == ErrNotFound {
			err = nil
			return
		}
		bb.l.Error().Err(err).Msg("Failed to list sessions")
	}
	if len(toEvict) > 0 {
		err = bb.Update(func(tx *bbolt.Tx) error {
			bucket := tx.Bucket(BucketSession)
			for i := 0; i < len(toEvict); i++ {
				bucket.Delete(toEvict[i])
			}
			return nil
		})
	}
	return
}
func (bb *BBolter) CreateSession(key string, session types.Session) (types.Session, error) {
	err := bb.Update(func(tx *bbolt.Tx) error {
		b, err := bb.Marshal(session)
		if err != nil {
			return fmt.Errorf("failed to marshal session: %w", err)
		}
		bucket := tx.Bucket(BucketSession)
		return bucket.Put([]byte(key), b)
	})
	if err != nil {
		bb.l.Error().Err(err).Msg("Failed to create session")
	}
	return session, err
}
func (bb *BBolter) EvictSession(key string) error {
	err := bb.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(BucketSession)
		return bucket.Delete([]byte(key))
	})
	if err != nil {
		bb.l.Error().Err(err).Msg("Failed to evict session")
	}
	return err
}
