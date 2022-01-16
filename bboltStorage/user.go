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
	fmt.Println("user??", user.UserName)
	if v, err := b.GetUserByUserName(user.UserName); err != nil {
		fmt.Println("user??1", v, err)
		return user, err
	} else if v != nil {
		return user, fmt.Errorf("Username is taken")
	}

	entity, err := b.NewEntity(user.Entity)
	if err != nil {
		return user, err
	}

	user.Entity = entity

	err = b.Update(func(tx *bolt.Tx) error {
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
	i := 0
	err := bb.Iterate(BucketUser, func(key, b []byte) bool {
		var uu types.User
		err := bb.Unmarshal(b, &uu)

		if err != nil {
			bb.l.Error().Err(err).Msg("failed to unmarshal user")
			return false
		}
		i++
		fmt.Println("user", uu.UserName)
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
func (bb *BBolter) GetUsers(max int, filter ...types.User) (map[string]types.User, error) {
	u := make(map[string]types.User)
	len := 0
	err := bb.Iterate(BucketUser, func(key, b []byte) bool {
		var uu types.User
		err := bb.Unmarshal(b, &uu)
		if err != nil {
			bb.l.Error().Err(err).Msg("failed to unmarshal user")
			return false
		}
		for _, f := range filter {
			if f.OrganizationID != "" && f.OrganizationID != uu.OrganizationID {
				continue
			}
			if f.UserName != "" && f.UserName != uu.UserName {
				continue
			}
			if f.ID != "" && f.ID != uu.ID {
				continue
			}
			u[uu.ID] = uu
			len++
			if max == 0 {
				return false
			}
			return len >= max
		}
		return false
	})
	return u, err
}
