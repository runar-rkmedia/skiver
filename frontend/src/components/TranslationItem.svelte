<script lang="ts">
  import TranslationValueRow from './TranslationValueRow.svelte'
  import EntityDetails from './EntityDetails.svelte'
  export let translation: ApiDef.Translation
  /** Map by locale-id */
  export let translationValues: Record<string, ApiDef.TranslationValue>
  export let categoryKey: string
  export let projectKey: string
  export let selectedLocale = ''

  export let locales: ApiDef.Locale[]
  $: contextKeys = Array.from(
    Object.values(translationValues || {}).reduce((r, tv) => {
      if (!tv.context) {
        return r
      }
      for (const c of Object.keys(tv.context)) {
        r.add(c)
      }
      return r
    }, new Set<string>())
  )
</script>

{#if translation}
  <div class="desc">
    <h4>
      <code>
        {`${categoryKey !== '___root___' ? categoryKey + '.' : ''}${
          translation.key
        }`}
      </code>
      {translation.title}
    </h4>
    <div>
      <small>
        {translation.description || ''}
      </small>
      <EntityDetails entity={translation} />
    </div>
  </div>
  Selected {selectedLocale}
  <table>
    <thead>
      <th>Language</th>
      <th>Value</th>
    </thead>
    <tbody>
      {#if locales && translationValues}
        {#each locales as locale}
          <TranslationValueRow
            {categoryKey}
            {projectKey}
            bind:selectedLocale
            {translation}
            {locale}
            translationValue={translationValues[locale.id]} />
        {/each}
      {/if}
    </tbody>
  </table>

  {#if contextKeys && contextKeys.length}
    <hr />
    <h5>Contexts</h5>
    <p>
      Contexts are variations of the default value, often used programmatically.
      If not set, the value will typically fall back to the default value
    </p>

    {#each contextKeys as contextKey}
      <h6>
        {contextKey}
      </h6>
      <table>
        <thead>
          <th>Language</th>
          <th>Value</th>
        </thead>
        <tbody>
          {#each locales as locale}
            <TranslationValueRow
              {categoryKey}
              {projectKey}
              {translation}
              bind:selectedLocale
              {contextKey}
              {locale}
              translationValue={translationValues[locale.id]} />
          {/each}
        </tbody>
      </table>
    {/each}
  {/if}
  {#if translation?.variables}
    <h6>Variables</h6>
    {#each Object.entries(translation.variables) as [k, v]}
      <var-pair>
        <var-key>{k}</var-key>
        <var-value>{JSON.stringify(v)}</var-value>
      </var-pair>
      <!-- TODO: allow editing -->
    {/each}
  {/if}
{:else}
  ... (no translation???)
{/if}

<style>
  .desc {
    display: flex;
    justify-content: space-between;
  }
  .desc > * {
    display: block;
  }
  .desc > :last-child {
    text-align: right;
  }
  th {
    padding-inline: var(--size-4);
    padding-block: var(--size-2);
  }
  var-pair {
    display: block;
  }
  var-key::after {
    content: ': ';
  }
</style>
