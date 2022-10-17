<script lang="ts">
  import TranslationValueRow from './TranslationValueRow.svelte'
  import EntityDetails from './EntityDetails.svelte'
  import { api, db } from 'api'
  import Button from './Button.svelte'
  import ContextForm from './ContextForm.svelte'
  import TranslationRef from './TranslationRef.svelte'
  import { showDialog } from 'state'
  export let translation: ApiDef.Translation
  /** Map by locale-id */
  // export let translationValues: Record<string, ApiDef.TranslationValue>
  export let categoryKey: string
  export let projectKey: string

  export let locales: ApiDef.Locale[]
  $: translationValues = (translation?.value_ids || []).reduce((r, tvid) => {
    const tv = $db.translationValue[tvid]
    if (!tv || !tv.locale_id) {
      return r
    }
    r[tv.locale_id] = tv
    return r
  }, {} as Record<string, ApiDef.TranslationValue>)
  let contextKeys: string[] = []
  $: contextKeys = Array.from(
    Object.values(translationValues || {}).reduce((r, tv) => {
      if (!tv || !tv.context) {
        return r
      }
      for (const c of Object.keys(tv.context)) {
        r.add(c)
      }
      return r
    }, new Set<string>())
  )
  let addContext = false
</script>

{#if translation}
  <div class="desc">
    <h4>
      <slot name="categoryHeader" {categoryKey} {translation}>
        <code>
          {`${categoryKey !== '' ? categoryKey + '.' : ''}${translation.key}`}
        </code>
        {translation.title}
      </slot>
      <Button
        icon="edit"
        on:click={() =>
          showDialog({
            kind: 'editTranslation',
            id: translation.id,
            title: `Edit ${translation.title}`,
          })}>Edit</Button>
    </h4>
    <div>
      <small />
      <EntityDetails entity={translation} />
    </div>
  </div>
  <p>
    {translation.description || ''}
  </p>
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
            {translation}
            {locale}
            translationValue={translationValues[locale.id]} />
        {/each}
      {/if}
    </tbody>
  </table>

  <hr />
  <h5>
    Contexts

    {#if !addContext}
      <Button icon="create" on:click={() => (addContext = true)}>Add</Button>
    {/if}
  </h5>

  <!-- <p> -->
  <!--   Contexts are variations of the default value, often used programmatically. -->
  <!--   If not set, the value will typically fall back to the default value -->
  <!-- </p> -->

  {#if addContext}
    <paper>
      <ContextForm
        {categoryKey}
        {projectKey}
        {translation}
        {locales}
        {translationValues}
        on:complete={() => (addContext = false)}
        on:abort={() => (addContext = false)} />
    </paper>
  {/if}
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
            {contextKey}
            {locale}
            translationValue={translationValues[locale.id]} />
        {/each}
      </tbody>
    </table>
  {/each}
  {#if translation?.variables}
    <h6>Variables</h6>
    <code>{JSON.stringify(translation.variables, null, 2)}</code>
  {/if}
  {#if translation?.references}
    <h6>References</h6>
    <ul>
      {#each translation.references as ref}
        <li>
          <TranslationRef {ref} />
        </li>
      {/each}
    </ul>
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
  code {
    white-space: pre;
    display: inline-block;
  }
</style>
