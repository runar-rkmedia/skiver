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

  import TranslationValueForm from 'forms/TranslationValueForm.svelte'
  import TranslationForm from 'forms/TranslationForm.svelte'
  import TranslationItem from 'components/TranslationItem.svelte'
  import EntityDetails from 'components/EntityDetails.svelte'
  import {
    createCategoryAnchorProps,
    scrollToCategory,
  } from 'util/scrollToCategory'
  import { inview } from 'svelte-inview'
  import ScrollAnchor from './ScrollAnchor.svelte'

  let isInView = false
  const options = {
    rootMargin: '50px',
    unobserveOnEnter: true,
  }

  const handleViewChange = (e) => {
    if (!e.detail) {
      return
    }
    isInView = e.detail.inView
  }
  export let category: ApiDef.Category
  export let locales: ApiDef.Locale[]
  export let projectKey: ApiDef.Locale[]
  export let selectedLocale: string
  export let selectedTranslation: string
  export let selectedCategory: string
  export let visibleForm: string
  let showDeleted = true
  // export let forceExpand = false
  $: translations = (category.translation_ids || [])
    .map((tid) => $db.translation[tid])
    // .filter(Boolean)
    .filter((t) => !!t && (!t.deleted || showDeleted))

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
  class="item"
  id={'cat-' + (category.key || '_root_')}>
  <div class="desc category-item-header">
    <div>
      <h3>
        <button
          title="Overview-menu"
          class="btn-reset menu"
          on:click={() => ($state.sidebarVisible = !$state.sidebarVisible)}>
          <Icon icon="menu" />
        </button>
        {#each categoryPath as subPath, i}
          {#if i !== categoryPath.length - 1}
            <a
              href={createCategoryAnchorProps({
                key: categoryPath.slice(0, i + 1).join('.'),
              }).href}
              on:click|preventDefault={scrollToCategory}>{subPath}</a>
            <span class="sep">/</span>
          {/if}
        {/each}
        <span title={category.key}>
          {category.title || '(Root)'}
          {#if category.translation_ids?.length}
            <small>
              ({category.translation_ids.length})
            </small>
          {/if}
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
            on:complete={(c) => {
              visibleForm = null
            }}>
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
          <paper class="translation-item" class:deleted={!!translation.deleted}>
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
    <paper class="placeholder tbox">
      <div class="placeholder-entity" />
      <div class="placeholder-title" />
      <div class="placeholder-table" />
    </paper>
  {/if}
</div>

<style>
  .item {
    display: flex;
    flex-direction: column;
  }
  .box {
    overflow-x: hidden;
  }
  .deleted {
    background-image: repeating-linear-gradient(
      45deg,
      var(--color-grey-100),
      var(--color-grey-100) 30px,
      var(--color-grey-300) 30px,
      var(--color-grey-300) 60px
    );
  }
  .deleted > :global(:not(form)) {
    opacity: 0.7;
  }
  .desc {
    position: sticky;
    position: -webkit-sticky;
    top: 0px;
    max-height: 200px;
    z-index: 1;
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
    background-color: #092335f0;
    padding-inline: var(--size-4);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  @supports ((-webkit-backdrop-filter: none) or (backdrop-filter: none)) {
    .category-item-header {
      /* background-color: unset; */
      backdrop-filter: brightness(20%) saturate(180%) blur(2px);
    }
  }

  .category-item-header small {
    opacity: 0.7;
  }
  a {
    font-size: small;
    color: inherit;
  }
  .sep {
    opacity: 0.5;
    padding-inline: var(--size-2);
  }
  .menu {
    margin-inline-start: calc(var(--size-4) * -1);
    font-size: 110%;
    transition: transform, color 120ms var(--easing-standard);
    color: var(--color-primary);
  }
  .menu:hover {
    transform: scale(1.16);
  }
  .placeholder-entity {
    width: 230px;
    height: 37px;
    position: absolute;
    right: var(--size-2);
    top: var(--size-4);
  }
  .placeholder-title {
    width: 250px;
    height: 37px;
    position: absolute;
    left: var(--size-2);
    top: var(--size-4);
  }
  .placeholder-table {
    height: 100px;
    position: absolute;
    left: var(--size-2);
    right: var(--size-2);
    bottom: var(--size-4);
  }
</style>
