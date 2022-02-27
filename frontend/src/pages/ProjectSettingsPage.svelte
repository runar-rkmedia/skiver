<script lang="ts">
  import { db } from 'api'
  import EntityDetails from 'components/EntityDetails.svelte'
  import Spinner from 'components/Spinner.svelte'
  import Tip from 'components/Tip.svelte'
  import ProjectForm from 'forms/ProjectForm.svelte'
  import sortOn from 'sort-on'
  import { apiUrl } from 'util/appConstants'
  export let projectID: string
  $: project = $db.project[projectID]
</script>

{#if project}
  <h2>Settings for project: {project.title}</h2>
  <paper>
    <ProjectForm {project} />
  </paper>

  <paper>
    <div class="tipheader">
      <h3>Tags</h3>
      <Tip key="project-tags">
        <p>
          Projects can be snapshotted at a given point in time and given tags to
          refer to them.
        </p>

        <p>
          This is useful if you want to control when applications receive
          updates to their translations, and still be able to work with
          translations across time.
        </p>

        <p>There are a few special tags.</p>
        <ul>
          <li>
            Tags that are <a href="https://semver.org/">Semver</a>-compatible,
            like
            <code> v2.1.3 </code>
          </li>
          <li><code>latest</code>Refers to the latest snapshot created</li>
          <li>No tag. Will refer to the live project</li>
        </ul>
      </Tip>
    </div>
    {#each sortOn(Object.entries(project.snapshots || {}), '-1.created_at') as [tag, snapshot]}
      <paper>
        <h4>{tag}</h4>
        <a href={apiUrl(`/export/p=${projectID}&f=i18n&t=${tag}&l=nb`)}>
          i18n
        </a>
        {snapshot.description}
        <EntityDetails entity={snapshot} />
      </paper>
    {/each}
  </paper>
{:else if $db.responseStates.project.loading}
  Gathering information... hold on... <Spinner />
{/if}

<style>
  li {
    display: list-item;
  }
  li:not(:last-of-type) {
    margin-block-end: var(--size-3);
  }
  li code {
    margin: 0;
  }
</style>
