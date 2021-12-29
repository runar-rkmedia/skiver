<script lang="ts">
  import router from 'util/router'
  import Icon from './Icon.svelte'
  import Tab from './Tab.svelte'
</script>

<nav>
  <Tab href="#project/" active={$router.hash.startsWith('#project/')}>
    <div>
      <Icon icon={'gEndpoint'} color="tertiary" />
      Projects
    </div>
  </Tab>
  <Tab href="#locale/" active={$router.hash.startsWith('#locale/')}>
    <div>
      <Icon icon={'gRequest'} color="tertiary" />
      Locales
    </div>
  </Tab>
</nav>

<style>
  nav {
    display: flex;
    position: relative;
  }
  nav::after {
    content: '';
    width: 1px;

    position: absolute;
    left: 0;
    top: var(--size-6);
    bottom: var(--size-6);

    background-color: var(--color-red-500);
  }
</style>
