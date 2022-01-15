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
  $: {
    if (ns && locale && input) {
      console.log('updating input', t.i18next)
      t.i18next.addResource(locale, '__derived__' + ns, key, input)
      myT = t.i18next.getFixedT(locale, '__derived__' + ns)
    }
  }

  const defaultVariables = {
    count: 42,
    total: 42,
    year: `{{now | formatDate "2006" }}`,
    color: 'blue',
    colour: 'green',
    error: 'Simulated error',
    errormessage: 'Simulated error',
    errormsg: 'Simulated error',
    regionname: 'Gigantis',
    region: 'Gigantis',
    country: 'Japan',
    countryname: 'Japan',
    companyname: 'Wily',
    name: 'Douglas',
    price: 'Ƶ5000',
    email: 'roll@example.com',
    date: new Date('1987-12-17T06:00:00-09:00'),
    expires: new Date('1995-03-24T06:00:00-09:00'),
    days: 6,
  }
</script>

{#if myT}
  <div>
    <div />

    <span>
      {myT(key, { defaultVariables, ...variables })}
    </span>
  </div>
{/if}
<input bind:value={input} />

{#if !ns}
  <Alert kind="warning">Cannot preview without namespace</Alert>
{:else if !locale}
  <Alert kind="warning">Cannot preview without locale</Alert>
{:else if !input}
  <Alert kind="warning">Cannot preview without input</Alert>
{/if}
