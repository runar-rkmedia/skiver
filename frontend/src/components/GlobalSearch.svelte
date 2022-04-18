<script lang="ts">
  import { db } from 'api'
  import { state } from 'state'
  import GlobalSearchInner from './GlobalSearchInner.svelte'
  import Icon from './Icon.svelte'
  import Button from './Button.svelte'
  let query = ''
  export let project: ApiDef.Project
  let visible = false
  let input: HTMLDivElement
  function scrollTop() {
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }
</script>

<div class="wrapper" bind:this={input}>
  <div class="left">
    <Button on:click={() => ($state.sidebarVisible = !$state.sidebarVisible)}>
      <Icon icon="menu" />
    </Button>
  </div>
  <div class="center">
    <label for="global-search">
      <Icon icon="search" />
      Search</label>
    <input
      id="global-search"
      on:focus={scrollTop}
      placeholder="Search"
      type="search"
      bind:value={query}
      on:focus={() => (visible = true)} />
  </div>

  <div class="right">
    <a href={'#project/' + project.id + '/settings'} on:click={scrollTop}>
      <Icon icon="settings" />
      Settings
    </a>
  </div>
</div>
{#if project && visible && !Object.values($db.responseStates).some((rs) => rs.loading)}
  <GlobalSearchInner {project} {query} />
{/if}

<style>
  .wrapper {
    position: sticky;
    top: 0;
    z-index: 2;
    width: 100vw;
    margin-block-start: calc(var(--size-2) * -1);
    margin-inline-start: calc(var(--size-4) * -1);
    padding-inline-start: var(--size-2);
    padding-inline-end: var(--size-7);
    height: var(--size-12);
    transition-property: color, background-color;
    transition-duration: 150ms;
    transition-timing-function: var(--easing-standard);
    display: flex;
    justify-content: space-between;
    align-items: center;
    background-color: var(--color-primary-700);
    box-shadow: var(--elevation-2);
    color: var(--color-grey-100);
  }
</style>
