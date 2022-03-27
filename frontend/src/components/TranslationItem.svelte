<script lang="ts">
  import TranslationValueRow from './TranslationValueRow.svelte'
  import EntityDetails from './EntityDetails.svelte'
  import Alert from './Alert.svelte'
  import { api, db } from 'api'
  import Button from './Button.svelte'
  import ApiResponseError from './ApiResponseError.svelte'
  import ContextForm from './ContextForm.svelte'
import Dialog from './Dialog.svelte';
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
  let showDelete = false
  let editTitle = ''
  let editKey = ''
  let editDescription = ''
  let editVariables = ''
  let editError = ''
  let addContext = false
  function confirmDelete() {
    showDelete = true
    edit = false
  }
  function toggleEdit() {
    edit = !edit
    if (!editTitle) {
      editTitle = translation.title || ''
      editKey = translation.key || ''
      editDescription = translation.description || ''
      editVariables =
        JSON.stringify(translation.variables, null, 2) || '{\n  \n}'
    }
  }

  function submitEdit() {
    let payload: ApiDef.UpdateTranslationInput = {
      id: translation.id,
      ...(!!editTitle &&
        editTitle !== translation.title && { title: editTitle }),
      ...(!!editKey && editKey !== translation.key && { key: editKey }),
      ...(!!editDescription &&
        editDescription !== translation.description && {
          description: editDescription,
        }),
    }
    if (editVariables && editVariables.replace(/\s/g, '') !== '{}')
      try {
        payload.variables = JSON.parse(editVariables)
      } catch (err) {
        editError = 'Invalid variables: ' + err.message
        return
      }
    editError = ''

    api.translation.update(translation.id, payload).then(([_, err]) => {
      if (err) {
        return
      }
      edit = false
    })
  }
</script>

{#if translation}
  {#if showDelete}
    <form>
      <ApiResponseError key="translation" />
      <Alert kind="warning">
        <h3 slot="title">Are you sure you want to delete this translation?</h3>
        <p>The translation can be restored at a later time</p>
        <p>Deleted translations are not visible, and will not be exported</p>
        <p>
          Snapshots created before the deletion will still include the
          translation. New snapshots will not include it
        </p>
        <Button
          on:click={() =>
            api.translation
              .delete(translation.id, {
                undelete: false,
              })
              .then(([_, err]) => {
                if (!err) {
                  showDelete = false
                }
              })}
          color="danger">Yes, I am sure</Button>
        <Button
          color="secondary"
          on:click={() => {
            showDelete = false
          }}>No, not at this time</Button>
      </Alert>
    </form>
  {/if}
  {#if edit}
  <Dialog on:clickClose={toggleEdit}>
    <span slot="title">Edit Category</span>
    <paper>
    <form>
      <ApiResponseError key="translation" />
      {#if editError}
        <Alert kind="error">
          {editError}
        </Alert>
      {/if}
      <label
        >Title<input name="title" bind:value={editTitle} type="text" /></label>
      <label
        >Description<textarea
          name="description"
          rows={3}
          bind:value={editDescription}
          type="text" /></label>
      <label>Key<input name="key" bind:value={editKey} type="text" /></label>
      <label>
        Variables
        <textarea
          name="variables"
          rows={8}
          bind:value={editVariables}
          type="text" /></label>
      <div class="buttonRow">
        <Button color="primary" icon="submit" on:click={submitEdit}
          >Submit</Button>
        <Button color="secondary" icon="cancel" on:click={toggleEdit}
          >Cancel</Button>
        <div class="deleteButton">
          {#if translation.deleted}
            <Button
              color="primary"
              icon="delete"
              on:click={() =>
                api.translation.delete(translation.id, { undelete: true })}
              >Undelete</Button>
          {:else}
            <Button color="danger" icon="delete" on:click={confirmDelete}
              >Delete</Button>
          {/if}
        </div>
      </div>
    </form>
  </paper>
      </Dialog>
  {/if}
    <div class="desc">
      <h4>
        <slot name="categoryHeader" {categoryKey} {translation}>
          <code>
            {`${categoryKey !== '' ? categoryKey + '.' : ''}${translation.key}`}
          </code>
          {translation.title}
        </slot>
        <Button icon="edit" on:click={toggleEdit}>Edit</Button>
      </h4>
      <div>
        <small />
        <EntityDetails entity={translation} />
      </div>
    </div>
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

  <hr />
  <h5>
    Contexts

    {#if !addContext}
      <Button icon="create" on:click={() => (addContext = true)}>Add</Button>
    {/if}
  </h5>

  <!-- <p> -->
  <!--   Contexts are variations of the default value, often used programmatically. -->
  <!--   If not set, the value will typically fall back to the default value -->
  <!-- </p> -->

  {#if addContext}
    <paper>
      <ContextForm
        {categoryKey}
        {projectKey}
        {translation}
        {locales}
        {translationValues}
        on:complete={() => (addContext = false)}
        on:abort={() => (addContext = false)} />
    </paper>
  {/if}
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
  {#if translation?.variables}
    <h6>Variables</h6>
    <code>{JSON.stringify(translation.variables, null, 2)}</code>
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
  textarea[name='variables'] {
    font-family: var(--font-mono);
  }
  code {
    white-space: pre;
    display: inline-block;
  }
  .buttonRow {
    display: flex;
  }
  .deleteButton {
    align-self: flex-end;
  }
</style>
