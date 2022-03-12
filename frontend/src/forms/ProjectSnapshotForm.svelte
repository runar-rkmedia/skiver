<script lang="ts">
  import { api } from 'api'

  import Button from 'components/Button.svelte'
  import { toastApiErr } from 'state'
  import { createEventDispatcher } from 'svelte'

  export let projectID: string

  let state: ApiDef.CreateSnapshotInput = {
    project_id: projectID,
    description: '',
    tag: '',
  }
  const dispatch = createEventDispatcher()

  async function submit() {
    const [result, err] = await api.snapshotMeta.create(state)
    if (err) {
      toastApiErr(err as any)
      return
    }
    state.description = ''
    state.tag = ''
    dispatch('complete', result)
  }
</script>

<form>
  <label>
    Tag
    <input bind:value={state.tag} />
  </label>
  <label>
    Description
    <input bind:value={state.description} />
  </label>
  <Button
    on:click={submit}
    icon="edit"
    type="submit"
    color="primary"
    disabled={!state.tag || !state.project_id}>Create Snapshot</Button>
  <slot name="actions" />
</form>
