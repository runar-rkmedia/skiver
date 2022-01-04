<script lang="ts">
  import Button from 'components/Button.svelte'
  import TranslationValueForm from 'forms/TranslationValueForm.svelte'
  import { createEventDispatcher } from 'svelte'
  import { state } from '../state'
  import Icon from './Icon.svelte'
  export let translation: ApiDef.Translation
  /** Map by locale-id */
  export let translationValues: Record<string, ApiDef.TranslationValue>
  export let categoryKey: string
  export let showForm: boolean
  export let selectedLocale: string = ''
  export let locales: ApiDef.Locale[]
  const dispatch = createEventDispatcher()
</script>

{#if translation}
  <div>
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
                <TranslationValueForm
                  localeID={locale.id}
                  translationID={translation.id}
                  on:complete={() => dispatch('complete')}>
                  <Button
                    slot="actions"
                    color="secondary"
                    on:click={() => {
                      dispatch('showForm', { locale, show: false })
                    }}
                    icon={'cancel'}>Cancel</Button>
                </TranslationValueForm>
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
