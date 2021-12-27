<script lagn="ts">
  import EntityList from 'components/EntityList.svelte'
  import ListItem from 'components/ListItem.svelte'
  import { db } from './api'
  import { state } from './state'
  import formatDate from 'dates'
  import Tip from 'components/Tip.svelte'
</script>

{#if $state.tab === 'locale'}
  <h2>Locales</h2>
  <paper>
    <EntityList
      error={$db.responseStates.locale.error?.error}
      loading={$db.responseStates.locale.loading}>
      {#each Object.values($db.locale)
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
            {v.title}
          </svelte:fragment>
          <svelte:fragment slot="description">
            Created: {formatDate(v.createdAt)}

            {#if v.updatedAt}
              Updated: {formatDate(v.updatedAt)}
            {/if}
          </svelte:fragment>
        </ListItem>
      {/each}
    </EntityList>
  </paper>
{:else if $state.tab === 'project'}
  <paper>
    <h2>Project</h2>
    <Tip key="about-project">
      <p>A project is a typically an application, like a webapp or similar.</p>
      <p>
        By default, each project is seperated from eachother, but they can
        optionally use resources from other projects.
      </p>
    </Tip>
    <EntityList
      error={$db.responseStates.project.error?.error}
      loading={$db.responseStates.project.loading}>
      {#each Object.values($db.project)
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
            {v.title}
          </svelte:fragment>
          <svelte:fragment slot="description">
            Created: {formatDate(v.createdAt)}

            {#if v.updatedAt}
              Updated: {formatDate(v.updatedAt)}
            {/if}
          </svelte:fragment>
        </ListItem>
      {/each}
    </EntityList>
  </paper>
{/if}
