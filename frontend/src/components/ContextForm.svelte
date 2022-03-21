<script lang="ts">
  import { createEventDispatcher } from 'svelte'

  import Button from './Button.svelte'
  import TranslationValueRow from './TranslationValueRow.svelte'
  export let locales: ApiDef.Locale[]
  export let translation: ApiDef.Translation
  export let categoryKey: string
  export let projectKey: string
  export let translationValues: Record<string, ApiDef.TranslationValue>
  const dispatch = createEventDispatcher()

  let contextKey = ''
  let step = 0
  function next() {
    if (!contextKey) {
      return
    }
    step++
  }
</script>

<h5>
  New context :
  <code>
    {contextKey}
  </code>
</h5>
{#if step === 0}
  <label>
    <!-- svelte-ignore a11y-autofocus -->
    <input autofocus type="text" name="text" bind:value={contextKey} />
    <Button color="primary" on:click={next}>Next</Button>
  </label>
  <Button color="secondary" on:click={() => dispatch('abort')}>Abort</Button>
{/if}
{#if contextKey && step === 1}
  <Button color="secondary" on:click={() => step--}>Back</Button>
  <table>
    <thead>
      <th>Language</th>
      <th>Value</th>
    </thead>
    <tbody>
      {#each locales as locale}
        <TranslationValueRow
          on:complete
          {categoryKey}
          {projectKey}
          {translation}
          {contextKey}
          translationValue={translationValues[locale.id]}
          {locale} />
      {/each}
    </tbody>
  </table>
{/if}
