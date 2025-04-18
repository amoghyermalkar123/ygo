<script lang="ts">
  import { currentIndex, allEvents } from '../lib/store'
  let events
  $: allEvents.subscribe(e => events = e)
</script>

<div style="display: flex; gap: 0.5rem; margin-bottom: 1rem">
  {#each events as _, i}
    <button on:click={() => currentIndex.set(i)}>{i}</button>
  {/each}
</div>
