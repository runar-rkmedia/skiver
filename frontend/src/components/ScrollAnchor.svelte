<script lang="ts">
  import { createEventDispatcher } from 'svelte'

  import {
    createCategoryAnchorProps,
    scrollToCategory,
  } from 'util/scrollToCategory'
  export let category: ApiDef.Category
  const dispatch = createEventDispatcher()
  const props = createCategoryAnchorProps(category)
</script>

{#if props}
<a
  href={props.href}
  on:click|preventDefault={(e) => {
    const el = scrollToCategory(e)
    dispatch('click', e)
    if (el) {
      dispatch('scrollTo', { el, e })
    }
  }}>
  <slot name="pre" />
  <slot>
    {category.title || '(Root)'}
    {#if category.translation_ids?.length}
      <small>
        {category.translation_ids.length}
      </small>
    {/if}
  </slot>
</a>
  {:else}
  ???
{/if}
