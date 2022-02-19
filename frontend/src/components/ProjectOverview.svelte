<script lang="ts">
  import { db } from 'api'
  import Icon from 'components/Icon.svelte'
  import sortOn from 'sort-on'
  import { state } from 'state'
  import {
    createCategoryAnchorProps,
    scrollToCategory,
  } from 'util/scrollToCategory'
  import ScrollAnchor from './ScrollAnchor.svelte'

  export let project: ApiDef.Project
  $: categories = sortOn(
    (project.category_ids || []).map((cid) => $db.category[cid]),
    'key'
  )

  const toggleVisibility = () =>
    ($state.sidebarVisible = !$state.sidebarVisible)
  function scrollToC(e: any) {
    $state.sidebarVisible = false
  }
</script>

<div class="wrapper" class:visible={$state.sidebarVisible}>
  <div class="button">
    <button class="btn-reset" on:click={toggleVisibility}>
      <Icon icon="menu" />
    </button>
  </div>

  {#each categories as c}
    {#if c}
      <div class="key cat" data-depth={(c.key || '').split('.').length - 1}>
        <ScrollAnchor category={c} on:scrollTo={scrollToC} />
      </div>
    {/if}
  {/each}
</div>

<style>
  :root {
    --width: 280px;
    --bg: #000000aa;
    --bgb: hsl(240, 70%, 20%);
    --indent: 12px;
  }
  .wrapper {
    z-index: 1;
    overflow-y: auto;
    overflow-x: hidden;
    height: 100vh;
    scrollbar-width: none;
    -ms-overflow-style: none;

    font-size: 1rem;
    position: fixed;
    left: 0;
    width: var(--width);
    top: 0;
    background: var(--bg);
    color: var(--color-grey-100);
    transform: translateX(calc(var(--width) * -1));

    transition: transform 200ms var(--easing-standard);
    padding-inline: var(--size-2);
    padding-block: var(--size-4);
  }
  .wrapper::-webkit-scrollbar {
    width: 0 !important;
    background: transparent; /* make scrollbar transparent */
  }
  .visible .button {
    opacity: 0;
  }
  .cat :global(a) {
    color: white;
    display: flex;
    margin-block: var(--size-2);
    justify-content: space-between;
  }
  .cat :global(a small) {
    opacity: 0.7;
  }
  .button {
    position: absolute;
    top: 100px;
    font-size: 1.5rem;
    /* opacity: calc(177 / 255); */
    opacity: 0.4;
    transform: translateX(calc(var(--width) - 10px));
    transition: transform, opacity 100ms var(--easing-standard);
  }
  .button:hover {
    transform: translateX(var(--width));
    opacity: 1;
  }
  .key[data-depth='1'] {
    padding-inline-start: var(--indent);
  }
  .key[data-depth='2'] {
    padding-inline-start: calc(var(--indent) * 2);
    font-size: 0.9em;
  }
  .key[data-depth='3'] {
    padding-inline-start: calc(var(--indent) * 3);
    font-size: 0.8em;
  }

  button {
    background: var(--bgb);
    width: 48px;
    height: 48px;
    position: relative;
    border-top-right-radius: var(--radius-md);
    border-bottom-right-radius: var(--radius-lg);
  }
  .visible {
    transform: translateX(0);
  }
</style>
