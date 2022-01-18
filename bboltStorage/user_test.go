package bboltStorage

import (
	"io/ioutil"
	"log"
	"os"
	"sort"
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/runar-rkmedia/gabyoall/logger"
	"github.com/runar-rkmedia/skiver/types"
)

func TestUser(t *testing.T) {
	t.Run("Basic test", func(t *testing.T) {
		tmpFile, err := ioutil.TempFile(os.TempDir(), "prefix-")
		if err != nil {
			log.Fatal("Cannot create temporary file", err)
		}
		defer os.Remove(tmpFile.Name())

		l := logger.GetLogger("test")

		bb, err := NewBbolt(l, tmpFile.Name(), nil)
		testza.AssertNoError(t, err, "bb-bolt-creation should not err")

		// Ensure there are no users
		users, err := bb.FindUsers(0)
		testza.AssertNoError(t, err, "should not err on getting users")
		testza.AssertEqual(t, 0, len(users), "user-count should be nil at start-up")
		//
		org, err := types.SeedUsers(&bb, nil, func(s string) ([]byte, error) { return []byte("mock-" + s), nil })
		testza.AssertNoError(t, err)
		testza.AssertNotNil(t, org)
		testza.AssertEqual(t, "Initial organization", org.Title)

		// Test that all required fields are checked for
		input := types.User{UserName: "Rockman"}
		_, err = bb.CreateUser(input)
		if err == nil {
			t.Error("Expected err during create")
			return
		}
		testza.AssertEqual(t, err.Error(), "store must be set")
		input.Store = types.UserStoreLocal

		_, err = bb.CreateUser(input)
		if err == nil {
			t.Error("Expected err during create")
			return
		}
		testza.AssertEqual(t, err.Error(), "password must be set")
		input.PW = []byte("hashy-pass")

		_, err = bb.CreateUser(input)
		if err == nil {
			t.Error("Expected err during create")
			return
		}
		testza.AssertEqual(t, err.Error(), "OrganizationID was empty")
		input.OrganizationID = "org-123"

		_, err = bb.CreateUser(input)
		if err == nil {
			t.Error("Expected err during create")
			return
		}
		testza.AssertEqual(t, err.Error(), "CreatedBy was empty")
		input.CreatedBy = "user-123"

		// Actually create user
		u, err := bb.CreateUser(input)
		testza.AssertNoError(t, err, "should not err on creating user")
		if u.ID == "" {
			t.Error("User-id should have been set")
			return
		}
		// Create another user
		other, err := bb.CreateUser(types.User{
			Entity:   types.Entity{CreatedBy: "test", OrganizationID: "org-b"},
			Store:    types.UserStoreLocal,
			PW:       []byte("fake-hash"),
			UserName: "Rock",
		})
		testza.AssertNoError(t, err, "should not err on creating user")
		if other.ID == "" {
			t.Error("User-id should have been set")
			return
		}

		nonExistant, err := bb.GetUser("NonExistant")
		testza.AssertNoError(t, err, "should be nil of non-existant-user")
		testza.AssertNoError(t, err)
		testza.AssertNil(t, nonExistant)
		u2, err := bb.GetUser(u.ID)
		testza.AssertNoError(t, err, "should not err on getting user")
		testza.AssertEqual(t, u2.UserName, input.UserName, "username of returned username should match input")
		_, err = bb.CreateUser(input)
		if err == nil {
			t.Error("Expected recreating the same user to fail")
			return
		}
		testza.AssertEqual(t, err.Error(), "Username is taken")

		// Update user
		_, err = bb.UpdateUser(u.ID, types.User{UserName: "Mega Man"})
		testza.AssertEqual(t, err.Error(), "UpdatedBy is not set")

		u3, err := bb.UpdateUser(u.ID, types.User{UserName: "Mega Man", Entity: types.Entity{UpdatedBy: "Test"}})
		testza.AssertNoError(t, err, "should not err on updating user")
		testza.AssertEqual(t, u3.UserName, "Mega Man")

		u4, err := bb.GetUser(u.ID)
		testza.AssertNoError(t, err, "should not err on getting user")
		testza.AssertEqual(t, u4.UserName, "Mega Man", "username of returned user should have been changed")

		// Check all users
		{
			allUsers, err := bb.FindUsers(0)
			testza.AssertNoError(t, err)
			testza.AssertEqual(t, 3, len(allUsers))
			var uNames []string
			for _, v := range allUsers {
				uNames = append(uNames, v.UserName)
			}
			sort.Strings(uNames)
			testza.AssertEqual(t, []string{"Mega Man", "Rock", "admin"}, uNames)
		}
		// Find some users, by using an or-filter
		{
			allUsers, err := bb.FindUsers(0, types.User{UserName: "Mega Man"}, types.User{UserName: "Rock"})
			testza.AssertNoError(t, err)
			testza.AssertEqual(t, 2, len(allUsers))
			var uNames []string
			for _, v := range allUsers {
				uNames = append(uNames, v.UserName)
			}
			sort.Strings(uNames)
			testza.AssertEqual(t, []string{"Mega Man", "Rock"}, uNames)
		}

		// Run various tests to filter users
		{
			mUsers, err := bb.FindUsers(1)
			testza.AssertNoError(t, err)
			testza.AssertEqual(t, 1, len(mUsers))
		}
		{
			mUsers, err := bb.FindUsers(0, types.User{UserName: "Mega Man"})
			testza.AssertNoError(t, err)
			testza.AssertEqual(t, 1, len(mUsers))
			for _, v := range mUsers {
				testza.AssertEqual(t, v.UserName, "Mega Man")
			}
		}

		{
			mUsers, err := bb.FindUsers(0, types.User{UserName: "admin"})
			testza.AssertNoError(t, err)
			testza.AssertEqual(t, 1, len(mUsers))
			for _, v := range mUsers {
				testza.AssertEqual(t, v.UserName, "admin")
			}
		}
		{
			mUsers, err := bb.FindUsers(0, types.User{UserName: "Sigma"})
			testza.AssertNoError(t, err)
			testza.AssertEqual(t, 0, len(mUsers))
		}

	})

}
