<script lang="ts">
  import CategoryForm from 'forms/CategoryForm.svelte'

  import TranslationForm from 'forms/TranslationForm.svelte'
  import Embed from 'pages/Embed.svelte'

  import { closeDialog, showDialog, state } from 'state'
  import Button from './Button.svelte'
  import Dialog from './Dialog.svelte'
  import ScrollAnchor from './ScrollAnchor.svelte'

  export let projectID: string
</script>

{#if $state.dialog}
  <Dialog on:clickClose={closeDialog}>
    {#if $state.dialog.kind === 'createTranslation' && $state.dialog.parent}
      <Dialog
        on:closeClick={() => {
          showDialog(null)
        }}>
        <paper>
          <TranslationForm
            categoryID={$state.dialog.id || ''}
            on:complete={closeDialog}>
            <Button
              slot="actions"
              color="secondary"
              icon={'cancel'}
              on:click={() => {
                showDialog(null)
              }}>
              Cancel
            </Button>
          </TranslationForm>
        </paper>
      </Dialog>
    {:else if $state.dialog.kind === 'editCategory'}
      <Dialog on:clickClose={closeDialog}>
        <span slot="title">Edit Category</span>
        <paper>
          <CategoryForm
            {projectID}
            categoryID={$state.dialog.id}
            on:complete={closeDialog}>
            <Button
              slot="actions"
              color="secondary"
              icon="cancel"
              on:click={closeDialog}>Cancel</Button>
          </CategoryForm>
        </paper>
      </Dialog>
    {:else if $state.dialog.kind === 'createCategory'}
      <Dialog on:clickClose={closeDialog}>
        <span slot="title">Edit Category</span>
        <paper>
          <CategoryForm
            {projectID}
            categoryID={$state.dialog.id}
            on:complete={closeDialog}>
            <Button
              slot="actions"
              color="secondary"
              icon="cancel"
              on:click={closeDialog}>Cancel</Button>
          </CategoryForm>
        </paper>
      </Dialog>
    {:else if $state.dialog.kind === 'translation'}
      <Dialog on:clickClose={closeDialog}>
        <span slot="title">{$state.dialog.title || ''}</span>
        <paper>
          <Embed
            noHeader={true}
            categoryKey={$state.dialog.parent || ''}
            projectKey={projectID}
            translationKeyLike={$state.dialog.id || ''}>
            <h4 slot="categoryHeader" let:category let:translation>
              <ScrollAnchor
                {category}
                on:scrollTo={() => {
                  $state.dialog = null
                }}>
                <code>
                  {[category?.key, translation?.key].filter(Boolean).join('.')}
                </code>
              </ScrollAnchor>
            </h4>
          </Embed>
        </paper>
      </Dialog>
    {:else}
      <paper>
        <p>Unhandled dialog-option</p>
        <pre>{JSON.stringify($state.dialog, null, 2)}</pre>
        <p>Sorry for the inconvenience</p>
      </paper>
    {/if}
  </Dialog>
{/if}
