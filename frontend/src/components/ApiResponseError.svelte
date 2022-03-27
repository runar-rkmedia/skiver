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
  {#if s.error?.error}
    <Alert kind="error">
      {#if s.error.error.code}
        <h5>{s.error.error.code}</h5>
      {/if}
      <p>{s.error.error.error}</p>
      {#if s.error.details}
        {JSON.stringify(s.error.details)}
      {/if}
    </Alert>
  {/if}
{/if}
