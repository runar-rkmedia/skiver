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
  let sortCategoryOn: keyof ApiDef.Category = 'key'
  let selectedCategory = ''
  let selectedTranslation = ''
  let selectedLocale = ''
  async function onCreateTranslation() {
    $state.createTranslation.category_id = selectedCategory
    if (!$state.createTranslation.category_id) {
      toast({
        title: 'missing argument',
        message: 'category was not set',
        kind: 'warning',
      })
      return
    }
    const s = await api.translation.create($state.createTranslation)
    if (!s[1]) {
      visibleForm = null
      $state.createTranslation = { category_id: '', key: '' }
    }
  }
  async function onCreateTranslationValue() {
    $state.createTranslationValue.locale_id = selectedLocale
    $state.createTranslationValue.translation_id = selectedTranslation
    if (!$state.createTranslationValue.locale_id) {
      toast({
        title: 'missing argument',
        message: 'locale was not set',
        kind: 'warning',
      })
      return
    }
    if (!$state.createTranslationValue.translation_id) {
      toast({
        title: 'missing argument',
        message: 'translation was not set',
        kind: 'warning',
      })
      return
    }
    const s = await api.translationValue.create($state.createTranslationValue)
    if (!s[1]) {
      visibleForm = null
      $state.createTranslationValue = {
        locale_id: '',
        translation_id: '',
        value: '',
      }
    }
  }
  async function onCreateCategory() {
    $state.createCategory.project_id = projectID
    if (!$state.createCategory.project_id) {
      toast({
        title: 'missing argument',
        message: 'project was not set',
        kind: 'warning',
      })
      return
    }
    const s = await api.category.create($state.createCategory)
    if (!s[1]) {
      visibleForm = null
      $state.createCategory = { key: '', project_id: '', title: '' }
    }
  }
  function retry(f: AnyFunc) {
    const r = f()
    if (!r) {
      setTimeout(f, 10)
    }
  }
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
      </Collapse>
    </paper>
  {/if}
  <div>
    <h2>Categories</h2>
    <label>
      Sort by:
      <select bind:value={sortCategoryOn}>
        {#each categorySortOptions as option}
          <option value={option}>{option}</option>
        {/each}
      </select>
    </label>
    <Button
      color="secondary"
      icon={'create'}
      on:click={() => {
        visibleForm = 'category'
      }}>Create category</Button>
    {#if visibleForm === 'category'}
      <paper>
        <form id="category-form">
          <label for="category-id">Key</label>
          <input id="category-key" bind:value={$state.createCategory.key} />
          <label for="category-title">Title</label>
          <input id="category-title" bind:value={$state.createCategory.title} />
          <small>
            <label for="category-description">Description</label>
            <textarea
              id="category-description"
              bind:value={$state.createCategory.description} />
          </small>
          <Button
            color="primary"
            type="submit"
            icon={'create'}
            on:click={onCreateCategory}>
            Create
          </Button>
          <Button
            color="secondary"
            icon={'toggleOff'}
            on:click={() => (visibleForm = null)}>
            Cancel
          </Button>
        </form></paper>
    {/if}
    {#each sortOn(Object.values(project.categories), sortCategoryOn) as category}
      <paper class="category-item" transition fly local>
        <div class="category-item-header" transition:fly|local>
          <h3>
            <code>
              {category.key}
            </code>
            {category.title}
          </h3>
          <div class="description">
            {category.description || ''}
          </div>
          <div class="actions">
            <Button
              color="secondary"
              icon={'create'}
              disabled={selectedCategory === category.id &&
                visibleForm === 'translation'}
              on:click={() => {
                selectedCategory = category.id
                visibleForm = 'translation'
                retry(() => {
                  const el = document.querySelector(
                    '#form-' + category.id + ' input'
                  )
                  if (!el) {
                    return false
                  }
                  el.focus()
                  return true
                })
              }}>Create translation</Button>
          </div>
        </div>
        {#if visibleForm === 'translation' && selectedCategory === category.id}
          <paper>
            <form id={'form-' + selectedCategory}>
              <label>
                Key
                <input name="key" bind:value={$state.createTranslation.key} />
              </label>
              <label>
                Title
                <input
                  name="title"
                  bind:value={$state.createTranslation.title} />
              </label>
              <label>
                Description (Optional, but recommended)
                <textarea
                  name="description"
                  rows="5"
                  bind:value={$state.createTranslation.description} />
              </label>
              <label>
                Context (Optional)
                <input
                  name="prefix"
                  bind:value={$state.createTranslation.context} />
              </label>
              <Button
                color="primary"
                type="submit"
                icon={'create'}
                on:click={onCreateTranslation}>
                Create
              </Button>
              <Button
                color="secondary"
                icon={'cancel'}
                on:click={() => {
                  selectedCategory = ''
                  visibleForm = null
                }}>
                Cancel
              </Button>
            </form></paper>
        {/if}
        <div class="translations" key="={category.id}">
          {#each Object.values(category.translations) as translation}
            <paper class="translation-item">
              <div>
                <h4>
                  <code>
                    {category.key}.{translation.key}
                  </code>
                  {translation.title}
                </h4>
                <div>
                  <small>
                    {translation.description || ''}
                  </small>
                </div>
              </div>
              <table>
                <thead>
                  <th>Language</th>
                  <th>Value</th>
                </thead>
                <tbody>
                  {#each locales as locale}
                    <tr
                      class="locale-item"
                      class:missing={!translation.values?.[locale.id]?.value}
                      class:selected={visibleForm === 'translationValue' &&
                        selectedTranslation === translation.id &&
                        selectedLocale === locale.id}>
                      <td>{locale.title}</td>
                      <td>
                        {#if visibleForm === 'translationValue' && selectedTranslation === translation.id && selectedLocale === locale.id}
                          <form>
                            <!-- svelte-ignore a11y-autofocus -->
                            <textarea
                              autofocus
                              rows={5}
                              bind:value={$state.createTranslationValue.value}
                              type="text"
                              name="value" />
                            <Button
                              color="primary"
                              type="submit"
                              on:click={onCreateTranslationValue}
                              icon={'submit'}>Submit</Button>
                            <Button
                              color="secondary"
                              on:click={() => {
                                visibleForm = null
                              }}
                              icon={'cancel'}>Cancel</Button>
                          </form>
                        {:else}
                          <div
                            class="click-to-edit"
                            on:click={() => {
                              visibleForm = 'translationValue'
                              selectedLocale = locale.id
                              selectedTranslation = translation.id
                              const value =
                                translation.values?.[locale.id]?.value
                              if (value) {
                                $state.createTranslationValue.value = value
                              }
                            }}>
                            <Icon icon="edit" color="primary" />
                            {translation.values?.[locale.id]?.value ||
                              '<no value>'}
                          </div>
                        {/if}
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </paper>
          {/each}
        </div>
      </paper>
    {/each}
  </div>
{/if}

<style>
  .selected {
    outline: 1px dashed hotpink;
  }
  .translation-item {
    padding-block: var(--size-4);
    padding-inline: var(--size-2);
  }
  .translation-item:not(:last-of-type) {
    margin-block-end: var(--size-6);
  }
  .locale-item td:first-of-type {
    white-space: nowrap;
    width: min-content;
  }
  td,
  th {
    padding-inline: var(--size-4);
    padding-block: var(--size-2);
  }
  .locale-item td:not(:first-of-type) {
    width: 100%;
  }
  .missing {
    background-color: var(--color-danger-300);
  }
  .missing:nth-child(even) {
    background-color: var(--color-danger);
  }
  form {
    font-size: 1rem;
  }
  textarea {
    width: 100%;
    max-width: 100%;
    font-size: inherit;
    font-family: inherit;
  }
  #category-form {
    background-color: var(--color-green-300);
  }
  #category-form td {
    padding-block: var(--size-4);
  }
  .category-item-header {
    display: grid;
    grid-template-columns: 1fr 2fr 1fr;
    align-items: center;
  }
  .category-item form {
    display: block;
  }
  .click-to-edit {
    cursor: pointer;
  }
</style>
