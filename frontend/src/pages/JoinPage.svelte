<script lang="ts">
  import { api } from '../api'
  import Button from 'components/Button.svelte'
  import Alert from 'components/Alert.svelte'
  import { appUrl } from 'util/appConstants'
  export let joinID: string
  let org: ApiDef.Organization | null
  let joinedResponse: ApiDef.LoginResponse | null
  let error: ApiDef.APIError | null
  let username = ''
  let password = ''
  let password2 = ''

  const reload = async (id: string) => {
    if (!id) {
      return
    }

    const [res, err] = await api.join.get(joinID)
    if (!err) {
      org = res.data
    }
    error = err
  }
  $: {
    reload(joinID)
  }

  const submit = async () => {
    const [res, err] = await api.join.post(joinID, { username, password })
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
  <p>
    Please login to get started by <a href={appUrl('')}
      >going to the main-page</a>
  </p>
{:else if org}
  <paper>
    <h3>Join organization {org.title}</h3>
    <form>
      <label>
        Username
        <input name="username" bind:value={username} autocomplete="username" />
      </label>
      <label>
        Password
        <input
          name="password"
          bind:value={password}
          type="password"
          autocomplete="new-password" />
      </label>
      <label>
        Confirm Password
        <input
          name="password-2"
          bind:value={password2}
          type="password"
          autocomplete="new-password" />
      </label>
      <Button
        color="primary"
        disabled={!username || !password || password !== password2}
        on:click={submit}>Join</Button>
    </form>
  </paper>
{:else}
  <h3>Join for id {joinID}</h3>
{/if}
