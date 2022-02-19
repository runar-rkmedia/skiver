<script lang="ts">
  import { db } from 'api'
  import Fuse from 'fuse.js'
  import { scrollToCategory } from 'util/scrollToCategory'
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
  let limit = 10

  $: data = {
    translationValues: new Fuse(translationValues, tvOptions),
    translations: new Fuse(translations, tOptions),
    categories: new Fuse(categories, cOptions),
  }
  $: result = {
    translationValues: data.translationValues.search(query, { limit }),
    translations: data.translations.search(query, { limit }),
    categories: data.categories.search(query, { limit }),
  }
</script>

<div>
  {#if result && query}
    <div class="resultswrapper">
      <paper class="resultblock">
        <h3>Translation-values</h3>
        {#each result.translationValues as { item: tv }}
          <div>
            {tv.value}
            <ScrollAnchor
              category={$db.category[
                $db.translation[tv.translation_id || '']?.category || ''
              ]} />
          </div>
        {/each}
      </paper>
      <paper class="resultblock">
        <h3>Translations</h3>
        {#each result.translations as { item: tv }}
          <div>
            {tv.key}
            <ScrollAnchor category={$db.category[tv.category || '']} />
          </div>
        {/each}
      </paper>
      <paper class="resultblock">
        <h3>Categories</h3>
        {#each result.categories as { item: category }}
          <div>
            <ScrollAnchor {category} />
          </div>
        {/each}
      </paper>
    </div>
  {/if}
</div>

<style>
  .resultswrapper {
    display: grid;
    gap: 10px;
    grid-template-columns: repeat(3, 1fr);
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
