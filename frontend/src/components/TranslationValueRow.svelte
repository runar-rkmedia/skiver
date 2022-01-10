<script lang="ts">
  import TranslationValueForm from 'forms/TranslationValueForm.svelte'
  import { createEventDispatcher } from 'svelte'
  import { state } from '../state'
  import ApiResponseError from './ApiResponseError.svelte'
  import EntityDetails from './EntityDetails.svelte'
  import Icon from './Icon.svelte'
  /** Map by locale-id */

  let selectedLocale: string = ''
  const dispatch = createEventDispatcher()
  export let translationID: string
  export let locale: ApiDef.Locale
  export let translationValue: ApiDef.TranslationValue | undefined
</script>

<tr
  class="locale-item"
  class:auto-translate={translationValue?.source === 'system-translator'}
  class:missing={!translationValue?.value}
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
        {translationID}
        on:complete={() => {
          dispatch('complete')
          selectedLocale = ''
        }}
        on:cancel={() => {
          selectedLocale = ''
          dispatch('showForm', { locale, show: false })
        }} />
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
          const value = translationValue?.value
          if (value) {
            $state.createTranslationValue.value = value
          }
        }}>
        {#if translationValue?.source === 'system-translator'}
          <Icon icon="warning" color="warning" />
        {/if}
        <Icon icon="edit" color="primary" />
        {translationValue?.value || '<no value>'}
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
