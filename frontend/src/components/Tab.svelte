<script lang="ts">
  export let active: boolean
  export let disabled: boolean = false
  export let href: string
</script>

<li class:active><a {href} {disabled} class=""><slot /></a></li>

<style>
  .active a {
    background-color: hsla(240, 60%, 80%, 20%);
  }
  a:hover:not(:disabled) {
    background-color: hsla(240, 60%, 80%, 30%);
  }
  .active a:hover:not(:disabled) {
    background-color: hsla(240, 60%, 80%, 40%);
  }
  a:disabled {
    cursor: unset;
    opacity: 0.5;
  }
  a {
    display: block;
    font-weight: bold;
    cursor: pointer;
    font-family: revert;
    position: relative;
    transition: transform 60ms ease-in-out;
    color: revert;
    color: var(--color-grey-100);
    /* background-color: var(--color-grey-700); */
    padding-inline: 12px;
    height: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    text-decoration: none;
  }
</style>
