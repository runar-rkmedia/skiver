<script lang="ts">
  import { db, api } from 'api'
  import { state } from 'state'
  import Button from 'components/Button.svelte'
  import { objectKeys } from 'simplytyped'
  export let projectID: string

  $: project = $db.project[projectID]
  let create = true
  function onCreate() {
    $state.createTranslation.project = projectID
    api.translation.create($state.createTranslation)
  }
</script>

{#if !project}
  {#if $db.responseStates.project.loading}
    Loading...
  {:else if $db.responseStates.project.error}
    {$db.responseStates.project.error.error}
  {:else}
    Project not found: {projectID}
  {/if}
{:else}
  <h2>{project.title}</h2>
  <p>{project.description}</p>

  <hr />

  <paper>
    <table>
      <tr>
        <th />
        <th />
        <th />
      </tr>
      {#each Object.values($db.translation).filter((v) => {
        return true
      }) as v}
        <tr>
          <td>
            {v.prefix}.
            {v.key}
          </td>
          <td>
            {v.title}
            <small>
              {v.description}
            </small>
          </td>
          <td>
            {$db.locale[v.locale_id || '']?.title}
          </td>
          <td>
            {v.value}
          </td>
        </tr>
      {/each}
    </table>
  </paper>

  <Button color="secondary" icon={'edit'} on:click={() => (create = !create)}
    >Create</Button>
  <hr />
  {#if create}
    <paper>
      <form>
        <label>
          Title
          <input name="title" bind:value={$state.createTranslation.title} />
        </label>
        <label>
          Description
          <input
            name="description"
            bind:value={$state.createTranslation.description} />
        </label>
        <label>
          Prefix
          <input name="prefix" bind:value={$state.createTranslation.prefix} />
        </label>
        <label>
          Key
          <input name="key" bind:value={$state.createTranslation.key} />
        </label>
        <label>
          Locale
          <select
            name="locale_id"
            bind:value={$state.createTranslation.locale_id}>
            {#each Object.entries($db.locale) as [k, v]}
              <option value={k}>{v.title} {v.ietf} / {v.iso_639_3}</option>
            {/each}
          </select>
        </label>
        <label>
          Value
          <textarea
            rows={3}
            name="value"
            bind:value={$state.createTranslation.value} />
        </label>
        <Button color="primary" type="submit" icon={'edit'} on:click={onCreate}>
          Create
        </Button>
      </form>
    </paper>
  {/if}
{/if}

<style>
  table small {
    display: block;
  }
</style>
