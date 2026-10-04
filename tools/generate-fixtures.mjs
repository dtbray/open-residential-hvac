// SPDX-License-Identifier: AGPL-3.0-only
// Regenerates deliberately simple SI fixtures. Values are synthetic user inputs,
// not engineering recommendations or observations of real buildings.
import fs from 'node:fs';
import path from 'node:path';
import {fileURLToPath} from 'node:url';
const root=path.dirname(path.dirname(fileURLToPath(import.meta.url)));
const evidence={source:{kind:'user_input',reference:'synthetic validation fixture'}};
const surface=(id,area,assembly='wall',adjacent='outdoors')=>({id,name:id,area_m2:area,orientation:'west',tilt_degrees:90,adjacent,assembly,evidence:{area_m2:evidence}});
const room=(id='living')=>({id,name:'Living Room',floor_area_m2:30,volume_m3:75,occupants:0,occupant_sensible_w_per_person:0,occupant_latent_w_per_person:0,lighting_w:0,internal_sensible_w:0,internal_latent_w:0,walls:[surface(id+'-west',20)],windows:[],doors:[],floors:[],ceilings:[],evidence:{volume_m3:evidence}});
const base=()=>({version:1,building:{id:'house',name:'Simple Box',location:{description:'Synthetic fixture, not a geographic design recommendation'},design:{heating:{outdoor_db_c:-15,indoor_db_c:20},cooling:{outdoor_db_c:35,outdoor_rh_fraction:0.5,indoor_db_c:24,indoor_rh_fraction:0.5},pressure_pa:101325,evidence:{}},assemblies:[{id:'wall',name:'Explicit wall U',u_factor_w_m2_k:0.4,evidence},{id:'glass',name:'Explicit glazing U',u_factor_w_m2_k:2,evidence}],infiltration:{mode:'explicit',airflow_m3_s:0,evidence:{airflow_m3_s:evidence}},ventilation:{airflow_m3_s:0,evidence:{airflow_m3_s:evidence}},zones:[{id:'main',name:'Main',rooms:[room()]}]}});
const fixtures={};
fixtures['wall-conduction-only']=base();
fixtures['simple-box']=base();
fixtures['simple-box-hot-climate']=base();fixtures['simple-box-hot-climate'].building.design.cooling.outdoor_db_c=42;
fixtures['simple-box-cold-climate']=base();fixtures['simple-box-cold-climate'].building.design.heating.outdoor_db_c=-30;
fixtures['tight-house']=base();fixtures['tight-house'].building.infiltration={mode:'ach_natural',ach:0.1,evidence:{ach:evidence}};
fixtures['leaky-house']=base();fixtures['leaky-house'].building.infiltration={mode:'ach_natural',ach:1,evidence:{ach:evidence}};
fixtures['high-latent-climate']=base();fixtures['high-latent-climate'].building.design.cooling.outdoor_rh_fraction=0.8;fixtures['high-latent-climate'].building.infiltration={mode:'explicit',airflow_m3_s:0.05,evidence:{airflow_m3_s:evidence}};
fixtures['high-window-area']=base();fixtures['high-window-area'].building.zones[0].rooms[0].windows=[{...surface('living-glass',15,'glass'),parent_surface:'living-west',shgc:0.5,incident_solar_w_m2:400,shading_factor:0.8,evidence:{area_m2:evidence,shgc:evidence,incident_solar_w_m2:evidence,shading_factor:evidence}}];
fixtures['multi-room-ranch']=base();const ranch=fixtures['multi-room-ranch'].building;ranch.name='Example Ranch';ranch.zones[0].rooms.push({...room('bedroom'),name:'Bedroom',floor_area_m2:15,volume_m3:37.5});ranch.infiltration={mode:'ach_natural',ach:0.3,evidence:{ach:evidence}};ranch.ventilation.airflow_m3_s=0.015;
for(const r of ranch.zones[0].rooms){r.occupants=1;r.occupant_sensible_w_per_person=70;r.occupant_latent_w_per_person=45;r.lighting_w=50;r.internal_sensible_w=80;r.internal_latent_w=10;r.ceilings=[{...surface(r.id+'-ceiling',r.floor_area_m2),tilt_degrees:0}];r.floors=[{...surface(r.id+'-floor',r.floor_area_m2,'wall','conditioned'),tilt_degrees:180}];r.windows=[{...surface(r.id+'-window',3,'glass'),parent_surface:r.id+'-west',shgc:0.4,incident_solar_w_m2:350,shading_factor:0.8,evidence:{area_m2:evidence,shgc:evidence,incident_solar_w_m2:evidence,shading_factor:evidence}}];r.doors=[{...surface(r.id+'-door',2),parent_surface:r.id+'-west'}];}
fixtures['two-story']=structuredClone(fixtures['multi-room-ranch']);const upper=fixtures['two-story'].building.zones[0].rooms.pop();fixtures['two-story'].building.zones.push({id:'upper',name:'Upper Zone',rooms:[upper]});
fixtures['attic-duct-house']=structuredClone(fixtures['multi-room-ranch']);fixtures['attic-duct-house'].building.name='Attic boundary fixture (duct losses unsupported)';for(const r of fixtures['attic-duct-house'].building.zones[0].rooms){r.ceilings[0].adjacent='unconditioned_attic';r.ceilings[0].heating_adjacent_db_c=0;r.ceilings[0].cooling_adjacent_db_c=45;}
fs.mkdirSync(path.join(root,'testdata/buildings'),{recursive:true});
for(const [name,p]of Object.entries(fixtures))fs.writeFileSync(path.join(root,'testdata/buildings',name+'.json'),JSON.stringify(p,null,2)+'\n');
fs.mkdirSync(path.join(root,'app/frontend/src'),{recursive:true});fs.writeFileSync(path.join(root,'app/frontend/src/example.json'),JSON.stringify(fixtures['multi-room-ranch'],null,2)+'\n');
console.log('Generated',Object.keys(fixtures).length,'synthetic fixtures.');

