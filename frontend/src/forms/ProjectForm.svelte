<script lang="ts">
  import ApiResponseError from 'components/ApiResponseError.svelte'

  import Button from 'components/Button.svelte'
  import LocaleFlag from 'components/LocaleFlag.svelte'
  import LocaleSearch from 'components/LocaleSearch.svelte'
  import { createEventDispatcher } from 'svelte'
  import { api, db } from '../api'
  import { state, toast, toastApiErr } from '../state'
  export let project: ApiDef.Project | null = null
  export let shortNameReadOnly = false
  const dispatch = createEventDispatcher()

  async function createProject() {
    const [result, err] = await api.project.create($state.createProject)
    $state.createProject = {
      title: '',
      short_name: '',
      locales: {},
    }
    if (err) {
      toastApiErr(err as any)
      return
    }
    dispatch('complete', result)
  }
  async function updateProject() {
    if (!project) {
      return
    }
    const [result, err] = await api.project.update(project.id, {
      id: project.id,
      ...$state.createProject,
    })
    if (err) {
      toastApiErr(err as any)
      return
    }
    dispatch('complete', result)
  }
  function submit() {
    if (project) {
      updateProject()
      return
    }
    createProject()
  }
  function addLocale(idOrEvent: string | { detail: { item: { id: string } } }) {
    if (typeof idOrEvent !== 'string') {
      idOrEvent = idOrEvent.detail.item.id
    }
    if (!idOrEvent) {
      return
    }
    const p = project?.locales?.[idOrEvent]
    $state.createProject.locales = {
      ...$state.createProject.locales,
      [idOrEvent]: p
        ? {
            enabled: p.enabled,
            auto_translation: p.auto_translation,
            publish: p.publish,
          }
        : {
            enabled: true,
            auto_translation: false,
            publish: true,
          },
    }
  }
  function createInitial(p: ApiDef.Project) {
    return {
      short_name: p.short_name || '',
      title: p.title || '',
      description: p.description || '',
      locales: p.locales ? JSON.parse(JSON.stringify(p.locales)) : {},
    }
  }
  $: {
    if (project) {
      $state.createProject = createInitial(project)
    }
  }
  function hash(o: any) {
    // We should probably use something like deep-uqual or whatever some dependency is already using.
    if (!o) {
      return null
    }
    return JSON.stringify(createInitial(o))
  }
  $: projectHash = hash(project)
  $: inputHash = hash($state.createProject)

  $: didChange = projectHash !== inputHash
</script>

<form>
  <ApiResponseError key="project" />
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
          {#each Object.keys($state.createProject.locales || {}) as localeID}
            <tr class:created={!!project && !project?.locales?.[localeID]}>
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
                <LocaleFlag locale={$db.locale[localeID]} />
                {$db.locale[localeID]?.title}</td>
              <td
                class:modified={!!project &&
                  project.locales?.[localeID]?.enabled !==
                    $state.createProject.locales[localeID].enabled}>
                <input
                  type="checkbox"
                  bind:checked={$state.createProject.locales[localeID]
                    .enabled} />
              </td>
              <td
                class:modified={!!project &&
                  project.locales?.[localeID]?.auto_translation !==
                    $state.createProject.locales[localeID].auto_translation}>
                <input
                  type="checkbox"
                  bind:checked={$state.createProject.locales[localeID]
                    .auto_translation} />
              </td>
              <td
                class:modified={!!project &&
                  project.locales?.[localeID]?.publish !==
                    $state.createProject.locales[localeID].publish}>
                <input
                  type="checkbox"
                  bind:checked={$state.createProject.locales[localeID]
                    .publish} />
              </td>
            </tr>
          {:else}
            <p>No locales added.</p>
          {/each}

          {#if project}
            {#each Object.entries(project.locales || {}).filter(([k]) => !$state.createProject.locales[k]) as [localeID, setting]}
              <tr class="deleted">
                <td
                  ><Button
                    icon="add"
                    color="warning"
                    on:click={() => addLocale(localeID)} />
                  <LocaleFlag locale={$db.locale[localeID]} />
                  {$db.locale[localeID]?.title}</td>
                <td>
                  <input type="checkbox" checked={!!setting.enabled} />
                </td>
                <td>
                  <input type="checkbox" checked={!!setting.auto_translation} />
                </td>
                <td>
                  <input type="checkbox" checked={!!setting.publish} />
                </td>
              </tr>
            {/each}
          {/if}
        </tbody>
      </table>
      {#if project}
        <Button
          icon="clock"
          on:click={() => {
            if (!project) {
              return
            }
            $state.createProject.locales = createInitial(project).locales
          }}>Restore</Button>
      {/if}
    </paper>
    <paper class="locales-picker">
      <h5>
        Available Locales <small>Click on the locales you wish to add</small>
      </h5>
      <LocaleSearch
        locales={Object.values($db.locale).filter(
          (l) => !$state.createProject.locales?.[l.id]
        )}
        on:select={(e) => addLocale(e.detail.item.id)} />
    </paper>
  </div>
  <Button
    on:click={submit}
    icon="edit"
    type="submit"
    color="primary"
    disabled={!$state.createProject.title ||
      !$state.createProject.short_name ||
      !Object.keys($state.createProject.locales || {}).length ||
      (!!project && !didChange)}
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
  .created {
    background-color: var(--color-green-300);
  }
  .deleted {
    background-color: var(--color-red-300);
  }
  .created:nth-of-type(even) {
    background-color: var(--color-green-500);
  }
  tr:not(.created) .modified {
    background-color: var(--color-orange-300);
  }
</style>
