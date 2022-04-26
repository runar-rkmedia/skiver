<script lang="ts">
  import { db } from 'api'
  import CategoryItem from './CategoryItem.svelte'
  import sortOn from 'sort-on'
  import { closeDialog as _closeDialog, state } from 'state'
  import Dialogs from './Dialogs.svelte'

  export let locales: ApiDef.Locale[]
  export let projectID: string
  export let selectedTranslation: string
  let expandedCategory = ''
  $: categories =
    !!projectID &&
    sortOn(
      Object.values($db.category).filter((c) => c.project_id === projectID),
      ($state.categorySortAsc ? '' : '-') + $state.categorySortOn
    )
</script>

<Dialogs {projectID} />
{#if categories && categories?.length}
  {#each categories as category (category.id)}
    <CategoryItem
      bind:category
      bind:locales
      bind:projectKey={projectID}
      bind:selectedTranslation
      bind:expandedCategory
      forceExpand={false} />
  {/each}
{/if}
