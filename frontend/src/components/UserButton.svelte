<script lang="ts">
  import { api, db } from 'api'
</script>

{#if $db.login.ok}
  <div class="user-welcome">
    <select
      name="user-options"
      id="user-optinos"
      on:change={(e) => {
        if (!e.currentTarget) {
          console.warn('e.currentTarget was unexpecdedly null', e)
          return null
        }
        const selected = e.currentTarget.value
        switch (selected) {
          case 'logout':
            api.logout()
            window.history.pushState({}, '', '/#/')
            break
          case 'settings':
            window.history.pushState({}, '', '/#user/')
            break
          case 'orgSettings':
            window.history.pushState({}, '', '/#org/')
            break
        }
        e.currentTarget.value = ''
        e.currentTarget.blur()
      }}>
      <option value="" disabled selected>
        Welcome, {$db.login.username}
      </option>
      {#if $db.login.can_create_projects}
        <option value="orgSettings">
          Settings for {$db.login.organization?.title || 'organization'}
        </option>
      {/if}
      <option value="settings">Settings</option>
      <option value="logout">Logout</option>
    </select>
  </div>
{/if}

<style>
  .user-welcome select {
    background: unset;
    background-color: unset;
    color: var(--color-grey-100);
    border: unset;
    display: inline-block;
    margin-inline-end: var(--size-2);
    /* -webkit-appearance: none; */
    /* -moz-appearance: none; */
  }
  .user-welcome select option {
    background-color: var(--color-black);
  }
</style>
