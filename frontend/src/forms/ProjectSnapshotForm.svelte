<script lang="ts">
  import { api } from 'api'

  import Button from 'components/Button.svelte'
  import Icon from 'components/Icon.svelte'
  import { toastApiErr } from 'state'
  import { createEventDispatcher } from 'svelte'
  import isSemver from 'util/semver'

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
  {#if state.tag}
    {#if isSemver(state.tag)}
      <p><Icon icon="success" />The tag is semver-compatible</p>
    {:else}
      <p>
        <Icon icon="warning" />The tag is not semver-compatible, like
        <code>1.2.3</code>. You are not required to use semver-compatible tags,
        but it is recommended.
      </p>
    {/if}
  {/if}
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
