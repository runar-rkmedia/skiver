<script lang="ts">
  import { api } from '../api'
  import Button from 'components/Button.svelte'
  import { onMount } from 'svelte'
  import Alert from 'components/Alert.svelte'
  let selected: ApiDef.Organization | null
  export let joinID: string
  let loading = false
  let org: ApiDef.Organization | null
  let joinedResponse: ApiDef.LoginResponse | null
  let error: ApiDef.APIError | null
  let username = ''
  let password = ''

  const reload = async (id: string) => {
    if (!id) {
      return
    }

    loading = true
    const [res, err] = await api.join.get(joinID)
    if (!err) {
      org = res.data
    }
    error = err
    loading = true
  }
  $: {
    reload(joinID)
  }

  const submit = async () => {
    loading = true
    const [res, err] = await api.join.post(joinID, { username, password })
    loading = false
    if (!err) {
      joinedResponse = res.data
    }
    error = err
  }
</script>

{#if error}
  <Alert kind="error">{error.error?.error}</Alert>
{/if}

{#if joinedResponse}
  You have successfully joined {joinedResponse.organization?.title} as user {username}.
{:else if org}
  <paper>
    <h3>Join organization {org.title}</h3>
    <form>
      <label>
        Username
        <input name="username" bind:value={username} />
      </label>
      <label>
        Password
        <input name="password" bind:value={password} type="password" />
      </label>
      <Button color="primary" on:click={submit}>Join</Button>
    </form>
  </paper>
{:else}
  <h3>Join for id {joinID}</h3>
{/if}
