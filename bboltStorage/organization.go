package bboltStorage

import (
	"fmt"

	"github.com/runar-rkmedia/skiver/types"
	bolt "go.etcd.io/bbolt"
)

func (b *BBolter) GetOrganization(ID string) (*types.Organization, error) {
	var u types.Organization
	err := b.GetItem(BucketOrganization, ID, &u)
	return &u, err
}

func (b *BBolter) CreateOrganization(organization types.Organization) (types.Organization, error) {
	if organization.Title == "" {
		return organization, fmt.Errorf("Missing title: %w", ErrMissingIdArg)
	}
	if organization.CreatedBy == "" {
		return organization, fmt.Errorf("Missing createdBy: %w", ErrMissingIdArg)
	}
	orgs, err := b.GetOrganizations()
	if err != nil {
		return organization, err
	}
	for _, v := range orgs {
		if v.Title == organization.Title {
			return v, fmt.Errorf("Organizations already exists with title: %s", organization.Title)
		}
	}
	_ent := b.newEntity()
	organization.ID = _ent.ID
	organization.CreatedAt = _ent.CreatedAt
	if err != nil {
		return organization, err
	}

	err = b.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(BucketOrganization)
		existing := bucket.Get([]byte(organization.ID))
		if existing != nil {
			return fmt.Errorf("there already exists a organization with this ID")
		}
		bytes, err := b.Marshal(organization)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(organization.ID), bytes)
	})
	if err != nil {
		return organization, err
	}

	b.PublishChange(PubTypeOrganization, PubVerbCreate, organization)
	return organization, err
}

func orgFilter(f, uu types.Organization) bool {
	if f.Title != "" && f.Title != uu.Title {
		return false
	}
	if f.ID != "" && f.ID != uu.ID {
		return false
	}
	return true
}

func (bb *BBolter) FindOrganizationByIdOrTitle(titleOrID string) (*types.Organization, error) {
	return bb.FindOneOrganization(types.Organization{ID: titleOrID}, types.Organization{Title: titleOrID})
}
func (bb *BBolter) FindOneOrganization(filter ...types.Organization) (*types.Organization, error) {
	return FindOne(bb, BucketOrganization, func(t types.Organization) bool {
		for _, f := range filter {
			if orgFilter(f, t) {
				return true
			}
		}
		return false
	})
}

func (bb *BBolter) GetOrganizations() (map[string]types.Organization, error) {
	us := make(map[string]types.Organization)
	err := bb.Iterate(BucketOrganization, func(key, b []byte) bool {
		var u types.Organization
		bb.Unmarshal(b, &u)
		us[string(key)] = u
		return false
	})
	if err == ErrNotFound {
		return us, nil
	}
	return us, err
}

func (bb *BBolter) UpdateOrganization(id string, payload types.UpdateOrganizationPayload) (types.Organization, error) {
	if payload.UpdatedBy == "" {
		return types.Organization{}, fmt.Errorf("missing updatedBy")
	}
	return Update(bb, BucketOrganization, id, func(t types.Organization) (types.Organization, error) {
		shouldUpdate := false
		if payload.JoinID != nil && string(*payload.JoinID) != string(t.JoinID) {
			t.JoinID = *payload.JoinID
			shouldUpdate = true
		}
		if payload.JoinIDExpires != nil && *payload.JoinIDExpires != t.JoinIDExpires {
			t.JoinIDExpires = *payload.JoinIDExpires
			shouldUpdate = true
		}
		if !shouldUpdate {
			return t, ErrNoFieldsChanged
		}
		t.UpdatedAt = nowPointer()
		t.UpdatedBy = payload.UpdatedBy

		return t, nil
	})
}
