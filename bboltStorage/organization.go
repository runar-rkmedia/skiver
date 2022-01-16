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
	_ent, err := ForceNewEntity()
	if err != nil {
		b.l.Warn().Err(err).Msg("Failed to create entity")
	}
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
