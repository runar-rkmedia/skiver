<script lang="ts">
  import { api, db } from 'api'
  import Collapse from '../components/Collapse.svelte'
  import Button from '../components/Button.svelte'
  let currentPasswod = ''
  let newPassword = ''
  let confirmPassword = ''
  let tokens = ['abc', 'foo', 'bar']
  function days(n: number) {
    return 24 * n
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
            type="password"
            name="currentpassword"
            bind:value={currentPasswod} />
        </label>
        <label>
          New password
          <input type="password" name="password" bind:value={newPassword} />
        </label>
        <label>
          Confirm password
          <input
            type="password"
            name="confirmpassword"
            bind:value={confirmPassword} />
        </label>
        <Button color="primary">Submit</Button>
      </form>
    </Collapse>
  </paper>
  <paper>
    <Collapse>
      <h3 slot="title">Api-tokens</h3>
      <p>API-tokens can be used to programmatically use Skiver.</p>

      <div>
        <select name="token-lifetime" id="token-lifetime">
          <option value={days(30)}>30 days</option>
          <option value={days(90)}>90 days</option>
          <option value={days(133)}>Half a year</option>
          <option value={days(365)}>One year</option>
          <option value={days(2 * 365)}>Two years</option>
        </select>
        <Button color="primary">Create new token</Button>
      </div>
    </Collapse>
  </paper>
{:else}
  <h2>Not logged in</h2>
{/if}
