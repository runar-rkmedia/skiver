<script lang="ts">
  import { api } from 'api'

  import ApiResponseError from './ApiResponseError.svelte'
  import Alert from './Alert.svelte'
  import Dialog from './Dialog.svelte'
  import Button from './Button.svelte'
  import { closeDialog } from 'state'
  import { onMount } from 'svelte'
  export let translation: ApiDef.Translation
  let editError = ''
  let showDelete = false
  let editTitle = ''
  let editKey = ''
  let editDescription = ''
  let editVariables = ''

  onMount(() => {
    if (!translation) {
      return
    }
    editTitle = translation.title || ''
    editKey = translation.key || ''
    editDescription = translation.description || ''
    editVariables = JSON.stringify(translation.variables, null, 2) || '{\n  \n}'
  })

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
    })
    closeDialog()
  }
  function confirmDelete() {
    showDelete = true
  }
</script>

{#if translation}
  <Dialog on:clickClose={closeDialog}>
    <span slot="title">Edit translation </span>
    <paper>
      <form>
        <ApiResponseError key="translation" />
        {#if editError}
          <Alert kind="error">
            {editError}
          </Alert>
        {/if}
        <label
          >Title<input
            name="title"
            bind:value={editTitle}
            type="text" /></label>
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
          <Button color="secondary" icon="cancel" on:click={closeDialog}
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
        {#if showDelete}
          <form>
            <ApiResponseError key="translation" />
            <Alert kind="warning">
              <h3 slot="title">
                Are you sure you want to delete this translation?
              </h3>
              <p>The translation can be restored at a later time</p>
              <p>
                Deleted translations are not visible, and will not be exported
              </p>
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
      </form>
    </paper>
  </Dialog>
{/if}

<style>
  textarea[name='variables'] {
    font-family: var(--font-mono);
  }
  .buttonRow {
    display: flex;
  }
  .deleteButton {
    align-self: flex-end;
  }
</style>
