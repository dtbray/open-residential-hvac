<script lang="ts">
  import example from './example.json';
  import Editor from './Editor.svelte';
  import Inspector from './Inspector.svelte';
  let project: any = structuredClone(example);
  let tab = 'Project';
  let result: any = null;
  let error = '';
  let busy = false;
  let raw = JSON.stringify(project, null, 2);
  let dirty = false;
  const tabs = ['Project', 'Rooms & envelope', 'Assemblies', 'Airflow', 'Results', 'Project JSON'];
  const bridge = () => (window as any).go?.main?.Desktop;
  function changed() {result = null;dirty = true;queueMicrotask(() => {raw = JSON.stringify(project, null, 2);});}
  async function run(task: () => Promise<void>) {busy = true;error = '';try {if (!bridge()) throw new Error('Open this editor in the Wails desktop application to use the Go engine.');await task();} catch(e) {error = String(e);} finally {busy = false;}}
  async function open() {if (dirty && !window.confirm('Discard unsaved edits and open another project?')) return;await run(async () => {const text = await bridge().OpenProject();if (text) {project = JSON.parse(text);changed();dirty = false;}});}
  async function save() {await run(async () => {if (await bridge().SaveProject(JSON.stringify(project))) dirty = false;});}
  async function calculate() {await run(async () => {result = await bridge().Calculate(JSON.stringify(project));tab = 'Results';});}
  function applyRaw() {try {project = JSON.parse(raw);changed();error = '';} catch(e) {error = String(e);}}
  function fresh() {if (dirty && !window.confirm('Discard unsaved edits?')) return;project = structuredClone(example);project.building.name = 'New project — edit example inputs';changed();tab = 'Project';}
</script>

<svelte:head><title>Open Residential HVAC</title></svelte:head>
<header><div><h1>Open Residential HVAC</h1><p>Transparent residential design loads · AGPL-3.0 · Development preview</p></div><div class="actions"><button disabled={busy} onclick={fresh}>New</button><button disabled={busy} onclick={open}>Open</button><button disabled={busy} onclick={save}>Save{dirty ? ' *' : ''}</button><button class="primary" disabled={busy} onclick={calculate}>{busy ? 'Working…' : 'Calculate'}</button></div></header>
<nav>{#each tabs as name}<button class:active={tab === name} onclick={() => {tab = name;}}>{name}</button>{/each}</nav>
<main>
  {#if error}<div role="alert" class="error"><pre>{error}</pre></div>{/if}
  {#if tab === 'Project'}
    <h2>Project & design conditions</h2><p>Fields use the SI units shown in their names. Relative humidity is a fraction from 0 to 1. The example is synthetic; enter project-specific values.</p>
    <label>Name <input bind:value={project.building.name} oninput={changed}/></label>
    <Editor bind:value={project.building.location} label="Location" {changed}/><Editor bind:value={project.building.design} label="Design" {changed}/>
  {:else if tab === 'Rooms & envelope'}
    <h2>Rooms & envelope</h2><p>Surfaces use gross area; referenced openings are subtracted by the engine. IDs must be unique. Use Project JSON to add fields absent from the current object.</p>
    <Editor bind:value={project.building.zones} label="Zones" {changed}/>
  {:else if tab === 'Assemblies'}
    <h2>Reusable assemblies</h2><p>Specify either a U-factor or layers with explicit R-values. Include films and framing effects yourself; the engine adds none.</p><Editor bind:value={project.building.assemblies} label="Assemblies" {changed}/>
  {:else if tab === 'Airflow'}
    <h2>Infiltration & ventilation</h2><Editor bind:value={project.building.infiltration} label="Infiltration" {changed}/><Editor bind:value={project.building.ventilation} label="Ventilation" {changed}/>
  {:else if tab === 'Project JSON'}
    <h2>Complete project input</h2><p>Use this editor for optional fields, new surfaces, layers, or provenance. Apply before calculating or saving.</p><textarea bind:value={raw} aria-label="Project JSON"></textarea><button onclick={applyRaw}>Apply JSON</button>
  {:else if result}
    <h2>Calculated loads</h2><div class="totals"><p>Heating <strong>{result.heating_load_w.toFixed(1)} W</strong></p><p>Cooling sensible <strong>{result.cooling_sensible_w.toFixed(1)} W</strong></p><p>Cooling latent <strong>{result.cooling_latent_w.toFixed(1)} W</strong></p><p>Cooling total <strong>{result.cooling_load_w.toFixed(1)} W</strong></p></div>
    <table><thead><tr><th>Room</th><th>Heating W</th><th>Cooling sensible W</th><th>Cooling latent W</th></tr></thead><tbody>{#each result.rooms as room}<tr><td>{room.name}</td><td>{room.heating_load_w.toFixed(1)}</td><td>{room.cooling_sensible_w.toFixed(1)}</td><td>{room.cooling_latent_w.toFixed(1)}</td></tr>{/each}</tbody></table>
    <h2>Calculation inspector</h2><Inspector node={result.heating}/><Inspector node={result.cooling}/>
  {:else}<h2>Results</h2><p>Calculate the current project to inspect its loads.</p>{/if}
</main><footer>Steady-state open-physics model. No ACCA compliance claim. Instantaneous solar gains do not model thermal storage.</footer>
<style>:global(body){font-family:system-ui,sans-serif;color:#233849;background:#f6f8fa;margin:0}header{background:#fff;padding:20px 28px;border-bottom:1px solid #d4dfe7;display:flex;align-items:center;justify-content:space-between;gap:20px}h1{margin:0;font-size:24px}header p{font-size:13px;color:#647888;margin-bottom:0}.actions{display:flex;gap:8px}button{border:1px solid #a9bccb;background:#fff;color:#244358;padding:9px 14px;border-radius:5px;cursor:pointer}button:disabled{opacity:.5;cursor:wait}.primary,.active{background:#235b78;color:white}nav{display:flex;gap:8px;padding:16px 28px;flex-wrap:wrap}main{margin:0 auto;max-width:1060px;padding:12px 28px 40px}h2{font-size:21px}p{line-height:1.5}label{display:flex;gap:15px;align-items:center}input{padding:8px;min-width:300px}textarea{box-sizing:border-box;width:100%;height:60vh;font-family:monospace;padding:15px}.totals{display:flex;gap:24px;flex-wrap:wrap}.totals strong{display:block;font-size:24px}.error{border:1px solid #b65945;background:#fff0e9;padding:12px}pre{white-space:pre-wrap}table{width:100%;border-collapse:collapse}th,td{text-align:left;padding:10px;border-bottom:1px solid #d5dfe6}footer{padding:15px 28px;background:#e9eef2;font-size:12px}</style>
