<script lang="ts">
  import EntityList from '../components/EntityList.svelte'
  import ListItem from '../components/ListItem.svelte'
  import { db, api } from '../api'
  import EntityDetails from 'components/EntityDetails.svelte'
  import OrganizationForm from 'forms/OrganizationForm.svelte'
  import Button from 'components/Button.svelte'
  import { parseDate, formatDate } from 'dates'
  api.organization.list()
  let selected: ApiDef.Organization | null
  const now = new Date()
</script>

{#if selected}
  <Button color="secondary" on:click={() => (selected = null)}>
    Back to organization-list
  </Button>
  <paper>
    <h3>{selected.title}</h3>
  </paper>
{:else}
  <h2>Organizations</h2>
  <paper>
    <EntityList
      error={$db.responseStates.organization.error?.error}
      loading={$db.responseStates.organization.loading}>
      {#each Object.values($db.organization) as v}
        <ListItem ID={v.id} deleted={!!v.deleted}>
          <svelte:fragment slot="header">
            {v.title}
          </svelte:fragment>
          <svelte:fragment slot="description">
            {#if v.join_id && parseDate(v?.join_id_expires)?.getTime() > now.getTime()}
              Join-id: <a href={'#join/' + v.join_id}>{v.join_id}</a>
              {formatDate(v.join_id_expires)}
            {:else if v.join_id}
              Join-id has expired
            {:else}
              No join id
            {/if}
            <EntityDetails entity={v} />
          </svelte:fragment>
          <svelte:fragment slot="actions">
            <Button on:click={() => (selected = v)}>Select</Button>
          </svelte:fragment>
        </ListItem>
      {/each}
    </EntityList>
  </paper>
  <paper>
    <OrganizationForm />
  </paper>
{/if}
