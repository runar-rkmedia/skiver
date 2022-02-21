<script lang="ts">
  import Fuse from 'fuse.js'
  import { createEventDispatcher } from 'svelte'
  import LocaleFlag from './LocaleFlag.svelte'
  export let locales: ApiDef.Locale[]

  export let limit = 20
  let query = ''

  $: fuseLocales = new Fuse<ApiDef.Locale>(locales || [], {
    minMatchCharLength: -1,
    keys: ['title', 'ietf', 'iso_639_3', 'iso_639_2', 'iso_639_1'],
  })
  $: result = !!query
    ? fuseLocales.search(query, { limit })
    : locales.slice(0, limit).map((l) => ({ item: l }))
  const dispatch = createEventDispatcher<{ select: { item: ApiDef.Locale } }>()
</script>

<input
  disabled={!locales.length}
  placeholder="Spanish, es, es-MX"
  type="search"
  bind:value={query} />

{#each result as loc}
  <button
    class="btn-reset"
    on:click={() => dispatch('select', { item: loc.item })}>
    <div>
      <LocaleFlag locale={loc.item} />
      {loc.item.title}
    </div>

    <div>
      <small>
        {loc.item.iso_639_2} - {loc.item.ietf}
      </small>
    </div>
  </button>
{:else}
  <p>No locales</p>
{/each}

<style>
  button {
    display: flex;
    width: 100%;
    justify-content: space-between;
  }
  button:nth-of-type(even) {
    background-color: var(--color-grey-300);
  }
  button:hover {
    outline: 1px dashed black;
    z-index: 1;
  }
</style>
