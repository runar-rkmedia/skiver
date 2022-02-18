<script lang="ts">
  import { db, api } from 'api'
  import Alert from 'components/Alert.svelte'
  import Spinner from 'components/Spinner.svelte'

  import TranslationItem from 'components/TranslationItem.svelte'
  // export let translation: ApiDef.Translation
  // export let translationValues: Record<string, ApiDef.TranslationValue>
  export let categoryKey: string

  export let translationKeyLike: string
  export let projectKey: string
  export let selectedLocale = ''

  $: locales = Object.values($db.locale)
  $: project =
    $db.project[projectKey] ||
    Object.values($db.project).find((t) => t.short_name === projectKey)
  $: category =
    project &&
    ($db.category[categoryKey] ||
      Object.values($db.category).find(
        (t) => t.key === categoryKey && t.project_id === project.id
      ))
  $: translation =
    category &&
    ($db.translation[translationKeyLike] ||
      Object.values($db.translation).find(
        (t) => t.key === translationKeyLike && t.category === category.id
      ))
  $: loadingCount = Object.entries($db.responseStates).filter(
    ([_, v]) => v.loading
  ).length
  $: notLoadingCount = Object.entries($db.responseStates).filter(
    ([_, v]) => !v.loading
  ).length
  $: maxLoadingCount = Object.keys($db.responseStates).length
</script>

<Spinner active={loadingCount > 0} />
{#if translation && project && locales && category}
  <paper>
    <TranslationItem {locales} {translation} {categoryKey} {projectKey} />
  </paper>
{:else if loadingCount}
  <p>Gathering information... hold on...</p>
  <progress value={notLoadingCount} max={maxLoadingCount} />
  {notLoadingCount} / {maxLoadingCount}
{:else}
  No translation found for input '{translationKeyLike}'

  {#if project}
    <p>Project: <strong>{project.title}</strong></p>

    {#if !category}
      The category {categoryKey} was not found. Perhaps you meant one of these?
      {#each (project.category_ids || []).map((cid) => $db.category[cid]) as c}
        {#if c}
          <div>
            <a
              href={`#embed/${projectKey}/${c.key}/${
                translationKeyLike || ''
              }`}>
              <h5>
                {c.title}
                <code>{c.key}</code>
              </h5>
            </a>
          </div>
        {:else}
          ???
        {/if}
      {/each}
    {:else}
      The translation {translationKeyLike} was not found. Perhaps you meant one of
      these?
      {#each (category.translation_ids || []).map((tid) => $db.translation[tid]) as t}
        {#if t}
          <div>
            <a href={`#embed/${projectKey}/${categoryKey}/${t.key}`}>
              <h5>
                {t.title}
                <code>{t.key}</code>
              </h5>
            </a>
          </div>
        {:else}
          ???
        {/if}
      {/each}
    {/if}
  {/if}
{/if}
{#if !projectKey || !translationKeyLike}
  <Alert kind="error">
    <h3 slot="title">Missing arguments</h3>
    This route should include:

    <ul>
      <li>
        ProjectKey, the shortname or id: you provided <code>{projectKey}</code>
      </li>
      <li>
        CategoryKey, the direct key for the translation <code
          >{categoryKey}</code>
      </li>
      <li>
        TranslationKey, the key for the translation you wish to modify<code
          >${translationKeyLike}</code>
      </li>
    </ul>
  </Alert>
{/if}

<p>You are viewing the embed version of this page.</p>
{#if project}
  <a href={'#project/' + project.id}>Click her to go to the project-view</a>
{/if}