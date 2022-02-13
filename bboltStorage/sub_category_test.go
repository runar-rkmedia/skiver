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

	root, err := db.CreateCategory(newBaseCategoryFromProject(base, ""))
	testza.AssertNoError(t, err)
	testza.AssertNotEqual(t, root.ID, "")
	testza.AssertTrue(t, root.IsRoot())

	_, err = db.CreateCategory(newBaseCategoryFromProject(base, ""))
	testza.AssertErrorIs(t, err, ErrDuplicate, "should fail on attempting to insert a root-category into a root-category")

	// Create a sub-category at the root-level.
	mockTime.Tick()
	baseGeneral := newBaseCategoryFromProject(base, "General")
	catGeneral, err := db.CreateCategory(baseGeneral)
	testza.AssertNoError(t, err)
	testza.AssertNotEqual(t, catGeneral.ID, "")
	testza.AssertEqual(t, catGeneral.Path(), []string{"General"})

	// Create a sub-category the the second level, attached to the subcategory General.
	mockTime.Tick()
	baseForms := newBaseCategoryFromProject(base, "General.Forms")
	catForms, err := db.CreateCategory(baseForms)

	testza.AssertNoError(t, err)
	testza.AssertNotEqual(t, catForms.ID, "")
	testza.AssertEqual(t, catForms.Path(), []string{"General", "Forms"})

	// Create another sub-category the the second level, attached to the subcategory General.
	mockTime.Tick()

	baseButtons := newBaseCategoryFromProject(base, "General.Buttons")
	catButtons, err := db.CreateCategory(baseButtons)

	testza.AssertNoError(t, err)
	testza.AssertNotEqual(t, catButtons.ID, "")

	// Create a translation
	tr, err := db.CreateTranslation(types.Translation{
		Entity:     base.Entity,
		Key:        "My translation",
		CategoryID: catButtons.ID,
	})
	testza.AssertNoError(t, err)
	testza.AssertNotEqual(t, tr.ID, "")

	proj, err := db.GetProject(project.ID)
	testza.AssertNoError(t, err)
	p, err := proj.Extend(&db)
	internal.PrintMultiLineYaml("", p.CategoryTree)
	testza.AssertNoError(t, err)
	testza.AssertEqual(t, []string{}, p.CategoryTree.Path())
	testza.AssertEqual(t, []string{"General"}, p.CategoryTree.Categories["General"].Path())
	testza.AssertEqual(t, []string{"General", "Forms"}, p.CategoryTree.Categories["General"].Categories["Forms"].Path())
	testza.AssertEqual(t, []string{"General", "Buttons"}, p.CategoryTree.Categories["General"].Categories["Buttons"].Path())

	// Create sub-category the the third level, attached to the subcategory Forms.
	mockTime.Tick()

	baseLabels := newBaseCategoryFromProject(base, "General.Forms.Labels")
	catLabels, err := db.CreateCategory(baseLabels)

	testza.AssertNoError(t, err)
	testza.AssertNotEqual(t, catLabels.ID, "")
	proj, err = db.GetProject(project.ID)
	testza.AssertNoError(t, err)
	p, err = proj.Extend(&db)
	internal.PrintMultiLineYaml("", p.CategoryTree)
	testza.AssertEqual(t, p.CategoryTree.Key, "")
	testza.AssertEqual(t, []string{}, p.CategoryTree.Path())
	testza.AssertEqual(t, []string{"General"}, p.CategoryTree.Categories["General"].Path())
	testza.AssertEqual(t, []string{"General", "Forms"}, p.CategoryTree.Categories["General"].Categories["Forms"].Path())
	testza.AssertEqual(t, []string{"General", "Buttons"}, p.CategoryTree.Categories["General"].Categories["Buttons"].Path())
	testza.AssertEqual(t, []string{"General", "Forms", "Labels"}, p.CategoryTree.Categories["General"].Categories["Forms"].Categories["Labels"].Path())
	internal.MatchSnapshot(t, ".yml", p)
}

func newBaseCategoryFromProject(p types.Project, key string) types.Category {
	c := types.Category{}
	c.CreatedBy = p.CreatedBy
	c.ProjectID = p.ID
	c.OrganizationID = p.OrganizationID
	c.Key = key
	return c
}
