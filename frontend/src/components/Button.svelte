<script type="ts">
  import type { Icon as IconType } from './icons'
  import Icon from './Icon.svelte'
  import type { Colors } from '../appTypes'
  import { createEventDispatcher } from 'svelte'
  export let preventDefault = true

  export let active = false
  export let toggle: boolean | null = null
  export let icon: IconType | undefined = undefined
  // TODO: support all colors
  export let color: Colors | 'danger' = 'no-color'
  export let disabled: boolean = false
  export let type: string = ''
  const dispatch = createEventDispatcher()
  $: iconToUse =
    icon || (toggle === true && 'toggleOn') || (toggle === false && 'toggleOff')
</script>

<button
  class:btn-reset={true}
  class={color}
  class:active
  class:toggle
  {type}
  class:icon-button={!!icon || toggle}
  {disabled}
  on:click={(e) => {
    if (preventDefault) {
      e.preventDefault()
    }
    dispatch('click', e)
  }}>
  {#if iconToUse}
    <Icon icon={iconToUse} />
  {/if}
  <slot />
</button>

<style>
  button:focus:not(:disabled):not(.no-color),
  button:hover:not(:disabled):not(.no-color) {
    /* TODO: move to psudo-element and transition opacity for perf. */
    box-shadow: 0 8px 16px -2px rgba(0, 32, 128, 0.25);
  }
  button.no-color:focus:not(:disabled),
  button.no-color:hover:not(:disabled) {
    /* TODO: move to psudo-element and transition opacity for perf. */
    text-decoration: underline;
  }

  .active {
    filter: brightness(0.8);
    text-decoration: underline;
    text-decoration-color: var(--color-red);
    text-decoration-thickness: 4px;
  }
</style>
