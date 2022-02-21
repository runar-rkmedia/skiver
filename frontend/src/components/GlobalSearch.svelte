<script lang="ts">
  import { db } from 'api'

  import GlobalSearchInner from './GlobalSearchInner.svelte'
  let query = ''
  export let project: ApiDef.Project
  let visible = false
</script>

<h2>
  Search
  <input
    placeholder="John Doe"
    type="search"
    bind:value={query}
    on:focus={() => (visible = true)} />
</h2>
{#if project && visible && !Object.values($db.responseStates).some((rs) => rs.loading)}
  <GlobalSearchInner {project} {query} />
{/if}
