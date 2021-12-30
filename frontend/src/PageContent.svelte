<script lagn="ts">
  import Alert from 'components/Alert.svelte'
  import ProjectPage from 'pages/ProjectPage.svelte'
  import router from 'util/router'
  import { db } from './api'
  import LocalePage from './pages/LocalePage.svelte'
  import ProjectsPage from './pages/ProjectsPage.svelte'
  let mainRoute = ''
  let routeArgs = ''
  $: {
    routeArgs = $router.hash.replace('#', '').split('/')
    mainRoute = routeArgs[0]
    routeArgs = routeArgs.slice(1)
  }
  $: errs = Object.entries($db.responseStates).filter(([_, v]) => {
    return v && v.error
  })
</script>

<!-- Display any errors... -->
{#each Object.entries(errs) as [_, [k, v]]}
  <Alert kind="error">
    <h4 slot="title">{k}</h4>
    <h5>{v.error.code}</h5>
    <p>{v.error.error}</p>
    {#if v.error.details}
      {JSON.stringify(v.error.details)}
    {/if}
  </Alert>
{/each}

<!-- Simplified routing -->
{#if mainRoute === 'locale'}
  <LocalePage />
{:else if mainRoute === 'project'}
  {#if routeArgs[0]}
    <ProjectPage projectID={routeArgs[0]} />
  {:else}
    <ProjectsPage />
  {/if}
{:else if mainRoute === ''}
  <h2>Welcome to Skiver!</h2>
  <p>
    Skiver is a management-system for translations. It aims to be simple, and
    convinient to use.
  </p>
{:else}
  Not found
{/if}
