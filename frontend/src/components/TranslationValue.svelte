<script lang="ts">
  import TranslationValueForm from 'forms/TranslationValueForm.svelte'
  import { createEventDispatcher } from 'svelte'
  import { state } from '../state'
  import ApiResponseError from './ApiResponseError.svelte'
  import EntityDetails from './EntityDetails.svelte'
  import Icon from './Icon.svelte'
  import TranslationPreview from './TranslationPreview.svelte'
  /** Map by locale-id */

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
  $: selected =
    $state.openTranslationValueForm === translation.id + locale.id + contextKey
</script>

{#if selected}
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
    {contextKey}
    on:complete={() => {
      dispatch('complete')
      $state.openTranslationValueForm = ''
    }}
    on:cancel={() => {
      dispatch('showForm', { locale, show: false })
    }}>
    <div slot="preview">
      {#if locale && categoryKey && projectKey && translation}
        <TranslationPreview
          bind:locale={locale.id}
          bind:ns={projectKey}
          key={(categoryKey === '' ? '' : categoryKey + '.') + translation.key}
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
      {
        $state.createTranslationValue.value = value
      }

      {
        $state.createTranslationValue.context_key = contextKey
      }
      $state.openTranslationValueForm = translation.id + locale.id + contextKey
    }}>
    {#if translationValue?.source === 'system-translator'}
      <Icon icon="warning" color="warning" />
    {/if}
    <Icon icon="edit" color="primary" />
    {value || '<no value>'}
  </div>
{/if}

<style>
  .click-to-edit {
    cursor: pointer;
  }
  .keep-whitespace {
    white-space: pre-line;
  }
</style>
