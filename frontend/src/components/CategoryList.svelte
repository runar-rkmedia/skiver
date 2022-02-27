<script lang="ts">
  import { db } from 'api'
  import sortOn from 'sort-on'
  import { state } from 'state'

  import CategoryItem from './CategoryItem.svelte'
  export let locales: ApiDef.Locale[]
  export let selectedLocale: string
  export let projectID: string
  export let selectedTranslation: string
  export let selectedCategory: string
  export let visibleForm: string | null
  let expandedCategory = ''
  $: categories =
    !!projectID &&
    sortOn(
      Object.values($db.category).filter((c) => c.project_id === projectID),
      ($state.categorySortAsc ? '' : '-') + $state.categorySortOn
    )
</script>

{#if categories && categories?.length}
  {#each categories as category (category.id)}
    <CategoryItem
      bind:category
      bind:locales
      bind:projectKey={projectID}
      bind:selectedLocale
      bind:selectedTranslation
      bind:selectedCategory
      bind:expandedCategory
      forceExpand={false}
      bind:visibleForm />
  {/each}
{/if}
