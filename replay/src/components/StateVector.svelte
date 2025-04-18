<script lang="ts">
  import { currentEvent } from '../lib/store'
  let event
  $: currentEvent.subscribe(e => event = e)
</script>

<h3>🧠 State Vector</h3>
<pre>{JSON.stringify(event?.state_vector, null, 2)}</pre>
