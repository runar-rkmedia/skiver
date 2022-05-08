<script lang="ts">
  import { api, db } from '../api'
  import Alert from 'components/Alert.svelte'
  import isPast from 'date-fns/isPast'
  import { parseDate, formatDate } from 'dates'
  import Icon from 'components/Icon.svelte'
  import Button from 'components/Button.svelte'
  import { toast } from 'state'
  import { onMount } from 'svelte'
  import { addMonths } from 'date-fns'
  import Spinner from 'components/Spinner.svelte'
  export let organizationID: string
  onMount(() => {
    if (organizationID) {
      api.organization.list()
    }
    api.user.list()
  })
  $: organization = $db.organization[organizationID]
  $: users = Object.values($db.user)

  function renewJoinID() {
    if (!organization.id) {
      return
    }
    if (!organization.join_id) {
      return
    }
    api.organization.update(organization.id, {
      id: organization.id,
      join_id: organization.join_id,
      join_id_expires: addMonths(new Date(), 1).toISOString(),
    })
  }
  function invalidateJoinID() {
    if (!organization.id) {
      return
    }
    if (!organization.join_id) {
      return
    }
    api.organization.update(organization.id, {
      id: organization.id,
      join_id: organization.join_id,
      join_id_expires: new Date(),
    })
  }
  function createJoinID() {
    if (!organization.id) {
      return
    }
    api.organization.update(organization.id, {
      id: organization.id,
      join_id_expires: addMonths(new Date(), 1).toISOString(),
    })
  }
</script>

<Spinner active={$db.responseStates.organization.loading} />
{#if !organization && !$db.responseStates.organization.loading}
  <Alert kind="error">Organization Not found</Alert>
{:else}
  <paper>
    <h2>{organization.title}</h2>

    <h3>Join ID</h3>

    <div class="flex-column">
      {#if organization.join_id}
        <a href={'#join/' + organization.join_id}>{organization.join_id}</a>
        {#if organization.join_id_expires}
          {#if isPast(parseDate(organization.join_id_expires))}
            <Icon icon="warning" />
            <span class="color-error">
              Expired {formatDate(organization.join_id_expires)}
            </span>
            <Button icon="refresh" color="primary" on:click={renewJoinID}
              >Renew</Button>
          {:else}
            Valid until: {formatDate(organization.join_id_expires)}
            <Button icon="delete" color="danger" on:click={invalidateJoinID}
              >Invalidate</Button>
          {/if}
        {/if}
        <Button icon="create" color="primary" on:click={createJoinID}
          >Recreate</Button>
      {:else}
        <Button icon="create" color="primary" on:click={createJoinID}
          >Create</Button>
      {/if}
    </div>

    <h3>Users</h3>

    {#if $db.login.can_update_users}
      <div>
        <table>
          <thead>
            <th>UserName</th>
            <th>Create Organization</th>
            <th>Create User</th>
            <th>Create Locale</th>
            <th>Create Project</th>
            <th>Create Translation</th>

            <th>Update Organization</th>
            <th>Update User</th>
            <th>Update Locale</th>
            <th>Update Project</th>
            <th>Update Translation</th>
          </thead>
          <tbody>
            {#each users as u}
              <tr>
                <td>{u.username}</td>
                <td class="boolean">
                  <input
                    disabled
                    type="checkbox"
                    checked={u.can_create_organization} />
                </td>
                <td class="boolean">
                  <input type="checkbox" checked={u.can_create_users} />
                </td>
                <td class="boolean">
                  <input type="checkbox" checked={u.can_create_locales} />
                </td>
                <td class="boolean">
                  <input type="checkbox" checked={u.can_create_projects} />
                </td>
                <td class="boolean">
                  <input type="checkbox" checked={u.can_create_translations} />
                </td>

                <td class="boolean">
                  <input type="checkbox" checked={u.can_update_organization} />
                </td>
                <td class="boolean">
                  <input type="checkbox" checked={u.can_update_users} />
                </td>
                <td class="boolean">
                  <input type="checkbox" checked={u.can_update_locales} />
                </td>
                <td class="boolean">
                  <input type="checkbox" checked={u.can_update_projects} />
                </td>
                <td class="boolean">
                  <input type="checkbox" checked={u.can_update_translations} />
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {:else}
      <ul>
        {#each users as u}
          <li>{u}</li>
        {/each}
      </ul>
    {/if}
  </paper>
{/if}

<style>
</style>
