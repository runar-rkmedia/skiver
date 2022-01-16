package types

import "fmt"

type userSeeder interface {
	CreateOrganization(org Organization) (Organization, error)
	GetOrganizations() (map[string]Organization, error)

	CreateUser(user User) (User, error)
	GetUserByUserName(userName string) (*User, error)
}

// Creates users and the initial organization.
func SeedUsers(db userSeeder, users []User, pwHasher func(s string) ([]byte, error)) (*Organization, error) {
	initialOrgTitle := "Initial organization"
	orgs, err := db.GetOrganizations()
	if err != nil {
		return nil, err
	}
	if len(orgs) > 0 {
		for _, v := range orgs {
			if v.Title == initialOrgTitle {
				return &v, nil
			}
			return nil, nil
		}
	}
	org, err := db.CreateOrganization(Organization{Title: initialOrgTitle, CreatedBy: "seeder"})
	if err != nil {
		return &org, err
	}

	organizationID := org.ID
	if users == nil || len(users) == 0 {
		users = []User{{
			UserName:              "admin",
			Active:                true,
			Store:                 UserStoreLocal,
			CanCreateOrganization: true,
			CanCreateUsers:        true,
		}}
	}
	if v, _ := db.GetUserByUserName(users[0].UserName); v != nil {
		return &org, nil
	}
	for i := 0; i < len(users); i++ {
		if users[i].CreatedBy == "" {
			users[i].CreatedBy = "seeder"
		}
		if users[i].OrganizationID == "" {
			users[i].OrganizationID = organizationID
		}
		if users[i].PW == nil {
			pw, err := pwHasher(users[i].UserName)
			if err != nil {
				return &org, fmt.Errorf("failed to issue password for user")
			}
			users[i].PW = pw
		}
		if _, err := db.CreateUser(users[i]); err != nil {
			return &org, fmt.Errorf("failed to create user %s: %w", users[i].UserName, err)
		}
		users[i].PW = nil
	}
	return &org, nil
}
