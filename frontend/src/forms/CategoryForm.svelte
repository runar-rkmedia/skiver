<script lang="ts">
  import Button from 'components/Button.svelte'
  import { createEventDispatcher } from 'svelte'
  import { api } from '../api'
  import { state, toast } from '../state'
  export let projectID: string
  const dispatch = createEventDispatcher()
  async function onCreateCategory() {
    $state.createCategory.project_id = projectID
    if (!$state.createCategory.project_id) {
      toast({
        title: 'missing argument',
        message: 'project was not set',
        kind: 'warning',
      })
      return
    }
    const s = await api.category.create($state.createCategory)
    if (!s[1]) {
      dispatch('complete', s[0].data)
      $state.createCategory = { key: '', project_id: '', title: '' }
    }
  }
</script>

<form id="category-form">
  <label for="category-id">Key</label>
  <input id="category-key" bind:value={$state.createCategory.key} />
  <label for="category-title">Title</label>
  <input id="category-title" bind:value={$state.createCategory.title} />
  <small>
    <label for="category-description">Description</label>
    <textarea
      id="category-description"
      bind:value={$state.createCategory.description} />
  </small>
  <Button
    color="primary"
    type="submit"
    icon={'create'}
    on:click={onCreateCategory}>
    Create
  </Button>
  <slot name="actions" />
</form>

<style>
  form {
    display: block;
  }
</style>
