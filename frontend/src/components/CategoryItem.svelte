<script lang="ts">
  import { db, api } from 'api'
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
  import { inview } from 'svelte-inview'

  let isInView = false
  const options = {
    rootMargin: '50px',
    unobserveOnEnter: true,
  }
  const handleViewChange = ({ detail }) => {
    console.log('viewchange', detail)
    isInView = detail.inView
  }
  export let category: ApiDef.Category
  export let locales: ApiDef.Locale[]
  export let projectKey: ApiDef.Locale[]
  export let selectedLocale: string
  export let selectedTranslation: string
  export let selectedCategory: string
  export let visibleForm: string
  // export let forceExpand = false
  $: translations = (category.translation_ids || [])
    .map((tid) => $db.translation[tid])
    .filter(Boolean)

  // export let expandedCategory: string | boolean
  $: categoryPath = (category.key || '').split('.')
  function handleCatClick(e) {
    e.stopPropocation()
    e.preventDefault()
  }
</script>

<div
  use:inview={options}
  on:change={handleViewChange}
  on:init={(e) => console.log('init???', e)}
  class="item"
  id={'cat-' + category.key}>
  <div class="desc category-item-header">
    <div>
      <h3>
        {#each categoryPath as subPath, i}
          {#if i !== categoryPath.length - 1}
            <a
              href={'#cat-' + category.key}
              on:click|preventDefault={(e) => {
                const el = document.getElementById(
                  'cat-' + categoryPath.slice(0, i + 1).join('.')
                )
                console.log('el', el)
                if (!el) {
                  return
                }
                el.scrollIntoView({ behavior: 'smooth' })
              }}>{subPath}</a>
            <span class="sep">/</span>
          {/if}
        {/each}
        <span title={category.key}>
          {category.title}
        </span>
      </h3>
      <div class="description">
        {category.description || ''}
      </div>
    </div>
    <div class="right">
      <EntityDetails entity={category} />
    </div>
  </div>
  {#if isInView}
    <div class="box">
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
        {#each sortOn(translations, ($state.categorySortAsc ? '' : '-') + $state.sortCategoryOn) as translation}
          <paper class="translation-item">
            <TranslationItem
              {translation}
              {projectKey}
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
    </div>
  {:else}
    <div class="placeholder tbox">Placeholder</div>
  {/if}
</div>

<style>
  .item {
    display: flex;
    flex-direction: column;
  }
  .desc {
    position: sticky;
    position: -webkit-sticky;
    top: 0px;
    max-height: 200px;
    z-index: 10000;
    color: white;
  }
  .description {
    transition: max-height 1500ms ease-in-out;
    padding-bottom: var(--size-2);
  }

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
    margin-inline: calc(var(--size-4) * -1);
    backdrop-filter: brightness(20%) saturate(180%) blur(2px);
    padding-inline: var(--size-4);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  a {
    font-size: small;
    color: inherit;
  }
  .sep {
    opacity: 0.5;
    padding-inline: var(--size-2);
  }
</style>
