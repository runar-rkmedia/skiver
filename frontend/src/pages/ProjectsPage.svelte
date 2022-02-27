<script lagn="ts">
  import EntityList from '../components/EntityList.svelte'
  import ListItem from '../components/ListItem.svelte'
  import { db } from '../api'
  import { state } from '../state'
  import Tip from '../components/Tip.svelte'
  import ProjectForm from 'forms/ProjectForm.svelte'
  import EntityDetails from 'components/EntityDetails.svelte'
  import Button from 'components/Button.svelte'
  import Icon from 'components/Icon.svelte'
  $: projects = Object.values($db.project)
  let showCreate = false
</script>

<h2>Project</h2>
<Tip key="about-project">
  <p>A project is a typically an application, like a webapp or similar.</p>
  <p>
    By default, each project is seperated from eachother, but they can
    optionally use resources from other projects.
  </p>
</Tip>
{#if showCreate}
  <paper>
    <ProjectForm />
  </paper>
{:else}
  <Button color="primary" icon="create" on:click={() => (showCreate = true)}
    >Create project</Button>
{/if}
<div class="spacer" />
<paper>
  <EntityList
    error={$db.responseStates.project.error?.error}
    loading={$db.responseStates.project.loading}>
    {#each projects
      .filter((e) => {
        if (!$state.showDeleted) {
          return !e.deleted
        }
        return true
      })
      .sort((a, b) => {
        const A = a.createdAt
        const B = b.createdAt
        if (A > B) {
          return 1
        }
        if (A < B) {
          return -1
        }

        return 0
      })
      .reverse() as v}
      <ListItem
        deleteDisabled={true}
        editDisabled={true}
        ID={v.id}
        deleted={!!v.deleted}>
        <svelte:fragment slot="header">
          <div class="projectHeader">
            <a href={'#project/' + v.id}>
              {v.title}
            </a>
          </div>
        </svelte:fragment>
        <svelte:fragment slot="description">
          {v.description}
          <EntityDetails entity={v} />
        </svelte:fragment>
        <svelte:fragment slot="actions">
          <a href={'#project/' + v.id + '/settings'}>
            <Icon icon="settings" />
            Settings
          </a>
        </svelte:fragment>
      </ListItem>
    {:else}
      {#if $db.responseStates.project.loading}
        Gathering...
      {:else}
        No projects created yet
      {/if}
    {/each}
  </EntityList>
</paper>

<style>
  .spacer {
    height: var(--size-6);
  }
  .projectHeader {
    display: flex;
    width: 100%;
    justify-content: space-between;
  }
  a {
    display: flex;
    gap: 0.8ch;
  }
</style>
