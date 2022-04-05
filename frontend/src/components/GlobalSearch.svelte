<script lang="ts">
  import { db } from 'api'
  import { onMount } from 'svelte'

  import GlobalSearchInner from './GlobalSearchInner.svelte'
  import Icon from './Icon.svelte'
  let query = ''
  export let project: ApiDef.Project
  let visible = false
  let input: HTMLInputElement
  let stickyTop = false
  const observer = new IntersectionObserver(
    ([e]) => {
      const p = e.target as HTMLDivElement
      console.log('observer trigger', p)
      if (!p) {
        return
      }
      // p.classList.toggle('bg-dark', e.intersectionRatio < 1)
      stickyTop = e.intersectionRatio < 1
    },
    { threshold: [1], rootMargin: '-1px 0px 0px 0px' }
  )
  onMount(() => {
    observer.observe(input)
    console.log('observing', input, observer)
  })
  function scrollTop() {
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }
</script>

<div
  class="wrapper"
  bind:this={input}
  class:sticky={stickyTop}
  class:bg-dark={stickyTop}>
  <h2>
    Search
    <input
      on:focus={scrollTop}
      placeholder="John Doe"
      type="search"
      bind:value={query}
      on:focus={() => (visible = true)} />
  </h2>

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
    margin-inline-start: calc(var(--size-4) * -1);
    padding-inline-start: var(--size-4);
    padding-inline-end: var(--size-7);
    height: var(--size-12);
    transition: all 250ms ease-in-out;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  .wrapper.sticky h2,
  .wrapper.sticky .right a {
    color: var(--color-grey-100);
  }
</style>
