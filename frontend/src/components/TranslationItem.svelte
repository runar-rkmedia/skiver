<script lang="ts">
  import { db } from 'api'

  import { createEventDispatcher } from 'svelte'
  import { state } from '../state'
  import TranslationValueRow from './TranslationValueRow.svelte'
  import Alert from './Alert.svelte'
  import EntityDetails from './EntityDetails.svelte'
  import Icon from './Icon.svelte'
  export let translation: ApiDef.Translation
  /** Map by locale-id */
  export let translationValues: Record<string, ApiDef.TranslationValue>
  export let categoryKey: string
  let showForm: boolean = true

  let selectedLocale: string = ''
  export let locales: ApiDef.Locale[]
  const dispatch = createEventDispatcher()
  $: errs = Object.entries($db.responseStates.translationValue).reduce(
    (r, [k, v]) => {
      if (v && typeof v === 'object' && v.error) {
        r.push(v)
      }
      return r
    },
    [] as ApiDef.APIError[]
  )
</script>

{#if translation}
  <div class="desc">
    <h4>
      <code>
        {`${categoryKey !== "___root___" ? categoryKey + "." : ''}${translation.key}`}
      </code>
      {translation.title}
    </h4>
    <div>
      <small>
        {translation.description || ''}
      </small>
      <EntityDetails entity={translation} />
    </div>
  </div>
  <table>
    <thead>
      <th>Language</th>
      <th>Value</th>
    </thead>
    <tbody>
      {#if locales}
        {#each locales as locale}
          <TranslationValueRow
            translationID={translation.id}
            {locale}
            translationValue={translationValues[locale.id]} />
        {/each}
      {/if}
    </tbody>
  </table>
{:else}
  ... (no translation???)
{/if}

<style>
  .desc {
    display: flex;
    justify-content: space-between;
  }
  .desc > * {
    display: block;
  }
  .desc > :last-child {
    text-align: right;
  }
  th {
    padding-inline: var(--size-4);
    padding-block: var(--size-2);
  }
</style>
