<script lang="ts">
  export let latest: ApiDef.ReleaseInfo | undefined
  export let currentVersion: string
  $: isOldVersion =
    latest?.tag_name &&
    currentVersion &&
    latest.tag_name.replace(/^v/, '') !== currentVersion.replace(/^v/, '')
  $: v = currentVersion ? currentVersion.replace(/^v?/, 'v') : ''
  $: latestText =
    latest &&
    `There is a newer version available:

${latest.tag_name}

${latest.body}
`
  console.log('hello', latest)
</script>

{@debug latest}

{#if isOldVersion && latest}
  <div class="old" title={latestText}>
    {v} <span class="new">({latest.tag_name} is available!)</span>
  </div>
{:else if latest}
  <div class:latest={!!latest} title={latestText}>
    {v}
  </div>
{:else}
  <div>
    {v}
  </div>
{/if}

<style>
  .latest,
  .new {
    color: var(--color-success-icon);
  }
  .old {
    color: var(--color-warning-icon);
  }
</style>
