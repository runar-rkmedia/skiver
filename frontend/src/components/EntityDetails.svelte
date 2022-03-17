<script lang="ts">
  import formatDate from 'dates'

  export let entity: Partial<ApiDef.Entity>
</script>

{#if entity}
  <wrapper class:deleted={entity.deleted}>
    <created>
      <created-at>
        Created:
        {formatDate(entity.created_at)}
      </created-at>
    </created>
    {#if entity.updated_at}
      <updated>
        <updated-at>
          Updated:
          {formatDate(entity.updated_at)}
        </updated-at>
      </updated>
    {/if}
    {#if entity.deleted}
      <deleted>
        <deleted-at>
          Scheduled for deletion:
          {formatDate(entity.deleted)}
        </deleted-at>
      </deleted>
    {/if}
  </wrapper>
{/if}

<style>
  created,
  updated,
  deleted,
  wrapper {
    display: block;
  }
  wrapper {
    font-size: small;
  }
  deleted {
    background-color: var(--color-danger);
    color: var(--color-grey-100);
    font-size: 1.2rem;
  }
</style>
