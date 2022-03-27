<script lang="ts">
  import Button from 'components/Button.svelte'
  import Alert from 'components/Alert.svelte'
  import { createEventDispatcher } from 'svelte'
  import { api, db } from '../api'
  import { toast } from '../state'
  export let projectID: string
  /* If set, will update the category instead of creating one*/
  export let categoryID: string = ''
  const dispatch = createEventDispatcher()
  let payload: ApiDef.CategoryInput & { id?: string } = {
    project_id: projectID,
    title: '',
    description: '',
    key: '',
  }
  $: {
    if (categoryID && !payload.id) {
      reset()
      payload.id = categoryID
    }
  }
  function reset() {
    if (!categoryID) {
      payload = {
        project_id: projectID,
        title: '',
        description: '',
        key: '',
      }
      return
    }
    const c = $db.category[categoryID]
    if (!c) {
      return
    }
    payload = {
      project_id: projectID,
      title: c.title || '',
      description: c.description || '',
      key: c.key || '',
      id: c.id,
    }
  }
  async function onSubmit() {
    payload.project_id = projectID
    if (!payload.project_id) {
      toast({
        title: 'missing argument',
        message: 'project was not set',
        kind: 'warning',
      })
      return
    }
    if (categoryID) {
      const c = $db.category[categoryID]
      let p: ApiDef.UpdateCategoryInput = {
        id: c.id,
        ...(payload.title &&
          payload.title !== c.title && { title: payload.title }),
        ...(payload.key && payload.key !== c.key && { key: payload.key }),
        ...(payload.description &&
          payload.description !== c.description && {
            description: payload.description,
          }),
      }
      const [res, err] = await api.category.update(categoryID, p)
      if (err) {
        return
      }
      dispatch('complete', res.data)
    } else {
      const { id: _, ...p } = payload
      const [res, err] = await api.category.create(p)
      if (err) {
        return
      }
      dispatch('complete', res.data)
    }
    payload = { key: '', project_id: '', title: '' }
  }
</script>

<form id="category-form">
  <label for="category-id">Key</label>
  <input id="category-key" bind:value={payload.key} />
  {#if categoryID && $db.category[categoryID]?.key !== payload.key}
    <Alert kind="warning">
      Changing the <code>key</code>-property may result in older clients not
      being able to find the translations under this category.
      <Button color="secondary" on:click={() => {
        const c = $db.category[categoryID]
        if (!c || !c.key) {
          return
        }
      payload.key = c.key}}>Click to reset</Button>
    </Alert>
  {/if}
  <label for="category-title">Title</label>
  <input id="category-title" bind:value={payload.title} />
  <small>
    <label for="category-description">Description</label>
    <textarea id="category-description" bind:value={payload.description} />
  </small>
  {#if categoryID}
    <Button color="primary" type="submit" icon={'edit'} on:click={onSubmit}>
      Update
    </Button>
  {:else}
    <Button color="primary" type="submit" icon={'create'} on:click={onSubmit}>
      Create
    </Button>
  {/if}
  <slot name="actions" />
</form>

<style>
  form {
    display: block;
  }
</style>
