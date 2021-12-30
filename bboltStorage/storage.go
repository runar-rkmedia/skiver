package bboltStorage

import (
	"errors"
	"os"
	"time"

	"github.com/runar-rkmedia/gabyoall/logger"
	"github.com/runar-rkmedia/skiver/types"
	"go.etcd.io/bbolt"
	bolt "go.etcd.io/bbolt"
)

var (
	ErrMissingIdArg = errors.New("Missing id as argument")
	// deprecated. return pointer instead
	ErrNotFound      = errors.New("Not found")
	ErrMissingBucket = errors.New("Bucket not found")
)

type PubSubPublisher interface {
	Publish(kind, variant string, contents interface{})
}

// Caller must call close when ending
func NewBbolt(l logger.AppLogger, path string, pubsub PubSubPublisher) (bb BBolter, err error) {

	bb.l = l
	db, err := bolt.Open(path, 0666, &bolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		return
	}
	bb.DB = db
	bb.pubsub = pubsub
	bb.Marshaller = Gob{}
	err = bb.Update(func(t *bolt.Tx) error {
		buckets := [][]byte{BucketUser, BucketLocale, BucketTranslation, BucketProject, BucketSession, BucketCategory, BucketTranslationValue}
		for i := 0; i < len(buckets); i++ {
			_, err := t.CreateBucketIfNotExists(buckets[i])
			if err != nil {
				return err

			}
		}
		return nil
	})
	return
}

func (s *BBolter) PublishChange(kind PubType, variant PubVerb, contents interface{}) {
	if s.pubsub == nil {
		return
	}
	if s.l.HasDebug() {
		s.l.Debug().Str("kind", string(kind)).Str("variant", string(variant)).Msg("Entity-change")
	}
	s.pubsub.Publish(string(kind), string(variant), contents)
}

func (s *BBolter) GetItem(bucket []byte, id string, j interface{}) error {
	if id == "" {
		return ErrMissingIdArg
	}
	err := s.DB.View(func(t *bolt.Tx) error {
		bucket := t.Bucket(bucket)
		b := bucket.Get([]byte(id))
		if b == nil || len(b) == 0 {
			return ErrNotFound
		}
		return s.Unmarshal(b, j)
	})
	if err != nil {
		s.l.Error().Err(err).Bytes("bucket", bucket).Str("id", id).Msg("Failed to lookup item")
	}

	return err
}
func (s *BBolter) NewEntity() types.Entity {
	// ForceNewEntity may return an error, but it guarantees the the Entity is still usable.
	// The error should be logged, though.
	e, err := ForceNewEntity()
	if err != nil {
		s.l.Error().Err(err).Str("id", e.ID).Msg("An error occured while creating entity. ")
	}
	return e
}

func (s *BBolter) Size() (int64, error) {
	s.l.Info().Interface("stats", s.Stats()).Msg("DB-stats")

	stat, err := os.Stat(s.Path())
	if err != nil {
		return 0, err
	}
	return int64(stat.Size()), err
}
func (s *BBolter) updater(id string, bucket []byte, f func(b []byte) ([]byte, error)) error {
	if id == "" {
		return ErrMissingIdArg
	}
	if bucket == nil {
		return ErrMissingIdArg
	}
	err := s.Update((func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucket)
		b := bucket.Get([]byte(id))
		if len(b) == 0 {
			return ErrNotFound
		}
		newBytes, err := f(b)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(id), newBytes)
	}))

	return err
}

// iterates over objects within a bucket.
// The function-parameter is will receive each item as key/value
// Returning true within this function will stop the iteration
func (bb *BBolter) Iterate(bucket []byte, f func(key, b []byte) bool) error {
	err := bb.View(func(t *bbolt.Tx) error {
		b := t.Bucket(bucket)
		if b == nil {
			return ErrNotFound
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if f(k, v) {
				return nil
			}
		}

		return ErrNotFound

	})
	return err
}

type BBolter struct {
	*bolt.DB
	pubsub PubSubPublisher
	l      logger.AppLogger
	Marshaller
}

var (
	BucketSession          = []byte("sessions")
	BucketUser             = []byte("users")
	BucketLocale           = []byte("locales")
	BucketTranslation      = []byte("translations")
	BucketProject          = []byte("projects")
	BucketTranslationValue = []byte("translationValues")
	BucketCategory         = []byte("categories")
)
