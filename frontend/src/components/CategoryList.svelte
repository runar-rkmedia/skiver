<script lang="ts">
  import { db } from 'api'
  import CategoryForm from 'forms/CategoryForm.svelte'
  import CategoryItem from './CategoryItem.svelte'
  import sortOn from 'sort-on'
  import { showDialog, state } from 'state'
  import Button from './Button.svelte'
  import Dialog from './Dialog.svelte'
  import TranslationForm from 'forms/TranslationForm.svelte'
  import Embed from 'pages/Embed.svelte'
  import ScrollAnchor from './ScrollAnchor.svelte'

  export let locales: ApiDef.Locale[]
  export let selectedLocale: string
  export let projectID: string
  export let selectedTranslation: string
  export let selectedCategory: string
  export let visibleForm: string | null
  let expandedCategory = ''
  $: categories =
    !!projectID &&
    sortOn(
      Object.values($db.category).filter((c) => c.project_id === projectID),
      ($state.categorySortAsc ? '' : '-') + $state.categorySortOn
    )
  function closeDialog() {
    visibleForm = null
    $state.dialog = null
  }
</script>

{#if visibleForm === 'editCategory'}
  <Dialog on:clickClose={closeDialog}>
    <span slot="title">Edit Category</span>
    <paper>
      <CategoryForm
        {projectID}
        categoryID={selectedCategory}
        on:complete={closeDialog}>
        <Button
          slot="actions"
          color="secondary"
          icon="cancel"
          on:click={closeDialog}>Cancel</Button>
      </CategoryForm>
    </paper>
  </Dialog>
{/if}
{#if $state.dialog}
  <Dialog on:clickClose={closeDialog}>
    {#if $state.dialog.kind === 'createTranslation' && $state.dialog.parent}
      <Dialog
        on:closeClick={() => {
          showDialog(null)
        }}>
        <paper>
          <TranslationForm
            categoryID={selectedCategory}
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
    {:else if $state.dialog.kind === 'createCategory'}
      <Dialog on:clickClose={closeDialog}>
        <span slot="title">Edit Category</span>
        <paper>
          <CategoryForm
            {projectID}
            categoryID={selectedCategory}
            on:complete={closeDialog}>
            <Button
              slot="actions"
              color="secondary"
              icon="cancel"
              on:click={closeDialog}>Cancel</Button>
          </CategoryForm>
        </paper>
      </Dialog>
    {:else if $state.dialog.kind === 'translation' && $state.dialog.id && $state.dialog.parent}
      <Dialog on:clickClose={closeDialog}>
        <span slot="title">{$state.dialog.title || ''}</span>
        <paper>
          <Embed
            noHeader={true}
            categoryKey={$state.dialog.parent}
            projectKey={projectID}
            translationKeyLike={$state.dialog.id}>
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
{#if categories && categories?.length}
  {#each categories as category (category.id)}
    <CategoryItem
      bind:category
      bind:locales
      bind:projectKey={projectID}
      bind:selectedLocale
      bind:selectedTranslation
      bind:selectedCategory
      bind:expandedCategory
      forceExpand={false}
      bind:visibleForm />
  {/each}
{/if}
