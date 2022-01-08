<script lang="ts">
  import EntityList from '../components/EntityList.svelte'
  import ListItem from '../components/ListItem.svelte'
  import { db } from '../api'
  import { state } from '../state'
  import formatDate from 'dates'
  import EntityDetails from 'components/EntityDetails.svelte'
</script>

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
          <EntityDetails entity={v} />
        </svelte:fragment>
      </ListItem>
    {/each}
  </EntityList>
</paper>
