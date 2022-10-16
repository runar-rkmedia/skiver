<script lang="ts">
  import { createEventDispatcher, onMount, onDestroy } from 'svelte'
  import { fade, scale } from 'svelte/transition'
  const dispatch = createEventDispatcher()
  const handleKeydown = (e: KeyboardEvent) => {
    if (e.key === 'Escape') {
      dispatch('clickClose')
    }
  }
  onMount(() => {
    document.body.classList.add('stop-scrolling')
  })
  onDestroy(() => {
    document.body.classList.remove('stop-scrolling')
  })
</script>

<svelte:window on:keydown={handleKeydown} />
<embed-wrapper transition:fade={{ duration: 150 }}>
  <div
    class="simple-backdrop"
    on:click={() => {
      dispatch('clickClose')
    }} />
  <div class="content" transition:scale={{ duration: 250 }}>
    {#if $$slots.title}
      <div class="title">
        <paper>
          <h3>
            <slot name="title" />
          </h3>
        </paper>
      </div>
    {/if}
    <slot />
  </div>
</embed-wrapper>

<style>
  .title paper {
    margin: 0;
    border-bottom-right-radius: 0;
    border-bottom-left-radius: 0;
    border-bottom: 1px solid var(--color-grey-500);
  }
  .title :global(+ paper) {
    border-top-right-radius: 0;
    border-top-left-radius: 0;
  }
  .simple-backdrop {
    width: 100vw;
    height: 100vh;
    position: fixed;
    opacity: 0.5;
  }
  embed-wrapper {
    position: fixed;
    top: 0;
    right: 0;
    bottom: 0;
    left: 0;
    display: flex;
    justify-content: center;
    z-index: 2;
    overflow-y: auto;
  }
  embed-wrapper .content {
    overflow-y: auto;
    width: 100%;
    padding-block: var(--size-8);
    padding-inline: var(--size-8);
  }
</style>
