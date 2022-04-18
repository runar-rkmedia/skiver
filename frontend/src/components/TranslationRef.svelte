<script lang="ts">
  import { findTranslationByKey } from 'api'

  import { showDialog } from 'state'
  import Icon from './Icon.svelte'

  export let ref: string
  $: hits = findTranslationByKey(ref)
</script>

{#if hits}
  <div class:multiple={hits.length > 1}>
    <Icon icon="success" color="success" />
    {ref}
    {#each hits as hit}
      <button
        class="btn-reset"
        on:click={() => {
          if (!hit.translation || !hit.category) {
            return
          }
          showDialog({
            kind: 'translation',
            id: hit.translation.id,
            parent: hit.category.id,
            title: 'Translation:' + hit.translation.title,
          })
        }}>
        <code>
          {hit.translation.key}
        </code>
        {hit.translation.title}
      </button>
    {/each}
  </div>
{:else}
  <div class="not-found">
    <Icon icon="warning" color="warning" />
    {ref}
  </div>
{/if}

<style>
  .not-found {
    background: hotpink;
  }
  .multiple > * {
    display: block;
  }
</style>
