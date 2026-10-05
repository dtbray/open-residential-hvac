#!/usr/bin/env python3
# SPDX-License-Identifier: AGPL-3.0-only
"""Native Linux Wails acceptance through AT-SPI; no mocked Go bridge or test hooks.
Run under xvfb-run -a dbus-run-session with GTK_MODULES=gail:atk-bridge.
WebKit sandbox disabling is only needed in nested CI containers, never in releases.
"""
import json, os, pathlib, subprocess, sys, tempfile, time
import pyatspi

binary = sys.argv[1] if len(sys.argv) > 1 else 'bin/hvac-desktop'
os.environ['GTK_MODULES'] = 'gail:atk-bridge'
artifact = pathlib.Path('dist/desktop-acceptance')
artifact.mkdir(parents=True, exist_ok=True)
for stale in ['status.json','steps.json','result.json','saved.yaml']:
    (artifact / stale).unlink(missing_ok=True)
log = open(artifact / 'application.log', 'w')
process = subprocess.Popen([binary], stdout=log, stderr=subprocess.STDOUT)
stages = []

def nodes(root=None):
    if root is None: root = pyatspi.Registry.getDesktop(0)
    yield root
    try:
        for child in root:
            yield from nodes(child)
    except (RuntimeError, TypeError): pass

def wait(predicate, description, timeout=30):
    end = time.monotonic()+timeout
    while time.monotonic() < end:
        if process.poll() is not None: raise RuntimeError('Desktop exited: '+str(process.returncode))
        try:
            result = predicate()
            if result: return result
        except (RuntimeError, NotImplementedError): pass
        time.sleep(.2)
    raise RuntimeError('Timed out: '+description)

def find(role, name):
    return wait(lambda: next((n for n in nodes() if n.getRoleName()==role and n.name==name), None), role+' '+name)

def click(name):
    n=find('button',name)
    if not n.queryAction().doAction(0): raise RuntimeError('Cannot click '+name)

def type_text(node, text):
    node.queryComponent().grabFocus()
    box=node.queryComponent().getExtents(pyatspi.DESKTOP_COORDS)
    subprocess.run(['xdotool','mousemove',str(box.x+max(5,box.width//2)),str(box.y+max(5,box.height//2)),'click','1'],check=True)
    subprocess.run(['xdotool','key','--clearmodifiers','ctrl+a'],check=True)
    subprocess.run(['xdotool','type','--clearmodifiers','--delay','40',text],check=True)

def content():
    values=[]
    for n in nodes():
        if n.name: values.append(n.name)
        try: values.append(n.queryText().getText(0,-1))
        except (NotImplementedError, RuntimeError): pass
    return '\n'.join(values)

def stage(name):
    stages.append(name)
    (artifact / 'steps.json').write_text(json.dumps(stages,indent=2)+'\n')
    print(name,flush=True)

def expand(token):
    disclosure=wait(lambda:next((n for n in nodes() if token in n.name and n.getRoleName() not in ('document web','landmark','section','paragraph')),None),token+' inspector')
    component=disclosure.queryComponent()
    try: component.scrollTo(pyatspi.SCROLL_ANYWHERE)
    except (AttributeError, NotImplementedError, RuntimeError): component.grabFocus()
    time.sleep(.2)
    box=component.getExtents(pyatspi.DESKTOP_COORDS)
    subprocess.run(['xdotool','mousemove',str(box.x+10),str(box.y+10),'click','1'],check=True)

try:
    with tempfile.TemporaryDirectory(prefix='hvac-desktop-') as temp:
        destination=str(pathlib.Path(temp)/'edited.yaml')
        find('heading','Project & design conditions'); stage('native startup')
        click('New')
        wait(lambda:find('entry','Name').queryText().getText(0,-1)=='New project — edit example inputs','new project rendered')
        type_text(find('entry','Name'),'Native acceptance house')
        wait(lambda:find('entry','Name').queryText().getText(0,-1)=='Native acceptance house','name edit complete')
        click('Calculate')
        wait(lambda:'Calculation method: Open Design Load' in content(),'Go-engine result')
        stage('create, edit and calculate')
        # Expand a real result subtree using its accessible disclosure action.
        disclosure=wait(lambda:next((n for n in nodes() if '/heating' in n.name and n.getRoleName() not in ('document web','landmark','section','paragraph')),None),'heating inspector')
        box=disclosure.queryComponent().getExtents(pyatspi.DESKTOP_COORDS)
        subprocess.run(['xdotool','mousemove',str(box.x+10),str(box.y+10),'click','1'],check=True)
        wait(lambda:'sum(children)' in content() or 'Σ' in content(),'aggregation equation')
        stage('inspect result hierarchy and equation')
        for token in ['zone/main/heating','room/living/heating','room/living/heating/envelope','room/living/heating/envelope/living-west']:
            expand(token)
        wait(lambda:'living-west.net_area' in content() and 'wall.u_factor' in content() and 'A_net' in content() and 'user_input' in content(),'leaf inputs, equation and provenance')
        stage('inspect wall inputs, equation and provenance')
        click('Save *')
        wait(lambda:'Save HVAC project' in content(),'native save dialog')
        # GTK filename entry supports an absolute path.
        entries=[n for n in nodes() if n.getRoleName() in ('entry','text') and n.name in ('Name:','Name','')]
        target=next((n for n in entries if 'house.yaml' in n.queryText().getText(0,-1)),None)
        if target is None: raise RuntimeError('Save filename entry missing')
        type_text(target,destination)
        # Native dialog Save is the last Save button in the accessibility tree.
        [n for n in nodes() if n.getRoleName()=='button' and n.name=='Save'][-1].queryAction().doAction(0)
        wait(lambda:pathlib.Path(destination).exists(),'saved project')
        (artifact / 'saved.yaml').write_bytes(pathlib.Path(destination).read_bytes())
        cli=subprocess.run(['bin/hvac','load',destination,'--format','json'],capture_output=True,text=True,check=True)
        result=json.loads(cli.stdout)
        for key in ['heating_load_w','cooling_sensible_w','cooling_latent_w','cooling_load_w']:
            if f'{result[key]:.1f} W' not in content(): raise RuntimeError('CLI/UI total differs: '+key)
        stage('native save and CLI/UI totals at displayed precision')
        # Open the saved file through the actual native dialog.
        click('Project')
        click('Open')
        wait(lambda:'Open HVAC project' in content(),'native open dialog')
        subprocess.run(['xdotool','key','--clearmodifiers','ctrl+l'],check=True)
        time.sleep(.4)
        location=find('filler','Location Layer')
        field=next(n for n in nodes(location) if n.getRoleName() in ('text','entry'))
        type_text(field,destination)
        wait(lambda:field.queryText().getText(0,-1)==destination,'open location entered')
        subprocess.run(['xdotool','key','Return'],check=True)
        time.sleep(.5)
        if 'Open HVAC project' in content():
            [n for n in nodes() if n.getRoleName()=='button' and n.name=='Open'][-1].queryAction().doAction(0)
        wait(lambda:'Open HVAC project' not in content(),'native open completed')
        wait(lambda: next((n for n in nodes() if n.getRoleName()=='entry' and n.name=='Name' and n.queryText().getText(0,-1)=='Native acceptance house'),None),'reopened edited project')
        click('Calculate')
        wait(lambda:'Calculation method: Open Design Load' in content(),'recalculated project')
        stage('native reopen and recalculate')
        (artifact / 'result.json').write_text(cli.stdout)
        (artifact / 'status.json').write_text(json.dumps({'pass':True,'platform':'linux','steps':stages},indent=2)+'\n')
finally:
    if not (artifact / 'status.json').exists():
        (artifact / 'status.json').write_text(json.dumps({'pass':False,'platform':'linux','steps':stages},indent=2)+'\n')
    (artifact / 'content.txt').write_text(content())
    (artifact / 'accessibility.txt').write_text('\n'.join(n.getRoleName()+': '+n.name for n in nodes()))
    process.terminate()
    try: process.wait(timeout=5)
    except subprocess.TimeoutExpired: process.kill(); process.wait()
    log.close()
