<script type="ts">
  import Button from '../components/Button.svelte'

  import { slide } from 'svelte/transition'
  export let deleted: boolean
</script>

<li transition:slide|local class:deleted>
  <div class="item-content">
    <div class="item-header">
      <slot name="header" />
    </div>
    {#if $$slots.error}
      <div class="error">
        <slot name="error" class="error" />
      </div>
    {/if}

    {#if $$slots.details}
      <div class="sub-item">
        <slot name="details" />
      </div>
    {/if}
    <div class="item-description">
      <slot name="description" />
    </div>
  </div>
  {#if $$slots.actions}
    <div class="item-actions">
      <slot name="actions" />
    </div>
  {/if}
</li>

<style>
  .deleted {
    background-color: hsl(0, 88%, 96.1%);
  }
  .deleted .item-header {
    text-decoration: line-through;
  }
  .error {
    color: var(--color-danger-700);
  }
  li {
    background-color: var(--color-grey-100);
    display: flex;
    justify-content: space-between;
    width: 100%;
  }
  li:nth-child(even) {
    background-color: var(--color-grey-300);
  }
  .item-header {
    font-size: large;
  }
  .item-description {
    font-size: small;
  }
  .item-content {
    padding-inline: var(--size-4);
    margin-block-start: var(--size-3);
    margin-block-end: var(--size-2);
    width: 100%;
  }
  .sub-item {
    margin-inline-start: 16px;
  }
  .item-actions {
    padding: var(--size-2);
    display: flex;
    flex-direction: column;
  }
</style>
