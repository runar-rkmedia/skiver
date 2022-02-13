<script lang="ts">
  import TranslationValueForm from 'forms/TranslationValueForm.svelte'
  import { createEventDispatcher } from 'svelte'
  import { state } from '../state'
  import ApiResponseError from './ApiResponseError.svelte'
  import EntityDetails from './EntityDetails.svelte'
  import Icon from './Icon.svelte'
  import TranslationPreview from './TranslationPreview.svelte'
  /** Map by locale-id */

  let selectedLocale: string
  const dispatch = createEventDispatcher()
  export let translation: ApiDef.Translation
  export let categoryKey: string
  export let projectKey: string
  export let locale: ApiDef.Locale
  export let contextKey = ''
  export let translationValue: ApiDef.TranslationValue | undefined
  $: value = contextKey
    ? translationValue?.context?.[contextKey]
    : translationValue?.value
</script>

<tr
  class="locale-item"
  class:auto-translate={translationValue?.source === 'system-translator'}
  class:missing={!value}
  class:selected={selectedLocale === locale.id}>
  <td>{locale.title}</td>
  <td>
    {#if selectedLocale === locale.id}
      {#if translationValue?.source === 'system-translator'}
        <p>
          <Icon icon="warning" color="warning" />
          This value was auto-translated.
        </p>
      {/if}
      <!-- {#if translationValue.source === 'system-translator'} -->
      <TranslationValueForm
        existingID={translationValue?.id}
        localeID={locale.id}
        translationID={translation.id}
        on:complete={() => {
          dispatch('complete')
          selectedLocale = ''
        }}
        on:cancel={() => {
          selectedLocale = ''
          dispatch('showForm', { locale, show: false })
        }}>
        <div slot="preview">
          {#if locale && categoryKey && projectKey && translation}
            <TranslationPreview
              bind:locale={locale.id}
              bind:ns={projectKey}
              key={(categoryKey === '' ? '' : categoryKey + '.') +
                translation.key +
                (translation.context ? '_' + translation.context : '')}
              bind:variables={translation.variables}
              bind:input={$state.createTranslationValue.value} />
          {/if}
        </div>
      </TranslationValueForm>
      {#if translationValue}
        <EntityDetails entity={translationValue} />
      {/if}
      <ApiResponseError key="translationValue" />
    {:else}
      <div
        class="keep-whitespace click-to-edit"
        on:click={() => {
          dispatch('showForm', { locale, show: true })
          selectedLocale = locale.id
          if (value) {
            $state.createTranslationValue.value = value
          }
        }}>
        {#if translationValue?.source === 'system-translator'}
          <Icon icon="warning" color="warning" />
        {/if}
        <Icon icon="edit" color="primary" />
        {value || '<no value>'}
      </div>
    {/if}
  </td>
</tr>

<style>
  .selected {
    outline: 1px dashed hotpink;
  }
  .locale-item td:first-of-type {
    white-space: nowrap;
    width: min-content;
  }
  td {
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
