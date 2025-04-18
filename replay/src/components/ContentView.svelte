<script lang="ts">
  import { currentEvent } from '../lib/store'
  let event
  $: currentEvent.subscribe(e => event = e)

  function getContent(): string {
    return Object.values(event.blocks)
      .flat()
      .filter(b => !b.deleted)
      .map(b => b.content)
      .join('')
  }
</script>

<h3>📄 Content View</h3>
<pre>{getContent()}</pre>
