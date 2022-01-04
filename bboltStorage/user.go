package bboltStorage

import (
	"fmt"

	"github.com/runar-rkmedia/skiver/types"
	bolt "go.etcd.io/bbolt"
)

func (b *BBolter) GetUser(userId string) (types.User, error) {
	var u types.User
	err := b.GetItem(BucketUser, userId, &u)
	return u, err
}

func (b *BBolter) CreateUser(user types.User) (types.User, error) {
	if user.Store == 0 {
		return user, fmt.Errorf("store must be set")
	}
	if user.UserName == "" {
		return user, fmt.Errorf("username must be set")
	}
	if len(user.PW) == 0 {
		return user, fmt.Errorf("password must be set")
	}
	if v, err := b.GetUserByUserName(user.UserName); err != nil {
		return user, err
	} else if v != nil {
		return user, fmt.Errorf("Username is taken")
	}

	user.Entity = b.NewEntity()

	err := b.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(BucketUser)
		existing := bucket.Get([]byte(user.ID))
		if existing != nil {
			return fmt.Errorf("there already exists a user with this ID")
		}
		bytes, err := b.Marshal(user)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(user.ID), bytes)
	})
	if err != nil {
		return user, err
	}

	b.PublishChange(PubTypeUser, PubVerbCreate, user)
	return user, err
}
func (bb *BBolter) GetUserByUserName(userName string) (*types.User, error) {
	var u *types.User
	err := bb.Iterate(BucketUser, func(key, b []byte) bool {
		var uu types.User
		err := bb.Unmarshal(b, &uu)
		if err != nil {
			bb.l.Error().Err(err).Msg("failed to unmarshal user")
			return false
		}
		if uu.UserName == userName {
			u = &uu
			return true
		}
		return false
	})
	if err != nil {
		return nil, err
	}
	return u, err
}
