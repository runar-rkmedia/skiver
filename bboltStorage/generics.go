package bboltStorage

import (
	"errors"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

type Datastore interface {
	GetItem([]byte, string, interface{}) error
	Iterate(bucket []byte, f func(key, b []byte) bool) error
	// iterates over objects within a bucket.
	// The function-parameter will receive each item as key/value
	// Returning true within this function will stop the iteration
	Marshaller
	PublishChange(kind PubType, variant PubVerb, contents interface{})
}

// Generics function returning a single item, by its id
func Get[T Identifyable](bb Datastore, bucket []byte, id string) (*T, error) {
	var j T
	err := bb.GetItem(bucket, id, &j)
	// TODO: remove the ErrNotFound from bb.GetItem
	if err == ErrNotFound {
		return nil, nil
	}
	// sanity-check:
	if j.IDString() == "" {
		return nil, nil
	}
	return &j, err
}

// Returning a list of items, up to max-count.
// The in-function should return a boolean indicating if that item
// should be added to the map or not.
func Find[T Identifyable](bb Datastore, bucket []byte, max int, shouldAdd func(t T) bool) (map[string]T, error) {
	items := map[string]T{}
	len := 0
	var innerErr error
	err := bb.Iterate(bucket, func(key, b []byte) bool {
		var j T
		err := bb.Unmarshal(b, &j)
		if err != nil {
			innerErr = err
			return false
		}
		if shouldAdd(j) {
			items[string(key)] = j
			len++
			if max == 0 {
				return false
			}
			return len >= max
		}
		return false
	})
	if innerErr != nil {
		if err != nil {
			return items, fmt.Errorf("innerErr: %w, err: %s", innerErr, err)
		}
		return items, innerErr
	}
	return items, err
}

// Returns a single item by iterating. Mostly used for searching
// If you are attempting to find by ID, see Get instead
func FindOne[T Identifyable](bb Datastore, bucket []byte, isMatch func(t T) bool) (*T, error) {
	var t *T
	var innerErr error
	err := bb.Iterate(bucket, func(key, b []byte) bool {
		var j T
		err := bb.Unmarshal(b, &j)
		if err != nil {
			innerErr = err
			return false
		}
		if isMatch(j) {
			t = &j
			return true
		}
		return false
	})
	if innerErr != nil {
		if err != nil {
			return t, fmt.Errorf("innerErr: %w, err: %s", innerErr, err)
		}
		return t, innerErr
	}
	return t, err
}

type Identifyable interface {
	IDString() string
	Namespace() string
	Kind() string
}

// Creates an item in the assigned bucket
func Create[T Identifyable](bb *BBolter, bucket []byte, item T) error {
	err := bb.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucket)
		id := []byte(item.IDString())
		existing := bucket.Get(id)
		if existing != nil {
			return fmt.Errorf("there already exists an item with this ID")
		}
		bytes, err := bb.Marshal(item)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(id), bytes)
	})
	if err != nil {
		return err
	}
	bb.PublishChange(PubType(item.Kind()), PubVerbCreate, item)
	return nil
}

// Updates an item in the assigned bucket
func Update[T Identifyable](bb *BBolter, bucket []byte, id string, merge func(t T) (T, error)) (T, error) {
	var t T
	err := bb.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucket)
		idb := []byte(id)
		existing := bucket.Get(idb)
		if existing == nil {
			return ErrNotFound
		}
		err := bb.Unmarshal(existing, &t)
		if err != nil {
			return err
		}

		// Sanity check
		if t.IDString() != id {
			return ErrIDStringMismatch
		}
		item, err := merge(t)
		if err != nil {
			return err
		}
		t = item
		bytes, err := bb.Marshal(item)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(idb), bytes)
	})
	if err != nil {
		return t, err
	}
	bb.PublishChange(PubType(t.Kind()), PubVerbUpdate, t)
	return t, err
}

var (
	ErrNoFieldsChanged  = errors.New("No fields changed")
	ErrIDStringMismatch = errors.New("The IDString did not match the expected id. (Marshalling error?, wrong type?)")
)
