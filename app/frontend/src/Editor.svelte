<script lang="ts">
  // Structured input editing only; engineering calculations live in Go.
  import Editor from './Editor.svelte';
  export let value: any;
  export let label = '';
  export let changed: () => void = () => {};
  const title = (s: string) => s.replaceAll('_', ' ');
  const surface = () => ({id:crypto.randomUUID(),name:'New surface',area_m2:0,orientation:'north',tilt_degrees:90,adjacent:'outdoors',assembly:'',evidence:{}});
  const template = () => {
    if (['walls','floors','ceilings'].includes(label)) return surface();
    if (label === 'windows') return {...surface(),parent_surface:'',shgc:0,incident_solar_w_m2:0,shading_factor:1};
    if (label === 'doors') return {...surface(),parent_surface:''};
    if (label === 'layers') return {name:'New layer',r_m2_k_w:0,evidence:{source:{kind:'user_input'}}};
    if (label === 'assemblies') return {id:crypto.randomUUID(),name:'New assembly',u_factor_w_m2_k:0,evidence:{source:{kind:'user_input'}}};
    if (label === 'rooms') return {id:crypto.randomUUID(),name:'New room',floor_area_m2:0,volume_m3:0,occupants:0,occupant_sensible_w_per_person:0,occupant_latent_w_per_person:0,lighting_w:0,internal_sensible_w:0,internal_latent_w:0,walls:[],windows:[],doors:[],floors:[],ceilings:[],evidence:{}};
    if (label === 'zones') return {id:crypto.randomUUID(),name:'New zone',rooms:[]};
    if (label === 'assumptions') return {id:crypto.randomUUID(),description:''};
    return {};
  };
  function add() {
    value = [...value, template()];
    changed();
  }
</script>

{#if Array.isArray(value)}
  <fieldset><legend>{title(label)}</legend>
    {#each value as item, i}
      <details open><summary>{item?.name || item?.id || `Item ${i + 1}`}</summary>
        <Editor bind:value={value[i]} {changed}/>
        <button type="button" onclick={() => {value = value.filter((_: any, n: number) => n !== i); changed();}}>Remove</button>
      </details>
    {/each}
    <button type="button" onclick={add}>Add {title(label)}</button>
  </fieldset>
{:else if value !== null && typeof value === 'object'}
  <fieldset><legend>{title(label)}</legend>
    {#each Object.keys(value) as key}
      <Editor bind:value={value[key]} label={key} {changed}/>
    {/each}
  </fieldset>
{:else if typeof value === 'number'}
  <label>{title(label)} <input type="number" step="any" bind:value oninput={changed}/></label>
{:else if typeof value === 'boolean'}
  <label>{title(label)} <input type="checkbox" bind:checked={value} onchange={changed}/></label>
{:else}
  <label>{title(label)} <input type="text" bind:value oninput={changed}/></label>
{/if}

<style>
  fieldset{border:1px solid #cbd5df;border-radius:6px;margin:12px 0;padding:12px;min-width:0}legend{font-weight:600}label{display:flex;justify-content:space-between;gap:12px;margin:8px 0;align-items:center;font-size:14px}input{max-width:65%;width:320px;border:1px solid #a5b3c2;padding:7px;border-radius:4px}details{margin:10px 0;padding:8px;border-left:3px solid #b6c8d8}summary{cursor:pointer}button{padding:6px 12px;cursor:pointer}
</style>
