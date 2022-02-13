<script lang="ts">
  import { db } from 'api'
  import { state } from 'state'
  import Button from 'components/Button.svelte'
  import Collapse from 'components/Collapse.svelte'
  export let projectID: string
  import sortOn from 'sort-on'
  import { t } from '../util/i18next'

  import CategoryForm from 'forms/CategoryForm.svelte'
  import CategoryList from '../components/CategoryList.svelte'
  import EntityDetails from 'components/EntityDetails.svelte'

  $: project = $db.project[projectID]
  $: {
    if (!$state.projectSettings[projectID]?.localeIds) {
      $state.projectSettings[projectID] = {
        localeIds: Object.keys($db.locale || {}),
      }
    }
  }
  $: locales = $state.projectSettings[projectID]?.localeIds?.length
    ? $state.projectSettings[projectID].localeIds.map((id) => $db.locale[id])
    : Object.values($db.locale || {})
  $: categories =
    (project &&
      (project.category_ids || [])
        .map((cid) => $db.category[cid])
        .filter(Boolean)) ||
    []
  let visibleForm: null | 'translation' | 'category' | 'translationValue' = null
  let selectedCategory = ''
  let selectedTranslation = ''
  let selectedLocale = ''
  const categorySortOptions: Array<keyof ApiDef.Category> = [
    'key',
    'title',
    'createdAt',
    'updatedAt',
  ]
</script>

{#if !project}
  {#if $db.responseStates.project.loading}
    Loading...
  {:else if $db.responseStates.project.error}
    {$db.responseStates.project.error.error}
  {:else}
    Project not found: {projectID}
  {/if}
{:else}
  <h2>{project.title}</h2>
  <p>{project.description}</p>
  <EntityDetails entity={project} />
  {#if $state.projectSettings[projectID]?.localeIds}
    <paper>
      <Collapse key={'ps-' + projectID}>
        <h3 slot="title">Project-settings</h3>
        <h4>Only show these locales</h4>
        {#each Object.values($db.locale) as locale}
          <div>
            <label>
              <input
                type="checkbox"
                name="locale-ids"
                value={locale.id}
                bind:group={$state.projectSettings[projectID].localeIds} />

              {locale.title}
            </label>
          </div>
        {/each}
        <!-- svelte-ignore a11y-missing-content -->
        <a href={`/api/export/p=${project.short_name || project.id}`}>
          Exported raw
        </a>
      </Collapse>
    </paper>
  {/if}
  <div>
    <h2>Categories</h2>
    <label>
      Sort by: {$state.categorySortOn}
      <select bind:value={$state.categorySortOn}>
        {#each categorySortOptions as option}
          <option value={option}>{option}</option>
        {/each}
      </select>
    </label>
    <label>
      <input bind:checked={$state.categorySortAsc} type="checkbox" />
      Ascending
    </label>
    <Button
      color="secondary"
      icon={'create'}
      on:click={() => {
        visibleForm = 'category'
      }}>Create category</Button>
    {#if visibleForm === 'category'}
      <paper>
        <CategoryForm on:complete={() => (visibleForm = null)} {projectID}>
          <Button
            slot="actions"
            color="secondary"
            icon={'toggleOff'}
            on:click={() => (visibleForm = null)}>
            Cancel
          </Button>
        </CategoryForm>
      </paper>
    {/if}
  </div>
  {#if categories}
    <CategoryList
      {locales}
      {selectedCategory}
      {selectedTranslation}
      {selectedLocale}
      {visibleForm}
      projectKey={project.short_name || project.id}
      categories={sortOn(
        categories,
        ($state.categorySortAsc ? '' : '-') + $state.categorySortOn
      )} />
  {/if}
{/if}
