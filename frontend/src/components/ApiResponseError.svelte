<script lang="ts">
  import { db, type DB } from 'api'
  import Alert from './Alert.svelte'

  // $: errs = Object.entries($db.responseStates.translationValue).reduce(
  //   (r, [k, v]) => {
  //     if (v && typeof v === 'object' && v.error) {
  //       r.push(v)
  //     }
  //     return r
  //   },
  //   [] as ApiDef.APIError[]
  // )
  export let key: keyof DB['responseStates']
  $: s = $db.responseStates[key]
</script>

{#if s && !s.loading && s.error}
  <Alert kind="error">
    <h5>{s.error.code}</h5>
    <p>{s.error.error}</p>
    {#if s.error.details}
      {JSON.stringify(s.error.details)}
    {/if}
  </Alert>
{/if}
