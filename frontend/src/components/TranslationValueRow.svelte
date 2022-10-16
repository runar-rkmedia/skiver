<script lang="ts">
  import TranslationValue from './TranslationValue.svelte'
  import { createEventDispatcher } from 'svelte'
  import { state } from '../state'
  import LocaleFlag from './LocaleFlag.svelte'
  /** Map by locale-id */

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

<tr
  class="locale-item"
  class:auto-translate={translationValue?.source === 'system-translator'}
  class:missing={!value}
  class:selected>
  <td>
    <LocaleFlag {locale} />
    {locale.title}</td>

  <td>
    <TranslationValue
      on:showForm
      on:complete
      {translation}
      {categoryKey}
      {projectKey}
      {locale}
      {contextKey}
      {translationValue} />
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
</style>
