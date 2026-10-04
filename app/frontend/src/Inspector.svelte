<script lang="ts">
  import Inspector from './Inspector.svelte';
  export let node: any;
</script>
<details>
  <summary>{node.name} — {node.value_w.toFixed(1)} W <small>{node.id}</small></summary>
  <p>{node.method}</p><code>{node.equation}</code>
  {#if node.inputs?.length}
    <table><thead><tr><th>Input</th><th>Value</th><th>Source</th></tr></thead><tbody>
      {#each node.inputs as input}
        <tr><td>{input.name}</td><td>{input.value} {input.unit}</td><td>{input.source.kind}: {input.source.name || ''} {input.source.reference || ''} {input.source.license || ''}
          {#each input.assumptions || [] as assumption}<p>Assumed: {assumption.description}</p>{/each}
        </td></tr>
      {/each}
    </tbody></table>
  {/if}
  {#each node.warnings || [] as warning}<p class="warning">{warning}</p>{/each}
  {#each node.assumptions || [] as assumption}<p class="warning">Assumption: {assumption.description}</p>{/each}
  {#each node.children || [] as child}<Inspector node={child}/>{/each}
</details>
<style>details{border-left:2px solid #aebfcd;padding:10px;margin:10px 0}summary{cursor:pointer;font-weight:600}small{font-weight:400;color:#526777;margin-left:10px}table{border-collapse:collapse;width:100%;font-size:13px;margin:10px 0}th,td{text-align:left;vertical-align:top;padding:8px;border-bottom:1px solid #dce4e8}code{white-space:normal}.warning{color:#795323;font-size:13px}</style>

