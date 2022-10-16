<!-- 
Sorry, but the whole concept of missing translations is a big mess.

Mostly, this comes from the fact that the server does not really trust the user input
of anything that is reported in a `missing-translation`.

It will then try to parse this information, and attempt to map each reported field to
a project, category, translation and locale.

Therefore, any missing translation may or may not map into a valid object.

Still, we should present this information.
-->
<script type="ts">
  import ProjectForm from 'forms/ProjectForm.svelte'

  import { db } from '../api'
  import { fly } from 'svelte/transition'
  import Button from './Button.svelte'
  import ListItem from './ListItem.svelte'
  import { state } from 'state'
  import CategoryForm from 'forms/CategoryForm.svelte'
  import TranslationForm from 'forms/TranslationForm.svelte'
  import TranslationItemLegacy from './TranslationItemLegacy.svelte'
  export let projectID = ''
  // import { t } from '../util/i18next'
  $: projectKeyMap = Object.values($db.missingTranslation).reduce((r, m) => {
    if (!m.project) {
      return r
    }
    if (r[m.project]) {
      return r
    }
    if (!m.project_id) {
      return r
    }
    r[m.project] = $db.project[m.project_id] || { id: m.project_id }

    return r
  }, {} as Record<string, ApiDef.Project>)
  $: categories = Object.values($db.missingTranslation).reduce((r, m) => {
    if (!m.category) {
      return r
    }
    if (r[m.category]) {
      return r
    }
    if (!m.category_id) {
      return r
    }
    r[m.category] = $db.category[m.category_id] || { id: m.category_id }

    return r
  }, {} as Record<string, ApiDef.Category>)
  $: missings = Object.values($db.missingTranslation).reduce((r, m) => {
    if (projectID && m.project_id !== projectID) {
      return r
    }
    if (!m.project) {
      return r
    }
    if (!m.category) {
      return r
    }
    r[m.project] = {
      ...r[m.project],
      [m.category]: [...(r[m.project]?.[m.category] || []), m],
    }
    return r
  }, {} as Record<string, Record<string, ApiDef.MissingTranslation[]>>)
  $: {
  }
  let visibleForm:
    | null
    | 'project'
    | 'category'
    | 'translation'
    | 'translationValue' = null
  let selectedProjectIsh = ''
  let selectedCategoryIsh = ''
  let selectedTranslationIsh = ''
</script>

{#each Object.entries(missings) as [projectIsh, projectCategories]}
  <paper>
    {#if !projectID}
      {#if projectKeyMap[projectIsh]}
        <h3>
          Project: <a href={`#project/${projectKeyMap[projectIsh].id}`}>
            {#if projectKeyMap[projectIsh].title}
              {projectKeyMap[projectIsh].title}
              <code>
                {projectKeyMap[projectIsh].short_name}
              </code>
            {:else}
              {projectIsh}
            {/if}
          </a>
        </h3>
      {:else}
        <p>
          The Project '{projectIsh}' does not exist.
        </p>

        {#if visibleForm === 'project' && selectedProjectIsh === projectIsh}
          <paper in:fly|local>
            <ProjectForm shortNameReadOnly={true}>
              <Button
                icon="cancel"
                slot="actions"
                color="secondary"
                on:click={() => {
                  visibleForm = null
                }}>Cancel</Button>
            </ProjectForm>
          </paper>
        {:else}
          <Button
            icon="create"
            color="primary"
            on:click={() => {
              $state.createProject.short_name = projectIsh
              selectedProjectIsh = projectIsh
              visibleForm = 'project'
            }}>Create project!</Button>
        {/if}
      {/if}
    {/if}

    {#each Object.entries(projectCategories) as [categoryIsh, missings]}
      {#if categories[categoryIsh]}
        <h4>
          Category:
          {#if categories[categoryIsh].title}
            {categories[categoryIsh].title}
            <code>
              {categories[categoryIsh].key}
            </code>
          {:else}
            {categoryIsh}
          {/if}
        </h4>
      {:else}
        <p>
          The category '{categoryIsh}' does not exist.
        </p>

        {#if visibleForm === 'category' && selectedCategoryIsh === categoryIsh && selectedProjectIsh === projectIsh}
          <paper in:fly|local>
            <CategoryForm
              projectID={projectKeyMap[projectIsh].id}
              on:complete={() => (visibleForm = null)}>
              <Button
                icon="cancel"
                slot="actions"
                color="secondary"
                on:click={() => {
                  visibleForm = null
                }}>Cancel</Button>
            </CategoryForm>
          </paper>
        {:else if projectKeyMap[projectIsh]}
          <Button
            icon="create"
            color="primary"
            on:click={() => {
              $state.createCategory.key = categoryIsh
              selectedCategoryIsh = categoryIsh
              selectedProjectIsh = projectIsh
              visibleForm = 'category'
            }}>Create category!</Button>
        {/if}
      {/if}
      {#each missings as missing}
        {#if missing.translation_id && $db.translation[missing.translation_id]}
          <TranslationItemLegacy
            projectKey={projectIsh}
            translation={$db.translation[missing.translation_id]}
            translationValues={{}}
            categoryKey={categories[categoryIsh]?.key || categoryIsh}
            locales={Object.values($db.locale)}
            on:complete={() => {
              visibleForm = null
            }}
            on:showForm={({ detail: { show } }) => {
              if (show) {
                // visibleForm = 'translationValue'
                // selectedTranslation = translation.id
                // selectedLocale = translation.locale_id
                return
              }
              visibleForm = null
            }} />
        {:else}
          <ListItem deleted={!!missing.deleted} ID={missing.id}>
            <span slot="header">
              {missing.category}.{missing.translation}
            </span>
            <span slot="description">
              {missing.locale}
              {$db.locale[missing.locale_id || '']?.title}
              <!-- {$t('missing.description', missing)} -->
              {#if visibleForm === 'translation' && selectedCategoryIsh === categoryIsh && selectedProjectIsh === projectIsh && selectedTranslationIsh === missing.translation}
                <paper in:fly|local>
                  <TranslationForm categoryID={missing.category_id || ''} />
                </paper>
              {:else if projectKeyMap[projectIsh] && categories[categoryIsh] && missing.translation}
                <Button
                  icon="create"
                  color="primary"
                  on:click={() => {
                    $state.createTranslation.key = missing.translation || ''
                    selectedCategoryIsh = categoryIsh
                    selectedProjectIsh = projectIsh
                    selectedTranslationIsh = missing.translation || ''
                    visibleForm = 'translation'
                  }}>Create translation!</Button>
              {/if}
            </span>
          </ListItem>
        {/if}
      {/each}
    {/each}
  </paper>
{/each}
