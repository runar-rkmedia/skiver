<script lang="ts">
  import Button from 'components/Button.svelte'
  import LocaleSearch from 'components/LocaleSearch.svelte'
  import { api, db } from '../api'
  import { state } from '../state'
  export let project: ApiDef.Project | null = null
  export let shortNameReadOnly = false

  function createProject() {
    api.project.create($state.createProject)
  }
  function updateProject() {
    if (!project) {
      return
    }
    api.project.update(project.id, $state.createProject)
  }
  function submit() {
    if (project) {
      updateProject()
      return
    }
    createProject()
  }
  $: {
    if (project) {
      $state.createProject = {
        short_name: project.short_name || '',
        title: project.title || '',
        description: project.description || '',
        locales: project.locales || {},
      }
    }
  }
</script>

<form>
  <label>
    Short-name
    <input
      readonly={shortNameReadOnly}
      bind:value={$state.createProject.short_name} />
  </label>
  <label>
    Title
    <input bind:value={$state.createProject.title} />
  </label>
  <label>
    Description
    <input bind:value={$state.createProject.description} />
  </label>

  <div class="locales">
    <paper>
      <h5>Active Locales</h5>
      <table>
        <thead>
          <th>Title</th>
          <th>Enabled</th>
          <th
            title="With auto-translate, non-existant translations for this locale will be sent to translation-services automatically. They will be based on the other translations provided for that key."
            >Auto-translate</th>
          <th>Publish</th>
        </thead>
        <tbody>
          {#each Object.entries($state.createProject.locales || {}) as [localeID, setting]}
            <tr>
              <td
                ><Button
                  icon="delete"
                  color="warning"
                  on:click={() => {
                    const l = $state.createProject.locales
                    const { [localeID]: _omit, ...rest } = l
                    {
                      $state.createProject.locales = rest
                    }
                  }} />
                {$db.locale[localeID]?.title}</td>
              <td>
                <input
                  type="checkbox"
                  bind:checked={$state.createProject.locales[localeID]
                    .enabled} />
              </td>
              <td>
                <input
                  type="checkbox"
                  bind:checked={$state.createProject.locales[localeID]
                    .auto_translation} />
              </td>
              <td>
                <input
                  type="checkbox"
                  bind:checked={$state.createProject.locales[localeID]
                    .publish} />
              </td>
            </tr>
          {:else}
            <p>No locales added.</p>
          {/each}
        </tbody>
      </table>
      <Button
        icon="clock"
        on:click={() => {
          $state.createProject.locales = project?.locales || {}
        }}>Restore</Button>
    </paper>
    <paper class="locales-picker">
      <h5>
        Available Locales <small>Click on the locales you wish to add</small>
      </h5>
      <LocaleSearch
        locales={Object.values($db.locale).filter(
          (l) => !$state.createProject.locales?.[l.id]
        )}
        on:select={(e) =>
          ($state.createProject.locales = {
            ...$state.createProject.locales,
            [e.detail.item.id]: {
              enabled: true,
              auto_translation: false,
              publish: true,
            },
          })} />
    </paper>
  </div>
  <Button on:click={submit} icon="edit" type="submit" color="primary"
    >{!!project ? 'Update' : 'Create'} Project</Button>
  <slot name="actions" />
</form>

<style>
  .locales {
    margin-top: var(--size-8);
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 20px;
  }
  .locales-picker {
    display: flex;
    flex-direction: column;
    min-height: 250px;
  }
  th {
    white-space: nowrap;
    padding-inline: var(--size-2);
  }
  th:first-of-type {
    width: 100%;
  }
</style>
