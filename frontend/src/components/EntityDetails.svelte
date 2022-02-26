<script lang="ts">
  import formatDate from 'dates'

  export let entity: Partial<ApiDef.Entity>
</script>

{#if entity}
  <wrapper>
    <created>
      <created-at>
        Created:
        {formatDate(entity.createdAt)}
      </created-at>
      <created-by>
        by:
        {entity.createdBy}
      </created-by>
    </created>
    {#if entity.updatedAt}
      <updated>
        <updated-at>
          Updated:
          {formatDate(entity.updatedAt)}
        </updated-at>
        {#if entity.updatedBy}
          <updated-at>
            by:
            {entity.updatedBy}
          </updated-at>
        {/if}
      </updated>
    {/if}
  </wrapper>
{/if}

<style>
  created,
  updated,
  wrapper {
    display: block;
    font-size: small;
  }
</style>
