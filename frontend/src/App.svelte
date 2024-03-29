<svelte:options immutable={true} />

<script lang="ts">
  import 'tippy.js/dist/tippy.css' // Tooltips popover
  import { api, db } from './api'
  import { scale, fly } from 'svelte/transition'
  import Alert from './components/Alert.svelte'
  import Tabs from './components/Tabs.svelte'
  import Button from './components/Button.svelte'
  import Spinner from './components/Spinner.svelte'
  import PageContent from 'PageContent.svelte'
  import { clearToast, state, toast, toastApiErr } from 'state'
  import { appUrl } from 'util/appConstants'
  import router from 'util/router'

  import ServerInfo from './components/ServerInfo.svelte'
  import { onMount } from 'svelte'
  import UserButton from 'components/UserButton.svelte'
  let username = $db.login.username
  let requiresLogin = false
  let password = ''
  let newPassword = ''
  let confirmPassword = ''
  onMount(() => {
    const loadingEl = document.getElementById('loading')
    if (loadingEl) {
      loadingEl.remove()
    }
    api.login.get()
    window.addEventListener('click', () => {
      const toast = Object.keys($state.toasts).pop()
      if (!toast) {
        return
      }
      const diff =
        new Date().getTime() - new Date($state.toasts[toast].created).getTime()
      if (diff < 1000) {
        return
      }
      clearToast(toast)
    })
  })

  let showHeader = true
  let showFooter = true
  $: {
    let routeArgs = $router.hash.replace('#', '').split('/')
    let mainRoute = routeArgs[0]
    if (mainRoute === 'embed') {
      showHeader = false
      showFooter = false
    } else {
      showHeader = true
      showFooter = true
    }
    switch (mainRoute) {
      case 'about':
      case 'join':
        requiresLogin = false
        break
      default:
        requiresLogin = true
    }
  }

  const dbWarnSizeGB = 0.5
  const dbWarnSize = dbWarnSizeGB * 1e9
</script>

<div class="toasts">
  {#each Object.values($state.toasts) as toast}
    <div class="toast" transition:fly|local>
      <Alert kind={toast.kind}>
        <svelte:fragment slot="title">{toast.title}</svelte:fragment>
        {toast.message}
      </Alert>
    </div>
  {/each}
</div>

<div class="wrapper">
  {#if showHeader}
    <header>
      <div>
        <a href="#/">
          <img src={appUrl('/logo.svg')} alt="Logo" />
        </a>
        <a href="#/">
          <h1>Skiver - Ski's the limit</h1>
        </a>
      </div>
      <Tabs />
      <div class="user-welcome">
        <UserButton />
      </div>
    </header>
  {/if}
  <div />

  {#if $db.login.ok === null && $db}
    <div style="float: right">
      <Alert kind="info">Checking login...</Alert>
    </div>
  {:else if (!$db.login.ok && requiresLogin) || $db.login.temporary_password}
    <div class="login" transition:scale|local>
      <paper>
        {#if $db.login.temporary_password}
          <h2>You must change your password</h2>
          <form
            on:submit|preventDefault={async () => {
              if (
                !newPassword ||
                !password ||
                newPassword !== confirmPassword
              ) {
                return
              }

              const [_res, err] = await api.changePassword({
                password,
                new_password: newPassword,
              })
              if (err != null) {
                toastApiErr(err)
                return
              }
              toast({
                kind: 'info',
                title: 'Success',
                message: 'Your password was successfully changed',
              })
              password = ''
              confirmPassword = ''
              newPassword = ''
              $db.login.temporary_password = false
            }}>
            <label>
              Username
              <input
                name="text"
                bind:value={username}
                readonly
                autocapitalize="none" />
            </label>

            <label>
              Current Password
              <!-- svelte-ignore a11y-autofocus -->
              <input
                autocomplete="current-password"
                name="current-password"
                type="password"
                placeholder="Password"
                required
                bind:value={password} />
            </label>
            <label>
              New Password
              <input
                autocomplete="new-password"
                type="password"
                placeholder="New Password"
                required
                bind:value={newPassword} />
            </label>
            <label>
              Confirm Password
              <input
                autocomplete="new-password"
                type="password"
                name="confirm_password"
                placeholder="Confirm Password"
                required
                bind:value={confirmPassword} />
            </label>
            <Button
              preventDefault={false}
              color="primary"
              disabled={!newPassword || newPassword !== confirmPassword}
              icon="signIn"
              type="submit">
              Change password
            </Button>
          </form>
        {:else}
          <h2>Login</h2>
          {#if $db.responseStates.login.loading}
            <Spinner />
          {/if}
          {#if $db.responseStates.login?.error?.error}
            <Alert kind="error"
              >{$db.responseStates.login.error.error.error}</Alert>
          {/if}
          <form
            on:submit|preventDefault={() => {
              if (!username || !password) {
                return
              }

              api.login.post({ username, password })
              password = ''
            }}>
            <label>
              Username
              <!-- svelte-ignore a11y-autofocus -->
              <input
                name="text"
                autocomplete="username"
                bind:value={username}
                autofocus={true}
                autocapitalize="none" />
            </label>

            <label>
              Password
              <input
                autocomplete="current-password"
                name="password"
                type="password"
                bind:value={password} />
            </label>
            <Button
              preventDefault={false}
              color="primary"
              disabled={!username || !password}
              icon="signIn"
              type="submit">
              Login
            </Button>
          </form>
        {/if}
      </paper>
    </div>
  {/if}

  {#if $db.login.ok || !requiresLogin}
    <main>
      {#if ($db.serverInfo?.database_size || 0) > dbWarnSize}
        <Alert kind="warning">
          <div slot="title">The servers database has grown a bit big.</div>

          <p>It is currently {$db.serverInfo.database_size_str}</p>
          <p>This may affect performance.</p>
          <p>Some functionality may have been disabled.</p>
          <p>It is adviced to clean the database</p>
        </Alert>
      {/if}
      <PageContent />
    </main>
  {/if}
  {#if showFooter}
    <footer>
      <ServerInfo />
      <a href={appUrl('/docs')} target="skiver-swagger">Docs</a>
      <a href="#about/" target="skiver-swagger">About</a>
    </footer>
  {/if}
</div>

<style>
  .toasts {
    position: fixed;
    top: var(--size-4);
    right: var(--size-4);
    z-index: 100;
  }
  .toast {
    box-shadow: var(--elevation-4);
  }
  header a {
    color: unset;
    text-decoration: unset;
  }
  footer a {
    color: unset;
  }
  .user-welcome {
    margin-left: auto;
    margin-block: auto;
    padding: var(--size-4);
    display: flex;
    align-items: center;
  }
  .login {
    position: fixed;
    top: 0;
    right: 0;
    bottom: 0;
    left: 0;
    width: 100%;
    display: flex;
    justify-content: center;
    align-items: center;

    z-index: 10000;
    background: #331415aa;
  }
  .login > * {
    margin-inline: auto;
    width: fit-content;
    height: fit-content;
    padding: 40px;
  }
  main {
    margin-block-end: var(--size-12);
    width: 100%;
    z-index: 1;
  }
  .wrapper {
    background-color: var(--color-blue-300);
    display: flex;
    flex-direction: column;
    min-height: 100%;
  }
  header {
    background-color: var(--color-black);
    color: hsl(240, 80%, 95%);
    display: flex;
    box-shadow: var(--elevation-4);
  }
  header div h1 {
    margin-inline: var(--size-4);
    align-self: center;
  }
  header div {
    display: flex;
  }

  main {
    margin-block-start: var(--size-2);
    padding-inline-start: var(--gutter-start);
    padding-inline-end: var(--gutter-end);
  }
  img {
    height: 100px;
    width: 100px;
    max-width: 20vw;
  }
  footer {
    margin-top: auto;
    display: flex;
    width: 100%;
    justify-content: space-between;
    padding: var(--size-4);
    background-color: var(--color-black);
    color: var(--color-grey-100);
  }
  form {
    max-width: 500px;
    margin-inline: auto;
  }
</style>
