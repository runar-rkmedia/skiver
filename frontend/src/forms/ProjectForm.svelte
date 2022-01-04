<script lang="ts">
  import Button from 'components/Button.svelte'
  import { api } from '../api'
  import { state } from '../state'
  export let shortNameReadOnly = false
  function createProject() {
    api.project.create($state.createProject)
  }
</script>

<form>
  <label>
    Short-name
    <input
      readonly={shortNameReadOnly}
      bind:value={$state.createProject.short_name} />
  </label>
  <label>
    Title
    <input bind:value={$state.createProject.title} />
  </label>
  <label>
    Description
    <input bind:value={$state.createProject.description} />
  </label>
  <Button on:click={createProject} icon="edit" type="submit" color="primary"
    >Create Project</Button>
  <slot name="actions" />
</form>
