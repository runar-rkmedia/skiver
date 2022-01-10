<script lang="ts">
  import CategoryItem from './CategoryItem.svelte'
  import Pagination from './Pagination.svelte'
  export let categories: ApiDef.Category[]
  export let locales: ApiDef.Locale[]
  export let selectedLocale: string
  export let selectedTranslation: string
  export let selectedCategory: string
  export let visibleForm: string
  let expandedCategory = ''
</script>

{#if categories.length}
  <Pagination count={categories.length} let:page let:pageSize>
    {#each categories.slice(pageSize * page, pageSize * page + pageSize) as category (category.id)}
      <CategoryItem
        bind:category
        bind:locales
        bind:selectedLocale
        bind:selectedTranslation
        bind:selectedCategory
        bind:expandedCategory
        forceExpand={false}
        bind:visibleForm />
    {/each}
  </Pagination>
{/if}
