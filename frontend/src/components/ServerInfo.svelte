<script lang="ts">
  import { db } from '../api'

  import formatDate, { isValidDate } from '../dates'
  import LatestRelease from './LatestRelease.svelte'

  $: serverStartedAt = isValidDate($db.serverInfo.server_started_at)
  $: buildDate = isValidDate($db.serverInfo.build_date)
</script>

{#if $db.serverInfo}
  <div>
    {#if serverStartedAt}
      {formatDate(serverStartedAt)}
    {/if}
  </div>
  {#if $db.serverInfo.latest_release}
    <LatestRelease
      currentVersion={$db.serverInfo.version || '0.1.0' || 'development'}
      latest={$db.serverInfo.latest_release} />
  {/if}
  Database-size: {$db.serverInfo.database_size_str}
  <div>
    {#if buildDate}
      {formatDate(buildDate)}
    {/if}
  </div>
{/if}
