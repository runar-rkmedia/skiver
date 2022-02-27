<script lang="ts">
  import formatDate from 'dates'

  export let entity: Partial<ApiDef.Entity>
</script>

{#if entity}
  <wrapper>
    <created>
      <created-at>
        Created:
        {formatDate(entity.created_at)}
      </created-at>
      <created-by>
        by:
        {entity.created_by}
      </created-by>
    </created>
    {#if entity.updated_by}
      <updated>
        <updated-at>
          Updated:
          {formatDate(entity.updated_at)}
        </updated-at>
        {#if entity.updated_by}
          <updated-at>
            by:
            {entity.updated_by}
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
