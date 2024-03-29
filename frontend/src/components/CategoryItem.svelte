<script lang="ts">
  import { db, api } from 'api'
  import { state, toast, showDialog } from 'state'
  import Button from 'components/Button.svelte'
  import LocaleFlag from 'components/LocaleFlag.svelte'
  import Dialog from 'components/Dialog.svelte'
  import type { AnyFunc } from 'simplytyped'
  import Collapse from 'components/Collapse.svelte'
  import sortOn from 'sort-on'
  import Icon from 'components/Icon.svelte'
  import id from 'date-fns/locale/id'
  import { t } from '../util/i18next'
  import { formatDate } from '../dates'

  import TranslationValueForm from 'forms/TranslationValueForm.svelte'
  import TranslationForm from 'forms/TranslationForm.svelte'
  import EntityDetails from 'components/EntityDetails.svelte'
  import {
    createCategoryAnchorProps,
    scrollToCategory,
  } from 'util/scrollToCategory'
  import { inview } from 'svelte-inview'
  import ScrollAnchor from './ScrollAnchor.svelte'
  import TranslationItemCompactRow from './TranslationItemCompactRow.svelte'
  import TranslationItemLegacy from './TranslationItemLegacy.svelte'

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
  export let selectedTranslation: string
  let visibleForm: string
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
  $: selectedLocale = $state?.columns?.valueForLocale?.[0] || locales[0]
  $: sortedTranslations = sortOn(
    translations,
    ($state.categorySortAsc ? '' : '-') + $state.sortCategoryOn
  )
</script>

<div
  use:inview={options}
  on:change={handleViewChange}
  class="item"
  id={'cat-' + (category.key || '_root_')}>
  <div class="desc category-item-header bg-dark">
    <div>
      <h3>
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
          {category.title || (!categoryPath.length ? '(Root)' : '(No name)')}
          {#if category.translation_ids?.length}
            <small>
              ({category.translation_ids.length})
            </small>
          {/if}
        </span>
        <Button
          icon="edit"
          on:click={() => {
            showDialog({
              kind: 'editCategory',
              id: category.id,
              title: `Edit ${category.title}`,
            })
          }}
          disabled={!!$state.dialog}>
          Edit
        </Button>
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
          disabled={!!$state.dialog}
          on:click={() => {
            showDialog({ kind: 'createTranslation', parent: category.id })
          }}>Create translation</Button>
      </div>
      <div class="translations" key="={category.id}">
        {#if $state.translationType === 'legacy'}
          {#each sortedTranslations as translation}
            <paper
              class="translation-item"
              class:deleted={!!translation.deleted}>
              <TranslationItemLegacy
                {translation}
                {projectKey}
                categoryKey={category.key}
                {locales}
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
        {:else}
          <paper>
            <table class="compactHeader">
              <thead>
                <th
                  title="Toggle key/title"
                  on:click={() => {
                    state.update((s) => ({
                      ...s,
                      columns: {
                        ...s.columns,
                        ...(!!s.columns.title
                          ? { title: false, key: true }
                          : { title: true, key: false }),
                      },
                    }))
                  }}>
                  {#if $state.columns.title}
                    Title
                  {:else}
                    Key
                  {/if}
                </th>
                {#each $state.columns.valueForLocale as l, i}
                  <th
                    title="Toggle locale"
                    on:click={() => {
                      state.update((s) => {
                        let loc = l
                        const index = locales.findIndex((f) => f.id === loc.id)
                        if (index < 0) {
                          loc = locales[0]
                        }
                        loc = locales[(index + 1) % locales.length]
                        const v = s.columns.valueForLocale
                        v[i] = loc
                        return {
                          ...s,
                          columns: {
                            ...s.columns,
                            valueForLocale: v,
                          },
                        }
                      })
                    }}
                    >Value
                    <span class="flag">
                      <LocaleFlag locale={l} />
                    </span>
                    {l.title}</th>
                {/each}
              </thead>
              <tbody>
                {#each sortedTranslations as translation}
                  <TranslationItemCompactRow
                    {translation}
                    columns={$state.columns}
                    {projectKey}
                    categoryKey={category.key}
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
                {/each}
              </tbody>
            </table>
          </paper>
        {/if}
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
  .compactHeader th:hover {
    cursor: pointer;
    text-decoration: underline;
  }
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
  .desc {
    position: sticky;
    position: -webkit-sticky;
    top: calc(var(--size-12) - 1px);
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
    margin-inline-start: calc(var(--gutter-start) * -1);
    margin-inline-end: calc(var(--gutter-end) * -1);
    padding-inline: var(--size-4);
    padding-inline-end: var(--size-6);
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .category-item-header :global(button),
  .category-item-header small {
    opacity: 0.7;
    font-size: small;
  }
  a {
    font-size: small;
    color: inherit;
  }
  .sep {
    opacity: 0.5;
    padding-inline: var(--size-2);
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
