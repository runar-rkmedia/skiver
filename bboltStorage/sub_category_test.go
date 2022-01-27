package bboltStorage

import (
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/runar-rkmedia/skiver/types"
)

func TestSubCategory(t *testing.T) {

	db := NewMockDB(t)
	testza.AssertNoError(t, db.StandardSeed())

	base := types.Project{Title: "project", ShortName: "p"}
	base.CreatedBy = "jimb"
	base.OrganizationID = "org-abc"
	project, err := db.CreateProject(base)
	testza.AssertNoError(t, err)
	base.ID = project.ID

	// Createing sub-category should fail on
	_, err = db.CreateSubCategory("", "", newBaseCategoryFromProject(base))
	testza.AssertErrorIs(t, err, ErrMissingIdArg)
	_, err = db.CreateSubCategory("abc", "", newBaseCategoryFromProject(base))
	testza.AssertErrorIs(t, err, ErrMissingIdArg)
	_, err = db.CreateSubCategory("", "abc", newBaseCategoryFromProject(base))
	testza.AssertErrorIs(t, err, ErrMissingIdArg)

	root, err := db.CreateCategory(newBaseCategoryFromProject(base))
	t.Log(root.ProjectID)
	testza.AssertNoError(t, err)
	testza.AssertNotEqual(t, root.ID, "")
	testza.AssertEqual(t, root.Key, types.RootCategory)

	_, err = db.CreateSubCategory(root.ID, root.ID, newBaseCategoryFromProject(base))
	testza.AssertErrorIs(t, err, ErrDuplicate)
	baseC := newBaseCategoryFromProject(base)
	baseC.Key = "General"
	cat, err := db.CreateSubCategory(root.ID, root.ID, baseC)
	testza.AssertNoError(t, err)
	testza.AssertNotEqual(t, cat.ID, "")
	testza.AssertNotNil(t, cat.SubCategories)
	testza.AssertEqual(t, baseC, cat.SubCategories)

	t.Fatal("not implemented")
}

func newBaseCategoryFromProject(p types.Project) types.Category {
	c := types.Category{}
	c.CreatedBy = p.CreatedBy
	c.ProjectID = p.ID
	c.OrganizationID = p.OrganizationID
	return c

}
