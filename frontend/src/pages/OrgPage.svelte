<script lang="ts">
  import { db } from '../api'
  export let organization: ApiDef.Organization
</script>

Org

{#if !organization}
  Not found
{:else}
  <pre>{JSON.stringify(organization, null, 2)}</pre>
{/if}
