<script lang="ts">
  import { db } from 'api'
  import Alert from 'components/Alert.svelte'
  import Button from 'components/Button.svelte'
  import Spinner from 'components/Spinner.svelte'

  import TranslationItem from 'components/TranslationItem.svelte'
  import UserButton from 'components/UserButton.svelte'
  import CategoryForm from 'forms/CategoryForm.svelte'
  import TranslationForm from 'forms/TranslationForm.svelte'
  import { state } from 'state'
  import { camelCaseToTitleCase } from 'util/titleCase'

  // export let translation: ApiDef.Translation
  // export let translationValues: Record<string, ApiDef.TranslationValue>
  export let categoryKey: string

  export let translationKeyLike: string
  export let projectKey: string

  let showCategoryForm = false
  let showTranslationForm = false
  export let noHeader = false

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

{#if !noHeader}
  <div class="user-welcome">
    <UserButton />
  </div>
{/if}
<Spinner active={loadingCount > 0} />
{#if locales}
  {#if project}
    <!-- Has Project -->
    {#if !noHeader}
      <h1>{project.title} <small><code>{project.short_name}</code></small></h1>
      {#if project.description}
        <p>{project.description}</p>
      {/if}
    {/if}
    {#if category}
      <!-- Has Category -->
      {#if !noHeader}
        <h2>{category.title} <small><code>{category.key}</code></small></h2>
        {#if category.description}
          <p>{category.description}</p>
        {/if}
      {/if}
      {#if translation}
        <!-- Has translation -->
        <paper>
          {#if $$slots.categoryHeader}
            <TranslationItem
              {locales}
              {translation}
              categoryKey={category.key || ''}
              {projectKey}>
              <slot
                slot="categoryHeader"
                let:translation
                name="categoryHeader"
                {category}
                {translation} />
            </TranslationItem>
          {:else}
            <TranslationItem
              {locales}
              {translation}
              categoryKey={category.key || ''}
              {projectKey} />
          {/if}
        </paper>
        <!-- End has translation -->
      {:else}
        <!-- No Translation -->
        <p>
          {#if translationKeyLike}
            The translation
            <code>{translationKeyLike}</code>
            in category
            <code>{category.key}</code>
            was not found.
          {:else}
            No translation-key specified for category
            <code>{category.key}</code>
          {/if}

          {#if category.translation_ids?.length}
            Perhaps you meant one of these?
          {:else}
            Do you wish to create it, perhaps?
          {/if}
        </p>
        {#if showTranslationForm}
          <paper>
            <TranslationForm
              categoryID={category.id}
              on:complete={() => (showTranslationForm = false)}>
              <Button
                slot="actions"
                color="secondary"
                icon="cancel"
                on:click={() => (showTranslationForm = false)}>Cancel</Button>
            </TranslationForm>
          </paper>
        {:else}
          <Button
            color="primary"
            icon="create"
            on:click={() => {
              if ($state.createTranslation.key !== translationKeyLike) {
                $state.createTranslation.key = translationKeyLike
                $state.createTranslation.title =
                  camelCaseToTitleCase(translationKeyLike)
                $state.createTranslation.description = ''
              }
              showTranslationForm = true
            }}>Create new translation</Button>
        {/if}
        <table class="resultblock">
          <caption>List of translations within category</caption>
          <thead>
            <th>Title</th>
            <th>Key</th>
          </thead>
          <tbody>
            {#each (category.translation_ids || []).map((tid) => $db.translation[tid]) as t}
              {#if t}
                <tr title={t.description}>
                  <td>
                    <a href={`#embed/${projectKey}/${categoryKey}/${t.key}`}>
                      {t.title}
                    </a>
                  </td>
                  <td>
                    <a href={`#embed/${projectKey}/${categoryKey}/${t.key}`}>
                      {t.key}
                    </a>
                  </td>
                </tr>
              {/if}
            {/each}
          </tbody>
        </table>
        <!-- End No Translation -->
      {/if}
      <!-- End has Category -->
    {:else}
      <!-- No Category -->
      <p>
        {#if categoryKey}
          The Category <code>{categoryKey}</code> was not found.
        {:else}
          Not category-key in input found.
        {/if}
        {#if project.category_ids?.length}
          Perhaps you meant one of following categories, or maybe you want to
          create a new category?
        {:else}
          Do you wish to create it, perhaps?
        {/if}
      </p>
      {#if showCategoryForm}
        <paper>
          <CategoryForm
            projectID={project.id}
            on:complete={() => (showCategoryForm = false)}>
            <Button
              slot="actions"
              color="secondary"
              icon="cancel"
              on:click={() => (showCategoryForm = false)}>Cancel</Button>
          </CategoryForm>
        </paper>
      {:else}
        <Button
          color="primary"
          icon="create"
          on:click={() => {
            if ($state.createCategory.key !== categoryKey) {
              $state.createCategory.key = categoryKey
              $state.createCategory.title = camelCaseToTitleCase(categoryKey)
              $state.createCategory.description = ''
            }
            showCategoryForm = true
          }}>Create new category</Button>
        <table class="resultblock">
          <caption>List of categories within project</caption>
          <thead>
            <th>Title</th>
            <th>Key</th>
          </thead>
          <tbody>
            {#each (project.category_ids || []).map((cid) => $db.category[cid]) as c}
              {#if c}
                <tr title={c.description}>
                  <td>
                    <a
                      href={`#embed/${projectKey}/${c.key}/${
                        translationKeyLike || ''
                      }`}>
                      {c.title}
                    </a>
                  </td>
                  <td>
                    <a
                      href={`#embed/${projectKey}/${c.key}/${
                        translationKeyLike || ''
                      }`}>
                      {c.key}
                    </a>
                  </td>
                </tr>
              {/if}
            {/each}
          </tbody>
        </table>
      {/if}

      <!-- End No Category -->
    {/if}
    <!-- End has Project -->
  {:else}
    <!-- No Project -->
    {#if projectKey}
      Project <code>{projectKey}</code> was not found.
    {:else}
      Not project specified
    {/if}
    <!-- End No Project -->
  {/if}
  <!-- End has Locales -->
{:else}
  <!-- No Locales -->
  <!-- End No Locales -->
{/if}
{#if !(translation && project && locales && category) && loadingCount}
  <p>Gathering information... hold on...</p>
  <progress value={notLoadingCount} max={maxLoadingCount} />
  {notLoadingCount} / {maxLoadingCount}
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

<slot name="after" {project} />

<style>
  .user-welcome {
    position: absolute;
    right: var(--size-2);
    top: var(--size-2);
    display: flex;
    justify-content: flex-end;
    align-items: center;
    padding-inline-end: 20px;
  }
  h1 {
    margin-block-start: 0;
  }
  .resultblock {
    background-color: var(--color-grey-100);
  }
  .resultblock a {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    gap: 20px;
  }
  caption {
    margin-block: var(--size-2);
  }
</style>
