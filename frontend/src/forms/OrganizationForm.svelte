<script lang="ts">
  import Button from 'components/Button.svelte'
  import { api } from '../api'
  import { state } from '../state'
  function createOrganization() {
    api.organization.create($state.createOrganization)
  }
</script>

<form>
  <label>
    Title
    <input bind:value={$state.createOrganization.title} />
  </label>
  <Button
    on:click={createOrganization}
    icon="edit"
    type="submit"
    color="primary">Create organization</Button>
  <slot name="actions" />
</form>
