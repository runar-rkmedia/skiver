<script lang="ts">
  export let json: string
  export let maxLength = 66
</script>

{#if json.length > maxLength}
  <details>
    <summary>{json}</summary>
    <pre>{json}</pre>
  </details>
{:else}
  {json}
{/if}

<style>
  summary {
    max-width: 100%;
    text-overflow: ellipsis;
    overflow-x: hidden;
    white-space: nowrap;
  }
</style>
