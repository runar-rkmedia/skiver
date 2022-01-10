<script lang="ts">
  import Button from './Button.svelte'
  export let count: number
  export let page = 0
  export let pageSize = 50
  export let position: 'top' | 'bottom' | 'both' = 'both'
</script>

{#if position !== 'bottom'}
  <pagination>
    <Button color={'secondary'} disabled={page <= 0} on:click={() => page--}
      >Prev</Button>
    <page-count>
      item {page * pageSize + 1}-{Math.min(
        Math.ceil((page + 1) * pageSize),
        count
      )}{#if count > pageSize}{' '}total: {count}{/if}
    </page-count>
    <select bind:value={pageSize}>
      <option value={10}>10 per page</option>
      <option value={50}>50 per page</option>
      <option value={100}>100 per page</option>
    </select>
    <Button
      color={'secondary'}
      disabled={page + 1 > count / pageSize}
      on:click={() => page++}>Next</Button>
  </pagination>
{/if}
<slot {page} {pageSize} {count} />
{#if position !== 'top'}
  <pagination>
    <Button color={'secondary'} disabled={page <= 0} on:click={() => page--}
      >Prev</Button>
    <page-count>
      item {page * pageSize + 1}-{Math.min(
        Math.ceil((page + 1) * pageSize),
        count
      )}{#if count > pageSize}{' '}total: {count}{/if}
    </page-count>
    <select bind:value={pageSize}>
      <option value={10}>10 per page</option>
      <option value={50}>50 per page</option>
      <option value={100}>100 per page</option>
    </select>
    <Button
      color={'secondary'}
      disabled={page + 1 > count / pageSize}
      on:click={() => page++}>Next</Button>
  </pagination>
{/if}

<style>
  pagination {
    display: flex;
    justify-content: flex-end;
    align-items: center;
  }
  page-count {
    user-select: none;
    padding-inline: var(--size-2);
  }
</style>
