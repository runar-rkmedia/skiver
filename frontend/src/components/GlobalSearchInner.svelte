<script lang="ts">
  import { db } from 'api'
  import Fuse from 'fuse.js'
  import Embed from 'pages/Embed.svelte'
  import { state } from 'state'
  import { scale } from 'svelte/transition'
  import Button from './Button.svelte'
  import Dialog from './Dialog.svelte'
  import ScrollAnchor from './ScrollAnchor.svelte'
  export let project: ApiDef.Project

  export let query = 'ich'
  const tvOptions = {
    limit: 5,
    keys: ['value'],
  }
  const tOptions = {
    keys: ['key'],
  }
  const cOptions = {
    keys: ['key'],
  }
  $: categories = (project.category_ids || []).map((id) => $db.category[id])
  $: translations = categories.reduce((r, c) => {
    if (!c) {
      return r
    }
    if (!c.translation_ids?.length) {
      return r
    }
    return [...r, ...c.translation_ids.map((id) => $db.translation[id])]
  }, [])
  $: translationValues = translations.reduce((r, t) => {
    if (!t) {
      return r
    }
    if (!t.value_ids?.length) {
      return r
    }
    return [...r, ...t.value_ids.map((id) => $db.translationValue[id])]
  }, [])
  let limit = 20

  $: data = {
    translationValues: new Fuse(translationValues, tvOptions),
    translations: new Fuse(translations, tOptions),
    categories: new Fuse(categories, cOptions),
  }
  $: result = {
    translationValues: data.translationValues.search(query),
    translations: data.translations.search(query),
    categories: data.categories.search(query),
  }
  let categoryKey: string = ''
  let translationID: string = ''
</script>

<div>
  {#if result && query}
    <div class="resultswrapper">
      {#if $state.searchInTranslationValues && result.translationValues.length}
        <paper class="resultblock tv" transition:scale>
          <h3>
            Translation-values {result.translationValues.length}
            <Button
              icon="closeCross"
              on:click={() => ($state.searchInTranslationValues = false)} />
          </h3>
          {#each result.translationValues.slice(0, limit) as { item: tv } (tv.id)}
            <div>
              <Button
                on:click={() => {
                  categoryKey =
                    $db.translation[tv.translation_id || '']?.category || ''
                  translationID = tv.translation_id || ''
                }}>
                {tv.value}
              </Button>
            </div>
          {/each}
        </paper>
      {/if}
      {#if $state.searchInTrasnaltions && result.translations.length}
        <paper class="resultblock t" transition:scale>
          <h3>
            Translations {result.translations.length}
            <Button
              icon="closeCross"
              on:click={() => ($state.searchInTrasnaltions = false)} />
          </h3>
          {#each result.translations.slice(0, limit) as { item: tv } (tv.id)}
            <div>
              <Button
                on:click={() => {
                  categoryKey = tv.category || ''
                  translationID = tv.id || ''
                }}>
                {tv.title}
              </Button>
            </div>
          {/each}
        </paper>
      {/if}
      {#if $state.searchInCategories && result.categories.length}
        <paper class="resultblock c" transition:scale>
          <h3>
            Categories {result.categories.length}

            <Button
              icon="closeCross"
              on:click={() => ($state.searchInCategories = false)} />
          </h3>
          {#each result.categories.slice(0, limit) as { item: category } (category.id)}
            <div>
              <ScrollAnchor {category} />
            </div>
          {/each}
        </paper>
      {/if}
    </div>
    {#if translationID || categoryKey}
      <Dialog
        on:clickClose={() => {
          translationID = ''
          categoryKey = ''
        }}>
        <Embed
          noHeader={true}
          {categoryKey}
          projectKey={project.id}
          translationKeyLike={translationID}>
          <h4 slot="categoryHeader" let:category let:translation>
            <ScrollAnchor
              {category}
              on:scrollTo={() => {
                translationID = ''
                categoryKey = ''
              }}>
              <code>
                {[category?.key, translation?.key].filter(Boolean).join('.')}
              </code>
            </ScrollAnchor>
          </h4>
        </Embed>
      </Dialog>
    {/if}
    {#if !result.translationValues.length && !result.translations.length && !result.categories.length}
      No result. Try refining your search.
    {/if}
    {#if !$state.searchInCategories}
      <Button
        color="secondary"
        on:click={() => ($state.searchInCategories = true)}
        >Show results for categories {result.categories.length}</Button>
    {/if}
    {#if !$state.searchInTrasnaltions}
      <Button
        color="secondary"
        on:click={() => ($state.searchInTrasnaltions = true)}
        >Show results for translations {result.translations.length}</Button>
    {/if}
    {#if !$state.searchInTranslationValues}
      <Button
        color="secondary"
        on:click={() => ($state.searchInTranslationValues = true)}
        >Show results for translation-values {result.translationValues
          .length}</Button>
    {/if}
  {/if}
</div>

<style>
  :global(.resultswrapper button) {
    padding: 0;
    text-align: inherit;
  }
  h3 {
    display: flex;
    justify-content: space-between;
  }

  .resultswrapper {
    display: flex;
    flex-wrap: wrap;
    gap: var(--size-4);
  }
  .resultswrapper paper {
    flex: 1;
    min-width: 200px;
    transition: flex 150ms var(--easing-standard);
  }
  .resultswrapper paper {
    max-width: 45rem;
  }
  .resultswrapper paper.tv {
    flex: 3;
    min-width: 300px;
  }
  .resultblock > div:nth-child(odd) {
    background-color: var(--color-grey-300);
  }
  .resultblock > div {
    display: flex;
    justify-content: space-between;
  }
  paper > div {
    display: flex;
  }
</style>
