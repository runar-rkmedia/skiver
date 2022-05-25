<script lang="ts">
  import { db, localeKeyToKeyOfLocaleMap } from 'api'
  import Button from 'components/Button.svelte'
  import Collapse from 'components/Collapse.svelte'
  import Dialogs from 'components/Dialogs.svelte'
  import EntityDetails from 'components/EntityDetails.svelte'
  import LocaleFlag from 'components/LocaleFlag.svelte'
  import Spinner from 'components/Spinner.svelte'
  import TagDiff from 'components/TagDiff.svelte'
  import Tip from 'components/Tip.svelte'
  import ProjectForm from 'forms/ProjectForm.svelte'
  import ProjectSnapshotForm from 'forms/ProjectSnapshotForm.svelte'
  import sortOn from 'sort-on'
  import { showDialog, toast } from 'state'
  import { apiUrl } from 'util/appConstants'
  import isPartialSemver from 'util/semver'
  export let projectID: string
  export let organizationKey: string
  let extendedDiff = false
  $: project = $db.project[projectID]
  $: tags =
    !!project &&
    sortOn(Object.entries(project.snapshots || {}), '-1.created_at')
  let showCreateSnapshotForm = false

  function onDiffClick(e: CustomEvent<ApiDef.Change>) {
    const diff = e.detail
    if (!diff || !diff.path) {
      return
    }
    const isCombo = typeof diff.to === 'object'
    const catKeyA = diff.path.slice(1, -1).join('.')
    const catKeyB = diff.path.slice(0, -1).join('.')
    const translationKey = diff.path[diff.path.length - 1].split('_')[0]
    const cat = Object.values($db.category).find((c) => {
      if (c.project_id !== projectID) {
        return false
      }
      if (!c.translation_ids) {
        return false
      }
      if (isCombo && c.key === translationKey) {
        return true
      }
      if (c.key === catKeyA) {
        return true
      }
      if (c.key === catKeyB) {
        return true
      }
      return false
    })
    if (!cat || !cat.translation_ids) {
      toast({
        kind: 'error',
        message:
          'Failed to find a category for this diff. See the project-view to find it. Sorry for the inconvenience.',
        title: 'Category not found',
      })
      return
    }
    if (isCombo && cat.key === translationKey) {
      // TODO: this should show a dialog with the category and all the keys.
      // Tracked in [issue #10 - Add dialog for modifying keys within a category](https://github.com/runar-rkmedia/skiver/issues/10)
      toast({
        kind: 'warning',
        message:
          'Sorry, the view for this combination-item has not yet been implemented.',
        title: 'Not implemented yet',
      })

      return
    }
    const translation = cat.translation_ids
      .map((id) => $db.translation[id])
      .find((t) => t && t.key === translationKey)
    if (!translation) {
      toast({
        kind: 'error',
        message:
          'Failed to find a translation for this diff. See the project-view to find it. Sorry for the inconvenience.',
        title: 'Translation not found',
      })
      return
    }
    showDialog({
      kind: 'translation',
      id: translation.id,
      parent: cat.id,
      title: `Edit ${translation.title}`,
    })
  }
</script>

{#if project}
  <Dialogs {projectID} />
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
        <a
          target={projectID}
          href={apiUrl(
            `/export/${organizationKey}/${project.short_name}/f=i18n`
          )}>
          i18n
        </a>
      </div>
      The latest changes to the project, always available
      {#if tags.length}
        <TagDiff
          tagA={tags[0][0]}
          tagB={''}
          on:diffClick={onDiffClick}
          {projectID}
          format={extendedDiff ? 'raw' : 'i18n'} />
      {/if}
    </paper>
    {#each tags as [tag, snapshot], i}
      <paper>
        <div class="tagHeader">
          <h4>{tag}</h4>
          <div>
            <Collapse>
              <h5 name="title">Exported files:</h5>
              <div>
                <a
                  target={projectID}
                  href={apiUrl(
                    `/export/${organizationKey}/${project.short_name}/f=i18n&t=${tag}`
                  )}>
                  Dynamically generated export
                </a>
              </div>
              {#if snapshot.uploadMeta}
                {#each snapshot.uploadMeta as u}
                  <div>
                    <a target={projectID} href={u.url}>
                      Exported
                      {u.provider_name || u.provider_id}
                      <LocaleFlag locale={$db.locale[u.locale || '']} />
                      {$db.locale[u.locale || '']?.title} ({$db.locale[
                        u.locale || ''
                      ]?.[localeKeyToKeyOfLocaleMap[u.locale_key || '']]}) {isPartialSemver(
                        u.tag || ''
                      )
                        ? 'v'
                        : ''}{u.tag}
                    </a>
                  </div>
                {/each}
              {/if}
            </Collapse>
          </div>
        </div>
        {snapshot.description}
        {#if tags[i + 1]}
          <TagDiff
            tagA={tags[i + 1][0]}
            tagB={tag}
            on:diffClick={onDiffClick}
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
