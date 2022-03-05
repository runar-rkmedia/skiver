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
    </created>
    {#if entity.updated_at}
      <updated>
        <updated-at>
          Updated:
          {formatDate(entity.updated_at)}
        </updated-at>
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
