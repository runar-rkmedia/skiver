package bboltStorage

import (
	"testing"

	"github.com/MarvinJWendt/testza"
	"github.com/runar-rkmedia/skiver/internal"
	"github.com/runar-rkmedia/skiver/types"
)

func TestSubCategory(t *testing.T) {

	db := NewMockDB(t)
	mockTime := internal.NewMockTimeNow()
	testza.AssertNoError(t, db.StandardSeed())

	base := types.Project{Title: "project", ShortName: "p"}
	base.CreatedBy = "jimb"
	base.OrganizationID = "org-abc"
	project, err := db.CreateProject(base)
	testza.AssertNoError(t, err)
	base.ID = project.ID

	// Createing sub-category should fail on
	_, err = db.CreateSubCategory("", "", newBaseCategoryFromProject(base, ""))
	testza.AssertErrorIs(t, err, ErrMissingIdArg)
	_, err = db.CreateSubCategory("abc", "", newBaseCategoryFromProject(base, ""))
	testza.AssertErrorIs(t, err, ErrMissingIdArg)
	_, err = db.CreateSubCategory("", "abc", newBaseCategoryFromProject(base, ""))
	testza.AssertErrorIs(t, err, ErrMissingIdArg)

	root, err := db.CreateCategory(newBaseCategoryFromProject(base, ""))
	testza.AssertNoError(t, err)
	testza.AssertNotEqual(t, root.ID, "")
	testza.AssertEqual(t, root.Key, types.RootCategory)

	_, err = db.CreateSubCategory(root.ID, root.ID, newBaseCategoryFromProject(base, ""))
	testza.AssertErrorIs(t, err, ErrDuplicate, "should fail on attempting to insert a root-category into a root-category")

	_, err = db.CreateSubCategory(root.ID, "I dont exist, how sad...", newBaseCategoryFromProject(base, "foobar"))
	testza.AssertErrorIs(t, err, ErrNotFound, "should fail on attempting to insert a category into a non-existant target-category")

	// Create a sub-category at the root-level.
	mockTime.Tick()
	baseGeneral := newBaseCategoryFromProject(base, "General")
	catGeneral, err := db.CreateSubCategory(root.ID, root.ID, baseGeneral)
	testza.AssertNoError(t, err)
	testza.AssertNotEqual(t, catGeneral.ID, "")
	testza.AssertNotNil(t, catGeneral.SubCategories)
	testza.AssertTrue(t, catGeneral.IsRoot())
	testza.AssertEqual(t, catGeneral.SubCategories[0].Key, "General")

	// Create a sub-category the the second level, attached to the subcategory General.
	mockTime.Tick()
	baseForms := newBaseCategoryFromProject(base, "Forms")
	catForms, err := db.CreateSubCategory(root.ID, catGeneral.SubCategories[0].ID, baseForms)

	testza.AssertNoError(t, err)
	testza.AssertNotEqual(t, catForms.ID, "")
	testza.AssertNotNil(t, catForms.SubCategories)
	testza.AssertTrue(t, catForms.IsRoot())
	testza.AssertEqual(t, catForms.SubCategories[0].Key, "General")
	testza.AssertEqual(t, catForms.SubCategories[0].SubCategories[0].Key, "Forms")

	// Create another sub-category the the second level, attached to the subcategory General.
	mockTime.Tick()

	baseButtons := newBaseCategoryFromProject(base, "Buttons")
	catButtons, err := db.CreateSubCategory(root.ID, catGeneral.SubCategories[0].ID, baseButtons)

	testza.AssertNoError(t, err)
	testza.AssertNotEqual(t, catButtons.ID, "")
	testza.AssertNotNil(t, catButtons.SubCategories)
	testza.AssertTrue(t, catButtons.IsRoot())
	testza.AssertEqual(t, catButtons.SubCategories[0].Key, "General")
	testza.AssertEqual(t, catButtons.SubCategories[0].SubCategories[0].Key, "Forms")
	testza.AssertEqual(t, catButtons.SubCategories[0].SubCategories[1].Key, "Buttons")

	// Create sub-category the the third level, attached to the subcategory Forms.
	mockTime.Tick()

	baseLabels := newBaseCategoryFromProject(base, "Labels")
	catLabels, err := db.CreateSubCategory(root.ID, catButtons.SubCategories[0].SubCategories[0].ID, baseLabels)

	testza.AssertNoError(t, err)
	testza.AssertNotEqual(t, catLabels.ID, "")
	testza.AssertNotNil(t, catLabels.SubCategories)
	testza.AssertTrue(t, catLabels.IsRoot())
	testza.AssertEqual(t, catLabels.SubCategories[0].Key, "General")
	testza.AssertEqual(t, catLabels.SubCategories[0].SubCategories[0].Key, "Forms")
	testza.AssertEqual(t, catLabels.SubCategories[0].SubCategories[1].Key, "Buttons")
	testza.AssertEqual(t, catLabels.SubCategories[0].SubCategories[0].SubCategories[0].Key, "Labels")
}

func newBaseCategoryFromProject(p types.Project, key string) types.Category {
	c := types.Category{}
	c.CreatedBy = p.CreatedBy
	c.ProjectID = p.ID
	c.OrganizationID = p.OrganizationID
	c.Key = key
	return c
}
