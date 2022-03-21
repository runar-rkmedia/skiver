<script lang="ts">
  import Button from 'components/Button.svelte'
  import { createEventDispatcher } from 'svelte'
  import { api, db } from '../api'
  import { state, toast } from '../state'
  export let localeID: string
  export let existingID: string = ''
  export let translationID: string
  export let withContext = false
  const dispatch = createEventDispatcher()
  async function onCreateTranslationValue() {
    let s: Awaited<ReturnType<typeof api.translationValue.create>>
    $state.createTranslationValue.locale_id = localeID
    $state.createTranslationValue.translation_id = translationID
    if (existingID) {
      // Update
      const p = $state.createTranslationValue
      if (p.context_key == '') {
        delete p.context_key
      }
      s = await api.translationValue.update(existingID, {
        id: existingID,
        ...p,
      })
    } else {
      // Create

      if (!$state.createTranslationValue.locale_id) {
        toast({
          title: 'missing argument',
          message: 'locale was not set',
          kind: 'warning',
        })
        return
      }
      if (!$state.createTranslationValue.translation_id) {
        toast({
          title: 'missing argument',
          message: 'translation was not set',
          kind: 'warning',
        })
        return
      }
      s = await api.translationValue.create($state.createTranslationValue)
    }
    if (!s) {
      console.error('result was not set')
      return
    }
    if (!s[1]) {
      dispatch('complete', s[0])
      $state.createTranslationValue = {
        locale_id: '',
        translation_id: '',
        value: '',
      }
    }
  }
  $: loading = $db.responseStates.translationValue.loading
</script>

<form>
  {#if withContext}
    <label
      >Context

      <input
        name="context"
        bind:value={$state.createTranslationValue.context_key} />
    </label>
  {/if}
  <!-- svelte-ignore a11y-autofocus -->
  <textarea
    autofocus
    rows={5}
    bind:value={$state.createTranslationValue.value}
    type="text"
    name="value" />
  <div>
    <slot name="preview">No preview available</slot>
  </div>
  <Button
    color="primary"
    type="submit"
    disabled={loading}
    on:click={onCreateTranslationValue}
    icon={loading ? 'loading' : 'submit'}>Submit</Button>
  <Button
    slot="actions"
    color="secondary"
    disabled={loading}
    on:click={() => {
      $state.openTranslationValueForm = ''
      dispatch('cancel')
    }}
    icon={'cancel'}>Cancel</Button>
  <slot name="actions" />
</form>
