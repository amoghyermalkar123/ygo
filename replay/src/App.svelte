<script lang="ts">
  import Timeline from './components/Timeline.svelte'
  import ContentView from './components/ContentView.svelte'
  import StateVector from './components/StateVector.svelte'
  import BlockView from './components/BlockView.svelte'
</script>

<main style="padding: 2rem; font-family: sans-serif">
  <h1>📜 CRDT Replay Viewer</h1>
  <Timeline />
  <ContentView />
  <StateVector />
  <BlockView />
</main>
