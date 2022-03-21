<script lang="ts">
  import { db } from 'api'

  import formatDate from 'dates'

  export let entity: Partial<ApiDef.Entity>
  $: createdBy = entity.created_by && $db.simpleUser[entity.created_by]
  $: updatedBy = entity.updated_by && $db.simpleUser[entity.updated_by]
</script>

{#if entity}
  <wrapper class:deleted={entity.deleted}>
    <created>
      <created-at>
        Created:
        {formatDate(entity.created_at)}
      </created-at>
      {#if createdBy}
        by {createdBy}
      {/if}
    </created>
    {#if entity.updated_at}
      <updated>
        <updated-at>
          Updated:
          {formatDate(entity.updated_at)}
        </updated-at>
        {#if updatedBy}
          by {updatedBy}
        {/if}
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
