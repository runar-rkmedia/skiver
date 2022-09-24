package bboltStorage

import (
	"encoding/gob"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/runar-rkmedia/go-common/logger"
	"github.com/runar-rkmedia/go-common/utils"
	"github.com/runar-rkmedia/skiver/types"
	"go.etcd.io/bbolt"
	bolt "go.etcd.io/bbolt"
)

var (
	ErrDuplicate      = errors.New("Duplication of entities is disallowed")
	ErrMissingIdArg   = errors.New("Missing id as argument")
	ErrMissingProject = errors.New("Missing ProjectID as argument")
	// deprecated. return pointer instead
	ErrNotFound      = errors.New("Not found")
	ErrMissingBucket = errors.New("Bucket not found")
)

type PubSubPublisher interface {
	Publish(kind, variant string, contents interface{})
}

type BBoltOptions struct {
	IDGenerator IDGenerator
}

// Caller must call close when ending
func NewBbolt(l logger.AppLogger, path string, pubsub PubSubPublisher, options ...BBoltOptions) (bb BBolter, err error) {
	gob.Register(time.Time{})
	var opts BBoltOptions
	if len(options) > 0 {
		opts = options[0]
		bb.idgenerator = opts.IDGenerator
	}

	bb.l = l
	if l.HasDebug() {
		fileInfo, err := os.Stat(path)
		if err != nil {
			l.Error().Str("db-location", path).Err(err).Msg("Reading database from disk failed...")
		} else {
			l.Debug().Str("db-location", path).Err(err).
				Str("size", humanize.Bytes(uint64(fileInfo.Size()))).
				Str("mode", fileInfo.Mode().String()).
				Msg("Database found on disk")
		}
	}
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
		for i := 0; i < len(allBuckets); i++ {
			_, err := t.CreateBucketIfNotExists(allBuckets[i])
			if err != nil {
				return err

			}
		}
		return nil
	})
	return
}

func (s *BBolter) WriteState() WriteStats {
	s.writeStats.Lock()
	defer s.writeStats.Unlock()
	return s.writeStats.WriteStats
}
func (s *BBolter) PublishChange(kind PubType, variant PubVerb, contents interface{}) {
	s.writeStats.Lock()
	now := time.Now()
	s.writeStats.LastWrite = &now
	s.writeStats.Unlock()

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

var (
	ErrMissingCreatedBy      = errors.New("CreatedBy was empty")
	ErrMissingOrganizationID = errors.New("OrganizationID was empty")
	ErrMissingProjectID      = errors.New("ProjectID was empty")
	ErrMissingTags           = errors.New("Missing tags")
)

func (s *BBolter) newUniqueID() string {
	if s.idgenerator == nil {
		str, err := utils.ForceCreateUniqueId()
		if err != nil {
			s.l.Error().Err(err).Msg("Failed during utils.ForceCreateUniqueId")
		}
		return str
	}
	return s.idgenerator.CreateUniqueID()
}

// Returns an entity for use by database, with id set and createdAt to current time.
// It is guaranteeed to be created correctly, even if it errors.
// The error should be logged.
func (s *BBolter) newEntity() types.Entity {

	return types.Entity{
		ID:        s.newUniqueID(),
		CreatedAt: time.Now(),
	}
}
func (s *BBolter) NewEntity(base types.Entity) (e types.Entity, err error) {

	if base.CreatedBy == "" {
		err = ErrMissingCreatedBy
	}
	if base.OrganizationID == "" {
		err = ErrMissingOrganizationID
	}
	// ForceNewEntity may return an error, but it guarantees the the Entity is still usable.
	// The error should be logged, though.
	e = s.newEntity()
	e.CreatedBy = base.CreatedBy
	e.OrganizationID = base.OrganizationID
	return e, err
}

func nowPointer() *time.Time {
	t := time.Now()
	return &t
}

func (bb *BBolter) Size() (int64, error) {
	bb.l.Info().Interface("stats", bb.Stats()).Msg("DB-stats")

	stat, err := os.Stat(bb.Path())
	if err != nil {
		return 0, err
	}
	return int64(stat.Size()), err
}

// Returns BucketStats for all buckets used
func (bb *BBolter) BucketStats() map[string]interface{} {
	stats := map[string]interface{}{}

	bb.DB.View(func(t *bolt.Tx) error {
		for i := 0; i < len(allBuckets); i++ {
			stats[string(allBuckets[i])] = t.Bucket(allBuckets[i]).Stats()
		}
		return nil
	})
	return stats

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
// The function-parameter will receive each item as key/value
// Returning true within this function will stop the iteration
func (bb *BBolter) Iterate(bucket []byte, f func(key, b []byte) bool) error {
	err := bb.View(func(t *bbolt.Tx) error {
		b := t.Bucket(bucket)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if f(k, v) {
				return nil
			}
		}

		return nil

	})
	return err
}

type writeStats struct {
	WriteStats
	sync.Mutex
}
type WriteStats struct {
	LastWrite *time.Time
}

type BBolter struct {
	*bolt.DB
	pubsub PubSubPublisher
	l      logger.AppLogger
	Marshaller
	idgenerator IDGenerator
	writeStats  writeStats
}

type IDGenerator interface {
	CreateUniqueID() string
}

var (
	BucketSession          = []byte("sessions")
	BucketUser             = []byte("users")
	BucketSys              = []byte("sys")
	BucketLocale           = []byte("locales")
	BucketSnapshot         = []byte("snapshot")
	BucketTranslation      = []byte("translations")
	BucketProject          = []byte("projects")
	BucketOrganization     = []byte("organizations")
	BucketTranslationValue = []byte("translationValues")
	BucketCategory         = []byte("categories")
	BucketMissing          = []byte("missing")
	allBuckets             = [][]byte{
		BucketSession,
		BucketUser,
		BucketLocale,
		BucketSnapshot,
		BucketTranslation,
		BucketProject,
		BucketOrganization,
		BucketTranslationValue,
		BucketCategory,
		BucketMissing,
		BucketSys,
	}
)
