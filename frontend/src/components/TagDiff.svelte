<script lang="ts">
  import { api } from '../api'
  import Alert from './Alert.svelte'
  import Collapse from './Collapse.svelte'
  import Icon from './Icon.svelte'
  import JsonDetail from './JsonDetail.svelte'
  export let tagA: string
  export let tagB: string
  export let format: 'i18n' | 'raw' = 'i18n'
  export let projectID: string
  let response: ApiDef.DiffResponse | undefined
  let err: ApiDef.APIError | null | undefined

  $: {
    if (projectID && (tagA || tagB) && format) {
      api
        .diff({
          format,
          a: { project_id: projectID, tag: tagA },
          b: { project_id: projectID, tag: tagB },
        })
        .then((r) => {
          response = r[0]?.data
          err = r[1]
        })
    }
  }
  $: areEqual =
    !!response && response.a?.hash && response.a.hash === response.b?.hash
  $: summary = response?.diff?.reduce(
    (r, d) => {
      if (!d.type) {
        return r
      }
      r[d.type] = (r[d.type] || 0) + 1
      return r
    },
    { delete: 0, update: 0, create: 0 }
  )

  function shortenPath(a: string[] | undefined, b?: string[]) {
    if (a && b) {
      for (let i = 0; i < a.length; i++) {
        const ap = a[i]
        const bp = b[i]
        if (ap === bp) {
          continue
        }
        return ' \u21b3 ' + a.slice(i).join('.\u200B')
      }
    }
    const s = (a || []).join('.\u200B')
    return s
  }
</script>

{#if err}
  <Alert kind="error">
    <pre>{JSON.stringify(err)}</pre>
  </Alert>
{/if}
{#if response}
  <Collapse let:show>
    <h5 slot="title">
      <span>
        {#if areEqual}
          No changes
        {:else}
          Changes:
          {#if summary}
            <span class="summary">
              {#if summary.delete}
                <span class="col-delete">
                  <Icon icon="delete" /> Deleted: {summary.delete}
                </span>
              {/if}
              {#if summary.create}
                <span class="col-create">
                  <Icon icon="create" /> Created: {summary.create}
                </span>
              {/if}
              {#if summary.update}
                <span class="col-update">
                  <Icon icon="edit" /> Updated: {summary.update}
                </span>
              {/if}
            </span>
          {/if}
        {/if}
      </span>
      <small>
        {#if tagA}
          since {tagA}
        {:else}
          since latest unreleased changes
        {/if}
      </small>
    </h5>
    {#if show}
      {#if response.diff}
        <table>
          <thead>
            <th class="type" />
            <th class="path">Path</th>
            <th class="content">Change</th>
          </thead>
          <tbody>
            {#each response.diff as d, i}
              <tr
                class={d.type}
                class:diff={true}
                title={JSON.stringify(d, null, 2)}>
                <td class="type">
                  {#if d.type === 'delete'}
                    <span class="col-delete">
                      <Icon icon="delete" />
                    </span>
                  {:else if d.type === 'create'}
                    <span class="col-create">
                      <Icon icon="create" />
                    </span>
                  {:else if d.type === 'update'}
                    <span class="col-update">
                      <Icon icon="edit" />
                    </span>
                  {/if}
                </td><td class="path"
                  >{shortenPath(d.path, response.diff[i - 1]?.path)}</td>
                <td class="content">
                  {#if d.type === 'delete'}
                    <span class="delete from">
                      <JsonDetail json={JSON.stringify(d.from, null, 2)} />
                    </span>
                  {:else if d.type === 'create'}
                    <JsonDetail json={JSON.stringify(d.to, null, 2)} />
                  {:else if d.type === 'update'}
                    {d.from} <Icon icon="longArrowRight" /> {d.to}
                  {/if}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      {/if}
      <div class="hashes">
        {#if areEqual}
          <img
            title={`SHA ${response.a?.tag || response.b?.tag || ''}: ${
              response.a?.hash || response.b?.hash
            }`}
            alt="Identicon for tag"
            src={'data:image/png;base64,' +
              (response.a?.identi_hash || response.b?.identi_hash)} />
        {:else}
          <img
            title={`SHA ${response.b?.tag || ''}: ${response.b?.hash}`}
            alt="Identicon for tag B"
            src={'data:image/png;base64,' + response.b?.identi_hash} />
          <img
            title={`SHA ${response.a?.tag || ''}: ${response.a?.hash}`}
            alt="Identicon for tag A"
            src={'data:image/png;base64,' + response.a?.identi_hash} />
        {/if}
      </div>
    {/if}
  </Collapse>
{/if}

<style>
  .diff {
    align-items: center;
  }
  .content {
    padding-inline: var(--size-2);
    overflow-x: auto;
  }
  table {
    table-layout: fixed;
  }
  tr.create {
    background-color: var(--color-green-500);
  }
  tr.update {
    background-color: var(--color-yellow-500);
  }
  tr.delete {
    background-color: var(--color-red-500);
  }
  tr.create:nth-of-type(odd) {
    background-color: var(--color-green-300);
  }
  tr.update:nth-of-type(odd) {
    background-color: var(--color-orange-500);
  }
  tr.delete:nth-of-type(odd) {
    background-color: var(--color-red-300);
  }
  th.type {
    width: 20px;
  }
  th.path {
    width: 30%;
  }
  tr.diff {
    padding-inline: var(--size-2);
  }
  td.path {
    font-family: var(--font-mono);
    word-break: break;
    max-width: 30%;
  }
  .hashes {
    padding-top: var(--size-4);
    display: flex;
    justify-content: center;
    gap: var(--size-3);
  }
  .hashes img {
    border-radius: var(--radius-xl);
    border: 2px solid var(--color-primary);
  }
  .hashes img:nth-of-type(2) {
    transform: scale(0.8);
    opacity: 0.8;
  }
  h5 {
    display: flex;
    width: 100%;
    justify-content: space-between;
    align-items: center;
    font-weight: var(--font-medium);
  }
  .summary {
    padding-inline-start: var(--size-2);
  }
  .summary span:not(:last-of-type) {
    margin-inline-end: var(--size-3);
  }
</style>
