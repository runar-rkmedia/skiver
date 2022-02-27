<script lang="ts">
  import { db } from 'api'
  import { state } from 'state'
  import Button from 'components/Button.svelte'
  import Collapse from 'components/Collapse.svelte'
  export let projectID: string

  import CategoryForm from 'forms/CategoryForm.svelte'
  import CategoryList from '../components/CategoryList.svelte'
  import EntityDetails from 'components/EntityDetails.svelte'
  import ProjectOverview from '../components/ProjectOverview.svelte'
  import GlobalSearch from '../components/GlobalSearch.svelte'
  import { scrollToCategoryByKey } from 'util/scrollToCategory'
  import Icon from 'components/Icon.svelte'

  $: project = $db.project[projectID]
  $: {
    if (!$state.projectSettings[projectID]?.localeIds) {
      $state.projectSettings[projectID] = {
        localeIds: Object.keys($db.locale || {}),
      }
    }
  }
  // Filter the locales, based on if they are activated in the project, and if the user has not ignored them
  $: locales =
    !!project &&
    Object.keys(project.locales || {})
      .filter((id) => {
        const pl = project.locales
        if (!pl) {
          return false
        }
        if (!pl[id].enabled) {
          return false
        }
        const lids = $state?.projectSettings?.[projectID]?.localeIds
        if (lids && lids.length !== pl.length) {
          if (lids.includes?.(id)) {
            return false
          }
        }
        return true
      })
      .filter((id) => {
        return !$state?.projectSettings?.[projectID]?.localeIds?.includes?.(id)
      })
      .map((id) => $db.locale[id])
      .filter(Boolean)
  let visibleForm: null | 'translation' | 'category' | 'translationValue' = null
  let selectedCategory = ''
  let selectedTranslation = ''
  let selectedLocale = ''
  const categorySortOptions: Array<keyof ApiDef.Category> = [
    'key',
    'title',
    'created_at',
    'updated_at',
  ]
</script>

<div class="r-wrapper">
  <div class="wrapper" class:sidebarVisible={$state.sidebarVisible}>
    {#if !project}
      {#if $db.responseStates.project.loading}
        Loading...
      {:else if $db.responseStates.project.error}
        {$db.responseStates.project.error.error}
      {:else}
        Project not found: {projectID}
      {/if}
    {:else}
      <h2>
        {project.title}
        <a href={'#project/' + project.id + '/settings'}>
          <Icon icon="settings" />
          Settings
        </a>
      </h2>

      {#if project.description}
        <p>{project.description}</p>
      {/if}
      <EntityDetails entity={project} />
      {#if $state.projectSettings[projectID]?.localeIds}
        <paper>
          <Collapse key={'ps-' + projectID} let:show>
            <h3 slot="title">Project-settings</h3>
            {#if show}
              <h3>My settings</h3>

              <!-- These settings will be moved to user-settings in the future -->

              <small>(these settings are not shared with others)</small>
              <h4>Locales to hide</h4>
              <p>
                This can be used to hide locales from translation-lists that you
                do not want to see
              </p>
              {#each Object.entries(project.locales || {}).filter(([_, setting]) => setting.enabled) as [id]}
                <div>
                  <label>
                    <input
                      type="checkbox"
                      name="locale-ids"
                      value={id}
                      bind:group={$state.projectSettings[projectID]
                        .localeIds} />

                    {$db.locale[id]?.title}
                  </label>
                </div>
              {/each}
              <!-- svelte-ignore a11y-missing-content -->
              <a href={`/api/export/p=${project.short_name || project.id}`}>
                Exported raw
              </a>
            {/if}
          </Collapse>
        </paper>
      {/if}
      <GlobalSearch {project} />
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
            <CategoryForm
              on:complete={(c) => {
                visibleForm = null
                if (c.detail.key === undefined) {
                  return
                }
                scrollToCategoryByKey(c.detail.key)
              }}
              {projectID}>
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
      <CategoryList
        {locales}
        {selectedCategory}
        {selectedTranslation}
        {selectedLocale}
        {visibleForm}
        projectID={project.id} />
    {/if}
  </div>
</div>
{#if project}
  <div
    class="backdrop"
    class:sidebarVisible={$state.sidebarVisible}
    on:click={() => ($state.sidebarVisible = !$state.sidebarVisible)} />
  <ProjectOverview {project} />
{/if}

<style>
  h2 {
    display: flex;
    justify-content: space-between;
  }
  .backdrop {
    position: fixed;
    top: 0;
    bottom: 0;
    left: 0;
    right: 0;
    z-index: 1;
    opacity: 0;
    background: #000000aa;
    display: none;
    transition: opacity 1800ms var(--easing-standard);
  }
  .backdrop.sidebarVisible {
    display: block;
    opacity: 1;
  }
</style>
