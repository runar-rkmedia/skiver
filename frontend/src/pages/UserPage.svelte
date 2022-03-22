<script lang="ts">
  import { api, db } from 'api'
  import Collapse from '../components/Collapse.svelte'
  import Button from '../components/Button.svelte'
  import Alert from '../components/Alert.svelte'
  import { toast } from 'state'
  import formatDate from 'dates'
  let currentPassword = ''
  let newPassword = ''
  let confirmPassword = ''
  let pwError = ''
  let tokenError = ''
  let tokenDuration = 24 * 30
  let token: null | ApiDef.TokenResponse = null
  let tokenDescription = ''
  function days(n: number) {
    return 24 * n
  }
  function handleChangePassword() {
    if (currentPassword === '') {
      pwError = 'Missing current password'
    }
    if (confirmPassword === '') {
      pwError = 'Missing confirm password'
    }
    if (newPassword === '') {
      pwError = 'Missing new password'
    }
    if (confirmPassword !== newPassword) {
      pwError = 'The passwords do not match'
    }
    api
      .changePassword({
        password: currentPassword,
        new_password: newPassword,
      })
      .then(([_, err]) => {
        if (err) {
          pwError = err.error?.error || ''
          return
        }
        currentPassword = ''
        confirmPassword = ''
        newPassword = ''
        toast({
          message: 'Password changed successfully',
          kind: 'info',
          title: 'Success',
        })
      })
  }
  function onPwFocus() {
    pwError = ''
  }
  function handleCreateToken() {
    tokenError = ''
    api
      .generateApiToken({
        ttl_hours: tokenDuration,
        description: tokenDescription,
      })
      .then(([res, err]) => {
        if (err) {
          tokenError = err.error?.error || ''
          return
        }
        token = res.data
      })
  }
  async function copyTokenToClipboard() {
    if (!token?.token) {
      return
    }
    try {
      await navigator.clipboard.writeText(token.token)
    } catch (err) {
      toast({
        kind: 'error',
        title: 'Failed to copy to clipboard',
        message: '',
      })
      return
    }
    toast({ kind: 'info', title: 'Token copied to clipboard', message: '' })
  }
</script>

{#if $db.login.ok}
  <h2>Settings</h2>

  <paper>
    <Collapse>
      <h3 slot="title">Change password</h3>
      <form>
        <label>
          Current password
          <input
            on:focus={onPwFocus}
            type="password"
            name="currentpassword"
            bind:value={currentPassword} />
        </label>
        <label>
          New password
          <input
            on:focus={onPwFocus}
            type="password"
            name="password"
            bind:value={newPassword} />
        </label>
        <label>
          Confirm password
          <input
            on:focus={onPwFocus}
            type="password"
            name="confirmpassword"
            bind:value={confirmPassword} />
        </label>
        {#if pwError}
          <div class="error-msg">
            <Alert kind="error">{pwError}</Alert>
          </div>
        {/if}
        <Button color="primary" on:click={handleChangePassword}>Submit</Button>
      </form>
    </Collapse>
  </paper>
  <paper>
    <Collapse>
      <h3 slot="title">Api-tokens</h3>
      <p>API-tokens can be used to programmatically use Skiver.</p>

      <div>
        {#if tokenError}
          <div class="error-msg">
            <Alert kind="error">{tokenError}</Alert>
          </div>
        {/if}
        {#if token}
          <div class="full-width-input">
            <Alert kind="success">
              <h4 slot="title">Your token was generated.</h4>
              Copy it for safekeeping. Once you navigate away from this page, you
              will not be be able to get it back.</Alert>
            <label>
              Token:
              <input type="text" readonly={true} value={token.token || ''} />
            </label>
            <Button color="primary" on:click={copyTokenToClipboard}
              >Copy to clipboard</Button>
            <div>
              {token.description}
              <small>
                Expires: {formatDate(token.expires)}
              </small>
            </div>
          </div>
        {:else}
          <select
            name="token-lifetime"
            id="token-lifetime"
            bind:value={tokenDuration}>
            <option value={days(30)}>30 days</option>
            <option value={days(90)}>90 days</option>
            <option value={days(133)}>Half a year</option>
            <option value={days(365)}>One year</option>
            <option value={days(2 * 365)}>Two years</option>
          </select>
          <div class="full-width-input">
            <label>
              Description
              <input
                type="text"
                name="token-description"
                bind:value={tokenDescription} />
            </label>
            <Button color="primary" on:click={handleCreateToken}
              >Create new token</Button>
          </div>
        {/if}
      </div>
    </Collapse>
  </paper>
{:else}
  <h2>Not logged in</h2>
{/if}

<style>
  .error-msg {
    padding-block-end: var(--size-2);
  }
  .full-width-input input {
    width: 100%;
  }
</style>
