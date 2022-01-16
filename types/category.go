package types

import (
	"fmt"
	"strings"
	"time"
)

// A Category is a "folder" for grouping translation-keys together

// swagger:model Category
type Category struct {
	Entity
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	// If the category-key is "___root___", it should be considered as if there are no categories,
	// but just a flat list of items
	// Will be changed to a string array
	Key            string   `json:"key"`
	ProjectID      string   `json:"project_id,omitempty"`
	TranslationIDs []string `json:"translation_ids,omitempty"`
}

// Splits the key into multiple keys. Root-values are removed
func (c Category) Keys() []string {
	list := []string{}
	if c.Key == RootCategory {
		return list
	}
	return strings.Split(c.Key, ".")
}

const (
	// RootCategories are accessible without a key, but we do need a key.
	// A bit dirty.
	// Deprecated, will simply be an empty path
	RootCategory = ""
)

func (c Category) IsRoot() bool {
	return c.Key == RootCategory
}
func (e Category) Namespace() string {
	return e.Kind()
}
func (e Category) Kind() string {
	return string(PubTypeCategory)
}

// Used to filter and search along with Category.Filter(CategoryFilter)
type CategoryFilter struct {
	OrganizationID string
	Key            string
	ID             string
	ProjectID      string
}

func (cat Category) Path() []string {
	if cat.IsRoot() {
		return []string{}
	}

	return strings.Split(cat.Key, ".")
}
func (cat *Category) Update(payload Category, options ...UpdateEntityOptions) error {

	err := cat.Entity.Update(payload.Entity, options...)
	if err != nil {
		return err
	}
outer:
	for _, v := range payload.TranslationIDs {
		for _, vv := range cat.TranslationIDs {
			if v == vv {
				continue outer
			}
		}
		cat.TranslationIDs = append(cat.TranslationIDs, v)
	}
	if payload.Key != "" {
		cat.Key = payload.Key
	}
	if payload.Title != "" {
		cat.Title = payload.Title
	}
	if payload.Description != "" {
		cat.Description = payload.Description
	}
	if cat.ProjectID != "" {
		cat.ProjectID = payload.ProjectID
	}
	return nil
}

type UpdateEntityOptions struct {
	IgnoreMissingFields   bool
	SkipUpdatingUpdatedAt bool
}

// Updates an existing entity-struct with changes.
// NOTE: This does NOT update the db-value itself., but is meant as a helper-func
func (existing *Entity) Update(changes Entity, options ...UpdateEntityOptions) error {
	var opts UpdateEntityOptions
	if len(options) > 0 {
		opts = options[0]
	}
	if existing.ID == "" {
		return fmt.Errorf("The existing-value does not have an ID.")
	}
	updatedBy := changes.UpdatedBy
	if updatedBy == "" {
		updatedBy = changes.CreatedBy
	}
	if updatedBy == "" {
		if !opts.IgnoreMissingFields {
			return fmt.Errorf("UpdatedBy is not set")

		}
	}

	if changes.UpdatedAt == nil {
		if !opts.SkipUpdatingUpdatedAt {
			now := time.Now()
			changes.UpdatedAt = &now
		}
	}
	existing.UpdatedAt = changes.UpdatedAt
	existing.UpdatedBy = changes.UpdatedBy
	return nil
}

// Used to filter and search
func (cat Category) Filter(f CategoryFilter) bool {
	if f.OrganizationID != "" && f.OrganizationID != cat.OrganizationID {
		return false
	}
	if f.Key != "" && f.Key != cat.Key {
		return false
	}
	if f.ID != "" && f.ID != cat.ID {
		return false
	}
	if f.ProjectID != "" && f.ProjectID != cat.ProjectID {
		return false
	}
	return true
}

func (cat Category) AsUniqueFilter() CategoryFilter {
	return CategoryFilter{
		OrganizationID: cat.OrganizationID,
		Key:            cat.Key,
		ID:             cat.ID,
		ProjectID:      cat.ProjectID,
	}
}
