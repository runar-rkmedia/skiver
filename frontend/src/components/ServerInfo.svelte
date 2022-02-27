<script lang="ts">
  import { db } from '../api'

  import formatDate, { isValidDate } from '../dates'

  $: serverStartedAt = isValidDate($db.serverInfo.server_started_at)
  $: buildDate = isValidDate($db.serverInfo.build_date)
</script>

<div>
  {#if serverStartedAt}
    {formatDate(serverStartedAt)}
  {/if}
</div>
<div>{$db.serverInfo.version} ({$db.serverInfo.version || 'development'})</div>
Database-size: {$db.serverInfo.database_size_str}
<div>
  {#if buildDate}
    {formatDate(buildDate)}
  {/if}
</div>
