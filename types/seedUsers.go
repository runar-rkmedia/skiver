package types

import "fmt"

type userSeeder interface {
	CreateUser(user User) (User, error)
	GetUserByUserName(userName string) (*User, error)
}

func SeedUsers(db userSeeder, users []User, pwHasher func(s string) ([]byte, error)) error {
	if users == nil || len(users) == 0 {
		users = []User{{
			UserName: "admin",
			Active:   true,
			Store:    UserStoreLocal,
		}}
	}
	if v, _ := db.GetUserByUserName(users[0].UserName); v != nil {
		return nil
	}
	for i := 0; i < len(users); i++ {
		if users[i].CreatedBy == "" {
			users[i].CreatedBy = "seeder"
		}
		if users[i].PW == nil {
			pw, err := pwHasher(users[i].UserName)
			if err != nil {
				return fmt.Errorf("failed to issue password for user")
			}
			users[i].PW = pw
		}
		if _, err := db.CreateUser(users[i]); err != nil {
			return fmt.Errorf("failed to create user %s: %w", users[i].UserName, err)
		}
		users[i].PW = nil
	}
	return nil
}
