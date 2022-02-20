<script lang="ts">
  import TranslationValueRow from './TranslationValueRow.svelte'
  import EntityDetails from './EntityDetails.svelte'
  import Alert from './Alert.svelte'
  import { api, db } from 'api'
  import Button from './Button.svelte'
  export let translation: ApiDef.Translation
  /** Map by locale-id */
  // export let translationValues: Record<string, ApiDef.TranslationValue>
  export let categoryKey: string
  export let projectKey: string

  export let locales: ApiDef.Locale[]
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
  let edit = false
  let editTitle = ''
  let editDescription = ''
  let editVariables = ''
  let editError = ''
  function toggleEdit() {
    edit = !edit
    if (!editTitle) {
      editTitle = translation.title || ''
      editDescription = translation.description || ''
      editVariables =
        JSON.stringify(translation.variables, null, 2) || '{\n  \n}'
    }
  }

  function submitEdit() {
    let payload: Partial<ApiDef.TranslationInput> = {
      title: editTitle,
      description: editDescription,
    }
    if (editVariables && editVariables.replace(/\s/g, '') !== '{}')
      try {
        payload.variables = JSON.parse(editVariables)
      } catch (err) {
        editError = 'Invalid variables: ' + err.message
        return
      }
    editError = ''

    api.translation
      .update(translation.id, payload as any)
      .then(() => (edit = false))
  }
</script>

{#if translation}
  {#if edit}
    <form>
      <label
        >Title<input name="title" bind:value={editTitle} type="text" /></label>
      <label
        >Description<textarea
          name="description"
          rows={3}
          bind:value={editDescription}
          type="text" /></label>
      {#if editError}
        <Alert kind="error">
          {editError}
        </Alert>
      {/if}
      <label
        >Variables<textarea
          name="variables"
          rows={8}
          bind:value={editVariables}
          type="text" /></label>
      <Button color="primary" icon="submit" on:click={submitEdit}
        >Submit</Button>
      <Button color="secondary" icon="cancel" on:click={toggleEdit}
        >Cancel</Button>
    </form>
  {:else}
    <div class="desc">
      <h4>
        <code>
          {`${categoryKey !== '' ? categoryKey + '.' : ''}${translation.key}`}
        </code>
        {translation.title}
        <Button icon="edit" on:click={toggleEdit}>Edit</Button>
      </h4>
      <div>
        <small />
        <EntityDetails entity={translation} />
      </div>
    </div>
  {/if}
  <p>
    {translation.description || ''}
  </p>
  <table>
    <thead>
      <th>Language</th>
      <th>Value</th>
    </thead>
    <tbody>
      {#if locales && translationValues}
        {#each locales as locale}
          <TranslationValueRow
            {categoryKey}
            {projectKey}
            {translation}
            {locale}
            translationValue={translationValues[locale.id]} />
        {/each}
      {/if}
    </tbody>
  </table>

  {#if contextKeys && contextKeys.length}
    <hr />
    <h5>Contexts</h5>
    <p>
      Contexts are variations of the default value, often used programmatically.
      If not set, the value will typically fall back to the default value
    </p>

    {#each contextKeys as contextKey}
      <h6>
        {contextKey}
      </h6>
      <table>
        <thead>
          <th>Language</th>
          <th>Value</th>
        </thead>
        <tbody>
          {#each locales as locale}
            <TranslationValueRow
              {categoryKey}
              {projectKey}
              {translation}
              {contextKey}
              {locale}
              translationValue={translationValues[locale.id]} />
          {/each}
        </tbody>
      </table>
    {/each}
  {/if}
  {#if translation?.variables}
    <h6>Variables</h6>
    {#each Object.entries(translation.variables) as [k, v]}
      <var-pair>
        <var-key>{k}</var-key>
        <var-value>{JSON.stringify(v)}</var-value>
      </var-pair>
      <!-- TODO: allow editing -->
    {/each}
  {/if}
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
  var-pair {
    display: block;
  }
  var-key::after {
    content: ': ';
  }
  textarea[name='variables'] {
    font-family: var(--font-mono);
  }
</style>
