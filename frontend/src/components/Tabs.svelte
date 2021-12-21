<script lang="ts">
  import type { api } from "../api";
  import Icon from "./Icon.svelte";

  import Tab from "./Tab.svelte";
  export let value = "request";
  const onClick = (s: keyof typeof api) => (e: MouseEvent) => (value = s);
</script>

<nav>
  <Tab active={value === "project"} on:click={onClick("project")}>
    <Icon icon={"gEndpoint"} color="tertiary" />
    Projects
  </Tab>
  <Tab active={value === "locale"} on:click={onClick("locale")}>
    <Icon icon={"gRequest"} color="tertiary" />
    Locales
  </Tab>
</nav>

<style>
  nav {
    display: flex;
    position: relative;
  }
  nav::after {
    content: "";
    width: 1px;

    position: absolute;
    left: 0;
    top: var(--size-6);
    bottom: var(--size-6);

    background-color: var(--color-red-500);
  }
</style>
