<script lang="ts">
  import TranslationValue from './TranslationValue.svelte'
  import { db } from 'api'
  import Icon from './Icon.svelte'
  import { showDialog } from 'state'
  import { fade, fly, scale, slide } from 'svelte/transition'
  export let translation: ApiDef.Translation
  export let columns: {
    title: boolean
    key: boolean
    valueForLocale: ApiDef.Locale[]
  }
  /** Map by locale-id */
  // export let translationValues: Record<string, ApiDef.TranslationValue>
  export let categoryKey: string
  export let projectKey: string

  $: translationValues = (translation?.value_ids || []).reduce((r, tvid) => {
    const tv = $db.translationValue[tvid]
    if (!tv || !tv.locale_id) {
      return r
    }
    r[tv.locale_id] = tv
    return r
  }, {} as Record<string, ApiDef.TranslationValue>)
  let contextKeys: string[] = []
  $: contextKeys = Array.from(
    Object.values(translationValues || {}).reduce((r, tv) => {
      if (!tv || !tv.context) {
        return r
      }
      for (const c of Object.keys(tv.context)) {
        r.add(c)
      }
      return r
    }, new Set<string>())
  )
  let addContext = false
  let contextToAdd = ''
  $: contextToAddIsValid =
    contextToAdd.trim().length > 0 &&
    !contextKeys.some((c) => contextToAdd === c)
  function onShowContextForm() {
    addContext = !addContext
    contextToAdd = ''
  }
  function onEditTranslation() {
    showDialog({
      kind: 'editTranslation',
      id: translation.id,
      title: `Edit ${translation.title}`,
    })
  }
</script>

<tr class:deleted={translation.deleted}>
  {#if columns.title}
    <td class="title" title={translation.description || ''}>
      <div class="withActions">
        <div class="title">
          {translation.title}
        </div>
        <div class="actions">
          <button
            class="btn-reset"
            title="Edit in modal"
            on:click={onEditTranslation}>
            <Icon icon="edit" color="inherit" noMinWidth />
          </button>
          {#if !translation.deleted}
            <button
              disabled={!!translation.deleted}
              class="btn-reset"
              title="Add context"
              on:click={onShowContextForm}>
              <Icon icon="createContext" color="inherit" noMinWidth />
            </button>
          {/if}
        </div>
      </div>
    </td>
  {/if}
  {#if columns.key}
    <td
      class="key"
      title={[translation.title, translation.description]
        .filter(Boolean)
        .join('\n\n')}>
      <div class="withActions">
        <div class="title">
          {translation.key}
        </div>
        {#if !columns.title}
          <div class="actions">
            <button
              class="btn-reset"
              title="Edit in modal"
              on:click={onEditTranslation}>
              <Icon icon="edit" color="inherit" noMinWidth />
            </button>
            {#if !translation.deleted}
              <button
                disabled={!!translation.deleted}
                class="btn-reset"
                title="Add context"
                on:click={onShowContextForm}>
                <Icon icon="createContext" color="inherit" noMinWidth />
              </button>
            {/if}
          </div>
        {/if}
      </div>
    </td>
  {/if}
  {#if columns.valueForLocale}
    {#each columns.valueForLocale as locale}
      <td class="value">
        {#if !translation.deleted}
          <TranslationValue
            on:showForm
            on:complete
            {translation}
            {categoryKey}
            {projectKey}
            {locale}
            translationValue={translationValues[locale.id]} />
        {:else}
          {translationValues[locale.id]?.value}
        {/if}
      </td>
    {/each}
  {/if}
</tr>
{#if addContext}
  <tr>
    <td class="key">
      <input
        placeholder="Key for context"
        type="context"
        bind:value={contextToAdd}
        autofocus />
    </td>
    {#if columns.title && columns.key}
      <td />
    {/if}
    {#if contextToAddIsValid}
      {#each columns.valueForLocale as locale}
        <td class="value">
          <TranslationValue
            on:showForm
            on:complete={(e) => {
              addContext = false
            }}
            contextKey={contextToAdd || '__non_existant'}
            {translation}
            {categoryKey}
            {projectKey}
            {locale}
            translationValue={translationValues[locale.id]} />
        </td>
      {/each}
    {/if}
  </tr>
{/if}

{#each contextKeys as context}
  <tr class:context>
    {#if columns.title}
      <td class="title" title={context}>
        <Icon icon="context" />
        {context}</td>
    {/if}
    {#if columns.key}
      <td class="key">
        {#if !columns.title}
          <Icon icon="context" />
        {/if}
        {context}
      </td>
    {/if}
    {#if columns.valueForLocale}
      {#each columns.valueForLocale as locale}
        <td class="value">
          <TranslationValue
            on:showForm
            on:complete
            {translation}
            {categoryKey}
            {projectKey}
            contextKey={context}
            {locale}
            translationValue={translationValues[locale.id]} />
        </td>
      {/each}
    {/if}
  </tr>
{/each}

<style>
  .context .key,
  .context .title {
    padding-inline-start: var(--size-2);
    color: var(--color-blue-700);
    font-size: 0.9em;
  }
  td:not(.value) {
    white-space: nowrap;
  }
  td.value {
    width: 100%;
  }
  .withActions {
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: nowrap;
  }
  td.key input,
  td.key .title {
    font-family: var(--font-mono);
  }
  .withActions button {
    min-width: var(--size-8);
    text-align: center;
    background-color: var(--color-purple-500);
    border-radius: var(--radius-md);
    color: var(--color-grey-100);
    transition: transform 150ms var(--easing-standard);
  }
  .withActions button:hover() {
    transform: scale(1.15);
  }
  .withActions button:nth-child(2) {
    background-color: var(--color-blue-700);
  }
  .deleted {
    text-decoration: line-through;
    background-image: repeating-linear-gradient(
      45deg,
      var(--color-grey-100),
      var(--color-grey-100) 30px,
      var(--color-grey-300) 30px,
      var(--color-grey-300) 60px
    );
  }
</style>
