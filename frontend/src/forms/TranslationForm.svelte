<script lang="ts">
  import Button from 'components/Button.svelte'
  import { createEventDispatcher } from 'svelte'
  import { api } from '../api'
  import { state, toast } from '../state'
  export let categoryID: string
  const dispatch = createEventDispatcher()
  async function onCreateTranslation() {
    $state.createTranslation.category_id = categoryID
    if (!$state.createTranslation.category_id) {
      toast({
        title: 'missing argument',
        message: 'category was not set',
        kind: 'warning',
      })
      return
    }
    const s = await api.translation.create($state.createTranslation)
    if (!s[1]) {
      dispatch('complete', s[0])
      $state.createTranslation = { category_id: '', key: '' }
    }
  }
</script>

<form id={'form-' + categoryID}>
  <label>
    Key
    <input name="key" bind:value={$state.createTranslation.key} />
  </label>
  <label>
    Title
    <input name="title" bind:value={$state.createTranslation.title} />
  </label>
  <label>
    Description (Optional, but recommended)
    <textarea
      name="description"
      rows="5"
      bind:value={$state.createTranslation.description} />
  </label>
  <Button
    color="primary"
    type="submit"
    icon={'create'}
    on:click={onCreateTranslation}>
    Create
  </Button>
  <slot name="actions" />
</form>
