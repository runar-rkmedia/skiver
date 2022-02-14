<script lang="ts">
  import { state } from '../state'

  import { onMount } from 'svelte'
  import Icon from './Icon.svelte'

  export let show = false
  export let key: string
  export let forceShow = false
  onMount(() => {
    if (forceShow) {
      return
    }
    if (!key) {
      return
    }
    show = $state.collapse[key]
  })
</script>

<div class="collapse">
  <button
    class="btn-reset toggle"
    aria-label="Collapse"
    on:click|preventDefault={() => {
      if (forceShow) {
        return
      }
      show = !show
      if (!key) {
        return
      }
      $state.collapse[key] = show
    }}>
    <slot name="title" class="title" />
    {#if !forceShow}
      <div class="icon">
        {#if show}
          <Icon icon={'collapseUp'} class="toggle-icon" />
        {:else}
          <Icon icon={'collapseDown'} class="toggle-icon" />
        {/if}
      </div>
    {/if}
  </button>
  {#if show || forceShow}
    <slot show={$state.collapse[key]} />
  {/if}
</div>

<style>
  button.toggle {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
  }

  .icon {
    font-size: 1.4rem;
  }
  :global(.collapse > fieldset) {
    border: unset;
  }
</style>
