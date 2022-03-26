<script lang="ts">
  export let input: string
  export let key: string
  export let locale: string
  export let variables: Record<string, any> | undefined | null
  export let ns: string
  import { t } from '../util/i18next'
  import type { TFunction } from 'i18next'
  import Alert from './Alert.svelte'
  let myT: TFunction
  // $: myT = t.i18next.getFixedT(locale, '__derived__' + ns)
  $: {
    if (ns && locale && key) {
      // TODO: we might need to debounce this...
      t.i18next.addResource(locale, '__derived__' + ns, key, input)
      myT = t.i18next.getFixedT(locale, '__derived__' + ns)
      const b = t.i18next.getResourceBundle(locale, '__derived__' + ns)
    }
  }
  function handleVariables(variables: Record<string, any> | undefined) {
    if (!variables) {
      return variables
    }
    return Object.entries(variables).reduce((previous, [k, v]) => {
      if (typeof v === 'string' && /\d{4}-\d{2}-\d{2}/.test(v)) {
        const d = new Date(v)
        v = d
      }

      previous[k] = v

      return previous
    }, {})
  }
</script>

{#if myT}
  <div>
    <div />

    <span>
      {myT(key, handleVariables(variables))}
    </span>
  </div>
{/if}

{#if !ns}
  <Alert kind="warning">Cannot preview without namespace</Alert>
{:else if !locale}
  <Alert kind="warning">Cannot preview without locale</Alert>
{/if}
