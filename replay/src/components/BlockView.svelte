<script lang="ts">
  import { currentEvent } from '../lib/store'
  let event
  $: currentEvent.subscribe(e => event = e)
</script>

<h3>🧱 Blocks By Client</h3>
{#each Object.entries(event?.blocks ?? {}) as [client, blocks]}
  <div>
    <strong>Client {client}</strong>
    <div style="display: flex; flex-wrap: wrap; gap: 0.25rem; margin-bottom: 0.5rem">
      {#each blocks as blk}
        <div style="border: 1px solid gray; padding: 4px; background: {blk.deleted ? '#f88' : '#8f8'}">
          <div style="font-size: 0.7rem">[{blk.id.client}:{blk.id.clock}]</div>
          <div>{blk.content || '(deleted)'}</div>
        </div>
      {/each}
    </div>
  </div>
{/each}
