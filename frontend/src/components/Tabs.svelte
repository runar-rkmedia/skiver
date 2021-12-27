<script lang="ts">
  import { state } from 'state'

  import Icon from './Icon.svelte'

  import Tab from './Tab.svelte'
</script>

<nav>
  <Tab
    active={$state.tab === 'project'}
    on:click={() => ($state.tab = 'project')}
  >
    <Icon icon={'gEndpoint'} color="tertiary" />
    Projects
  </Tab>
  <Tab
    active={$state.tab === 'locale'}
    on:click={() => ($state.tab = 'locale')}
  >
    <Icon icon={'gRequest'} color="tertiary" />
    Locales
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
