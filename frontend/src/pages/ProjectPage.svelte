<script lang="ts">
  import { db, api, projects } from 'api'
  import { state, toast } from 'state'
  import { fly, slide, draw, fade, blur, crossfade } from 'svelte/transition'
  import Button from 'components/Button.svelte'
  import type { AnyFunc } from 'simplytyped'
  import Collapse from 'components/Collapse.svelte'
  export let projectID: string
  import sortOn from 'sort-on'
  import Icon from 'components/Icon.svelte'
  import id from 'date-fns/locale/id'
  import { t } from '../util/i18next'
  import { formatDate } from '../dates'

  import CategoryForm from 'forms/CategoryForm.svelte'
  import TranslationValueForm from 'forms/TranslationValueForm.svelte'
  import TranslationForm from 'forms/TranslationForm.svelte'
  import TranslationItem from 'components/TranslationItem.svelte'

  $: project = $projects[projectID]
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

{$t('a.hello')}
{$t('actions.create')}
{$t('actions.update')}
{$t('actions.delete')}
{$t('forms.submit', { value: 4 })}

<button on:click={() => t.changeLanguage('nb')}>toggle lang</button>
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
  <small
    >Created: {formatDate(project.createdAt)} | Updated: {formatDate(
      project.updatedAt
    )}</small>
  {#if $state.projectSettings[projectID]?.localeIds}
    <paper>
      <Collapse>
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
        <a href={`/api/export/p=${project.shortName || project.id}`}>
          Exported raw
        </a>
      </Collapse>
    </paper>
  {/if}
  <div>
    <h2>Categories</h2>
    <label>
      Sort by: {$state.sortCategoryOn}
      <select bind:value={$state.sortCategoryOn}>
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
    {#each sortOn(Object.values(project.categories), ($state.categorySortAsc ? '' : '-') + $state.sortCategoryOn) as category}
      <paper class="category-item" transition:fly|local>
        <div class="category-item-header">
          <h3>
            <code>
              {category.key}
            </code>
            {category.title}
          </h3>
          <div class="description">
            {category.description || ''}
          </div>
          <small
            >Created: {formatDate(category.createdAt)} | Updated: {formatDate(
              category.updatedAt
            )}</small>
          <div class="actions">
            <Button
              color="secondary"
              icon={'create'}
              disabled={selectedCategory === category.id &&
                visibleForm === 'translation'}
              on:click={() => {
                selectedCategory = category.id
                visibleForm = 'translation'
              }}>Create translation</Button>
          </div>
        </div>
        {#if visibleForm === 'translation' && selectedCategory === category.id}
          <paper>
            <TranslationForm
              categoryID={selectedCategory}
              on:complete={() => (visibleForm = null)}>
              <Button
                slot="actions"
                color="secondary"
                icon={'cancel'}
                on:click={() => {
                  selectedCategory = ''
                  visibleForm = null
                }}>
                Cancel
              </Button>
            </TranslationForm>
          </paper>
        {/if}
        <div class="translations" key="={category.id}">
          {#each sortOn(Object.values(category.translations), ($state.categorySortAsc ? '' : '-') + $state.sortCategoryOn) as translation}
            <paper class="translation-item">
              <TranslationItem
                {translation}
                translationValues={translation.values}
                categoryKey={category.key}
                {locales}
                bind:selectedLocale
                on:complete={() => {
                  visibleForm = null
                }}
                on:showForm={({ detail: { show } }) => {
                  if (show) {
                    visibleForm = 'translationValue'
                    selectedTranslation = translation.id
                    return
                  }
                  visibleForm = null
                }}
                showForm={visibleForm === 'translationValue' &&
                  selectedTranslation === translation.id} />
            </paper>
          {/each}
        </div>
      </paper>
    {/each}
  </div>
{/if}

<style>
  .translation-item {
    padding-block: var(--size-4);
    padding-inline: var(--size-2);
  }
  .translation-item:not(:last-of-type) {
    margin-block-end: var(--size-6);
  }
  .category-item-header {
    display: grid;
    grid-template-columns: 1fr 2fr 1fr;
    align-items: center;
  }
</style>
