<script lang="ts">
  import { db } from 'api'
  import Button from 'components/Button.svelte'
  import Collapse from 'components/Collapse.svelte'
  import EntityDetails from 'components/EntityDetails.svelte'
  import Spinner from 'components/Spinner.svelte'
  import TagDiff from 'components/TagDiff.svelte'
  import Tip from 'components/Tip.svelte'
  import ProjectForm from 'forms/ProjectForm.svelte'
  import ProjectSnapshotForm from 'forms/ProjectSnapshotForm.svelte'
  import sortOn from 'sort-on'
  import { apiUrl } from 'util/appConstants'
  export let projectID: string
  let extendedDiff = false
  $: project = $db.project[projectID]
  $: tags =
    !!project &&
    sortOn(Object.entries(project.snapshots || {}), '-1.created_at')
  let showCreateSnapshotForm = false
</script>

{#if project}
  <h2>Settings for project: {project.title}</h2>
  <paper>
    <ProjectForm {project} />
  </paper>

  <paper>
    <div class="tipheader">
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

        <p>
          When a client requests a set a translations, they can specify a tag to
          refer to the accompanying snapshot.
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
      <h3>Tags</h3>
      {#if !showCreateSnapshotForm}
        <Button
          color="primary"
          icon="create"
          on:click={() => (showCreateSnapshotForm = true)}>New snapshot</Button>
        {#if tags.length}
          <label>
            <input type="checkbox" bind:checked={extendedDiff} />
            Show Extended diff
          </label>
        {/if}
      {:else}
        <ProjectSnapshotForm {projectID} />
      {/if}
    </div>
    <paper>
      <div class="tagHeader">
        <h4>Latest unreleased</h4>
        <a target={projectID} href={apiUrl(`/export/p=${projectID}&f=i18n`)}>
          i18n
        </a>
      </div>
      The latest changes to the project, always available
      {#if tags.length}
        <TagDiff
          tagA={tags[0][0]}
          tagB={''}
          {projectID}
          format={extendedDiff ? 'raw' : 'i18n'} />
      {/if}
    </paper>
    {#each tags as [tag, snapshot], i}
      <paper>
        <div class="tagHeader">
          <h4>{tag}</h4>
          <a
            target={projectID + tag}
            href={apiUrl(`/export/p=${projectID}&f=i18n&t=${tag}`)}>
            i18n
          </a>
        </div>
        {snapshot.description}
        {#if tags[i + 1]}
          <TagDiff
            tagA={tags[i + 1][0]}
            tagB={tag}
            {projectID}
            format={extendedDiff ? 'raw' : 'i18n'} />
        {/if}
        <EntityDetails entity={snapshot} />
      </paper>
    {:else}
      <p>There are currenly no snapshots created for this project</p>
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
  .tagHeader {
    display: flex;
    flex-direction: row;
    justify-content: space-between;
    align-items: center;
  }
</style>
