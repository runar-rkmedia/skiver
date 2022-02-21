<script lang="ts">
  export let locale: ApiDef.Locale
  let alpha2: string
  $: {
    // This is just temporary. The server will include alpha2 in the locale.
    if (locale) {
      if (locale.alpha2) {
        alpha2 = locale.alpha2.toUpperCase()
      }
      if (locale.ietf?.includes('-')) {
        alpha2 = locale.ietf.slice(-2).toUpperCase()
      } else {
        alpha2 = (locale.iso_639_1 || '').toUpperCase()
      }
    }
  }
</script>

{#if alpha2}
  <img
    class="flag"
    alt={`Flag for ${locale.ietf}`}
    src={`https://purecatamphetamine.github.io/country-flag-icons/3x2/${alpha2}.svg`} />
{/if}

<style>
  img {
    height: 2.5ex;
    width: 3.75ex;
    transform: translateY(0.3ex);
  }
</style>
