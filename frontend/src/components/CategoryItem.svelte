<script lang="ts">
  import { db, api, projects } from 'api'
  import { state, toast } from 'state'
  import { fly } from 'svelte/transition'
  import Button from 'components/Button.svelte'
  import type { AnyFunc } from 'simplytyped'
  import Collapse from 'components/Collapse.svelte'
  import sortOn from 'sort-on'
  import Icon from 'components/Icon.svelte'
  import id from 'date-fns/locale/id'
  import { t } from '../util/i18next'
  import { formatDate } from '../dates'

  import CategoryForm from 'forms/CategoryForm.svelte'
  import TranslationValueForm from 'forms/TranslationValueForm.svelte'
  import TranslationForm from 'forms/TranslationForm.svelte'
  import TranslationItem from 'components/TranslationItem.svelte'
  import EntityDetails from 'components/EntityDetails.svelte'
  export let category: ApiDef.Category
  export let locales: ApiDef.Locale[]
  export let projectKey: ApiDef.Locale[]
  export let selectedLocale: string
  export let selectedTranslation: string
  export let selectedCategory: string
  export let visibleForm: string
  export let forceExpand = false

  // export let expandedCategory: string | boolean
</script>

<paper class="category-item">
  <paper-count>{Object.keys(category.translations).length || 0}</paper-count>
  <Collapse key={category.id} forceShow={forceExpand}>
    <div class="category-item-header" slot="title">
      <h3>
        {#if category.key !== '___root___'}
          <code>
            {category.key}
          </code>
        {/if}
        {category.title}
      </h3>
      <div class="description">
        {category.description || ''}
      </div>
      <div class="right">
        <EntityDetails entity={category} />
      </div>
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
        }}>Create translation</Button>
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
            {projectKey}
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
  </Collapse>
</paper>

<style>
  .right {
    text-align: right;
  }
  .actions {
    padding-block: var(--size-4);
    display: flex;
    justify-content: flex-end;
  }
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
