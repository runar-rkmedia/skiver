<script lang="ts">
  import { db } from 'api'

  import Button from 'components/Button.svelte'
  import TranslationValueForm from 'forms/TranslationValueForm.svelte'
  import { createEventDispatcher } from 'svelte'
  import { state } from '../state'
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
        {categoryKey}.{translation.key}
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
          <tr
            class="locale-item"
            class:auto-translate={translationValues?.[locale.id]?.source ===
              'system-translator'}
            class:missing={!translationValues?.[locale.id]?.value}
            class:selected={showForm && selectedLocale === locale.id}>
            <td>{locale.title}</td>
            <td>
              {#if showForm && selectedLocale === locale.id}
                {#if translationValues?.[locale.id]?.source === 'system-translator'}
                  <p>
                    <Icon icon="warning" color="warning" />
                    This value was auto-translated.
                  </p>
                {/if}
                <!-- {#if translationValues?.[locale.id]?.source === 'system-translator'} -->
                <TranslationValueForm
                  existingID={translationValues?.[locale.id]?.id}
                  localeID={locale.id}
                  translationID={translation.id}
                  on:complete={() => {
                    dispatch('complete')
                    selectedLocale = ''
                  }}
                  on:cancel={() => {
                    selectedLocale = ''
                    dispatch('showForm', { locale, show: false })
                  }} />
                <EntityDetails entity={translationValues?.[locale.id]} />
                {#each errs as v}
                  <Alert kind="error">
                    <h5>{v.code}</h5>
                    <p>{v.error}</p>
                    {#if v.details}
                      {JSON.stringify(v.details)}
                    {/if}
                  </Alert>
                {/each}
              {:else}
                <div
                  class="keep-whitespace click-to-edit"
                  on:click={() => {
                    dispatch('showForm', { locale, show: true })
                    selectedLocale = locale.id
                    const value = translationValues?.[locale.id]?.value
                    if (value) {
                      $state.createTranslationValue.value = value
                    }
                  }}>
                  {#if translationValues?.[locale.id]?.source === 'system-translator'}
                    <Icon icon="warning" color="warning" />
                  {/if}
                  <Icon icon="edit" color="primary" />
                  {translationValues?.[locale.id]?.value || '<no value>'}
                </div>
              {/if}
            </td>
          </tr>
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
  .selected {
    outline: 1px dashed hotpink;
  }
  .locale-item td:first-of-type {
    white-space: nowrap;
    width: min-content;
  }
  td,
  th {
    padding-inline: var(--size-4);
    padding-block: var(--size-2);
  }
  .locale-item td:not(:first-of-type) {
    width: 100%;
  }
  .missing {
    background-color: var(--color-danger-300);
  }
  .missing:nth-child(even) {
    background-color: var(--color-danger);
  }
  .click-to-edit {
    cursor: pointer;
  }
  .keep-whitespace {
    white-space: pre-line;
  }
</style>
