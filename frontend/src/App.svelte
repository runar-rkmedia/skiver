<svelte:options immutable={true} />

<script lang="ts">
  import "tippy.js/dist/tippy.css"; // Tooltips popover
  import { api, db } from "./api";
  import { scale } from "svelte/transition";

  import ServerInfo from "./components/ServerInfo.svelte";
  import { onMount } from "svelte";
  let username = $db.login.userName;
  let password: "";
  onMount(() => {
    const loadingEl = document.getElementById("loading");
    if (loadingEl) {
      loadingEl.remove();
    }
    api.login.get().then((r) => {
      console.log("login", r);
    });
  });

  import Tabs from "./components/Tabs.svelte";
  import Alert from "./components/Alert.svelte";
  import { state } from "./state";
  import Button from "./components/Button.svelte";
  import Spinner from "components/Spinner.svelte";
  import EntityList from "components/EntityList.svelte";
  import ListItem from "components/ListItem.svelte";
  import formatDate from "dates";
  let didRunInital = false;
  $: {
    if (!didRunInital && $db.login.ok) {
      api.serverInfo();
      api.locale.list();
      api.project.list();
      didRunInital = true;
    }
  }

  const dbWarnSizeGB = 0.5;
  const dbWarnSize = dbWarnSizeGB * 1e9;
</script>

<div class="wrapper">
  <header>
    <img src="/logo.svg" alt="Logo" />
    <h1>Skiver - Ski's the limit</h1>
    {#if $db.login.ok}
      Welcome, {$db.login.userName}
    {/if}
    <Tabs bind:value={$state.tab} />
  </header>
  <div />

  {#if !$db.login.ok}
    <div class="login" transition:scale>
      <paper>
        <h2>Login</h2>
        {#if $db.responseStates.login.loading}
          <Spinner />
        {/if}
        {#if $db.responseStates.login?.error}
          <Alert kind="error">{$db.responseStates.login.error.error}</Alert>
        {/if}
        <form
          on:submit|preventDefault={() =>
            api.login.post({ username, password })}
        >
          <label>
            Username
            <input name="text" bind:value={username} />
          </label>

          <label>
            Password
            <input name="password" type="password" bind:value={password} />
          </label>
          <Button
            preventDefault={false}
            color="primary"
            icon="signIn"
            type="submit">Login</Button
          >
        </form>
      </paper>
    </div>
  {/if}

  <main>
    {#if ($db.serverInfo?.DatabaseSize || 0) > dbWarnSize}
      <Alert kind="warning">
        <div slot="title">The servers database has grown a bit big.</div>

        <p>It is currently {$db.serverInfo.DatabaseSizeStr}</p>
        <p>This may affect performance.</p>
        <p>Some functionality may have been disabled.</p>
        <p>It is adviced to clean the database</p>
      </Alert>
    {/if}
    {#if $state.tab === "locale"}
      <h2>Locales</h2>
      <paper>
        <EntityList
          error={$db.responseStates.locale.error?.error}
          loading={$db.responseStates.locale.loading}
        >
          {#each Object.values($db.locale)
            .filter((e) => {
              if (!$state.showDeleted) {
                return !e.deleted;
              }
              return true;
            })
            .sort((a, b) => {
              const A = a.createdAt;
              const B = b.createdAt;
              if (A > B) {
                return 1;
              }
              if (A < B) {
                return -1;
              }

              return 0;
            })
            .reverse() as v}
            <ListItem
              deleteDisabled={true}
              editDisabled={true}
              ID={v.id}
              deleted={!!v.deleted}
            >
              <svelte:fragment slot="header">
                {v.title}
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
    {:else if $state.tab === "project"}
      <EntityList
        error={$db.responseStates.project.error?.error}
        loading={$db.responseStates.project.loading}
      >
        {#each Object.values($db.project)
          .filter((e) => {
            if (!$state.showDeleted) {
              return !e.deleted;
            }
            return true;
          })
          .sort((a, b) => {
            const A = a.createdAt;
            const B = b.createdAt;
            if (A > B) {
              return 1;
            }
            if (A < B) {
              return -1;
            }

            return 0;
          })
          .reverse() as v}
          <ListItem
            deleteDisabled={true}
            editDisabled={true}
            ID={v.id}
            deleted={!!v.deleted}
          >
            <svelte:fragment slot="header">
              {v.title}
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
    {/if}
  </main>
  {#if $state.serverStats}
    <iframe
      title="Server Statistics"
      id="statsviz"
      height="600"
      width="100%"
      src="https://localhost/debug/statsviz/"
    />

    <a href="https://localhost/debug/statsviz/">Statwiz statistics</a>
  {/if}
  <footer>
    <ServerInfo />
  </footer>
</div>

<style>
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
  main {
    margin-block-end: var(--size-12);
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
  header h1 {
    margin-inline: var(--size-4);
    align-self: center;
  }
  main {
    margin-inline: var(--size-4);
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
