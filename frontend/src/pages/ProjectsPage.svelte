<script lagn="ts">
  import EntityList from '../components/EntityList.svelte'
  import ListItem from '../components/ListItem.svelte'
  import { db } from '../api'
  import { state } from '../state'
  import formatDate from '../dates'
  import Tip from '../components/Tip.svelte'
</script>

<h2>Project</h2>
<paper>
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
          <a href={'#project/' + v.id}>
            {v.title}
          </a>
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
