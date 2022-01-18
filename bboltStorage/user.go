package bboltStorage

import (
	"fmt"

	"github.com/runar-rkmedia/skiver/types"
)

func (bb *BBolter) GetUser(userId string) (*types.User, error) {
	return Get[types.User](bb, BucketUser, userId)
}
func (bb *BBolter) FindUsers(max int, filter ...types.User) (map[string]types.User, error) {
	return Find(bb, BucketUser, max, func(uu types.User) bool {
		if len(filter) == 0 {
			return true
		}
		for _, f := range filter {
			if userFilter(f, uu) {
				return true
			}
		}
		return false
	})
}

func userFilter(f, uu types.User) bool {
	if f.OrganizationID != "" && f.OrganizationID != uu.OrganizationID {
		return false
	}
	if f.UserName != "" && f.UserName != uu.UserName {
		return false
	}
	if f.ID != "" && f.ID != uu.ID {
		return false
	}
	return true
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
	if v, err := b.FindUserByUserName(user.OrganizationID, user.UserName); err != nil {
		return user, err
	} else if v != nil {
		return user, fmt.Errorf("Username is taken")
	}

	entity, err := b.NewEntity(user.Entity)
	if err != nil {
		return user, err
	}
	user.Entity = entity

	err = Create(b, BucketUser, user)
	return user, err
}

func (bb *BBolter) FindUserByUserName(organizationID string, userName string) (*types.User, error) {
	return bb.FindOne(types.User{UserName: userName, Entity: types.Entity{OrganizationID: organizationID}})
}
func (bb *BBolter) FindOne(filter ...types.User) (*types.User, error) {
	return FindOne(bb, BucketUser, func(t types.User) bool {
		for _, f := range filter {
			if f.OrganizationID == "" {
				bb.l.Warn().Msg("Received a user-filter without organization-id")
			}
			if userFilter(f, t) {
				return true
			}
		}
		return false
	})
}

func (bb *BBolter) UpdateUser(id string, payload types.User) (types.User, error) {
	return Update(bb, BucketUser, id, func(t types.User) (types.User, error) {
		shouldUpdate := false
		if payload.UserName != "" && payload.UserName != t.UserName {
			t.UserName = payload.UserName
			shouldUpdate = true
			e, err := updateEntity(t.Entity, payload.Entity)
			if err != nil {
				return t, err
			}
			t.Entity = e
		}
		if !shouldUpdate {
			return t, ErrNoFieldsChanged
		}

		return t, nil
	})
}
