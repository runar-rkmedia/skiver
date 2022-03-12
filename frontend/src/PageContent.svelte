<script lagn="ts">
  import ProjectPage from 'pages/ProjectPage.svelte'
  import router from 'util/router'
  import { db } from './api'
  import LocalePage from './pages/LocalePage.svelte'
  import ProjectsPage from './pages/ProjectsPage.svelte'
  import ProjectSettingsPage from './pages/ProjectSettingsPage.svelte'
  import MissingTranslationsPage from './pages/MissingTranslationsPage.svelte'
  import OrganizationPage from 'pages/OrganizationPage.svelte'
  import JoinPage from 'pages/JoinPage.svelte'
  import AboutPage from 'pages/AboutPage.svelte'
  import Embed from 'pages/Embed.svelte'
  let mainRoute = ''
  let routeArgs = ''
  $: {
    routeArgs = $router.hash.replace('#', '').split('/')
    mainRoute = routeArgs[0]
    routeArgs = routeArgs.slice(1)
  }
</script>

<!-- Simplified routing -->
{#if mainRoute === 'locale'}
  <LocalePage />
{:else if mainRoute === 'embed'}
  <Embed
    projectKey={routeArgs[0]}
    categoryKey={routeArgs[1]}
    translationKeyLike={routeArgs[2]}>
    <div slot="after" let:project>
      <p>
        You are viewing the embedded version of this page.
        {#if project}
          <a href={'#project/' + project.id}
            >Click her to go to the project-view</a>
        {:else}
          <a href={'#/'}>Click her to go back to the main page</a>
        {/if}
      </p>
    </div>
  </Embed>
{:else if mainRoute === 'missing'}
  <MissingTranslationsPage />
{:else if mainRoute === 'project'}
  {#if routeArgs[0]}
    {#if routeArgs[1] === 'settings'}
      <ProjectSettingsPage projectID={routeArgs[0]} />
    {:else}
      <ProjectPage projectID={routeArgs[0]} />
    {/if}
  {:else}
    <ProjectsPage />
  {/if}
{:else if mainRoute === 'join'}
  <JoinPage joinID={routeArgs[0]} />
{:else if mainRoute === 'about'}
  <AboutPage />
{:else if mainRoute === ''}
  <h2>Welcome to Skiver!</h2>
  <p>
    Skiver is a management-system for translations. It aims to be simple, and
    convenient to use.
  </p>
  {#if $db.login.ok && $db.login?.can_create_organization}
    <OrganizationPage />
  {/if}
{:else}
  Not found
{/if}
