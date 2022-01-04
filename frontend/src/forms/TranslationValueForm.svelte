<script lang="ts">
  import Button from 'components/Button.svelte'
  import { createEventDispatcher } from 'svelte'
  import { api } from '../api'
  import { state, toast } from '../state'
  export let localeID: string
  export let translationID: string
  const dispatch = createEventDispatcher()
  async function onCreateTranslationValue() {
    $state.createTranslationValue.locale_id = localeID
    $state.createTranslationValue.translation_id = translationID
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
    const s = await api.translationValue.create($state.createTranslationValue)
    if (!s[1]) {
      dispatch('complete', s[0])
      $state.createTranslationValue = {
        locale_id: '',
        translation_id: '',
        value: '',
      }
    }
  }
</script>

<form>
  <!-- svelte-ignore a11y-autofocus -->
  <textarea
    autofocus
    rows={5}
    bind:value={$state.createTranslationValue.value}
    type="text"
    name="value" />
  <Button
    color="primary"
    type="submit"
    on:click={onCreateTranslationValue}
    icon={'submit'}>Submit</Button>
  <slot name="actions" />
</form>
