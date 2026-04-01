package server

import "net/http"

const uiHTML = `<!DOCTYPE html><html lang="en"><head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Cipher — Stockyard</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=Libre+Baskerville:ital,wght@0,400;0,700;1,400&family=JetBrains+Mono:wght@400;600&display=swap" rel="stylesheet">
<style>:root{
  --bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;
  --rust:#c45d2c;--rust-light:#e8753a;--rust-dark:#8b3d1a;
  --leather:#a0845c;--leather-light:#c4a87a;
  --cream:#f0e6d3;--cream-dim:#bfb5a3;--cream-muted:#7a7060;
  --gold:#d4a843;--green:#5ba86e;--red:#c0392b;
  --font-serif:'Libre Baskerville',Georgia,serif;
  --font-mono:'JetBrains Mono',monospace;
}
*{margin:0;padding:0;box-sizing:border-box}
body{background:var(--bg);color:var(--cream);font-family:var(--font-serif);min-height:100vh}
a{color:var(--rust-light);text-decoration:none}a:hover{color:var(--gold)}
.hdr{background:var(--bg2);border-bottom:2px solid var(--rust-dark);padding:.9rem 1.8rem;display:flex;align-items:center;justify-content:space-between}
.hdr-left{display:flex;align-items:center;gap:1rem}
.hdr-brand{font-family:var(--font-mono);font-size:.75rem;color:var(--leather);letter-spacing:3px;text-transform:uppercase}
.hdr-title{font-family:var(--font-mono);font-size:1.1rem;color:var(--cream);letter-spacing:1px}
.badge{font-family:var(--font-mono);font-size:.6rem;padding:.2rem .6rem;letter-spacing:1px;text-transform:uppercase;border:1px solid;color:var(--green);border-color:var(--green)}
.main{max-width:1000px;margin:0 auto;padding:2rem 1.5rem}
.cards{display:grid;grid-template-columns:repeat(auto-fit,minmax(130px,1fr));gap:1rem;margin-bottom:2rem}
.card{background:var(--bg2);border:1px solid var(--bg3);padding:1rem 1.2rem}
.card-val{font-family:var(--font-mono);font-size:1.6rem;font-weight:700;color:var(--cream);display:block}
.card-lbl{font-family:var(--font-mono);font-size:.58rem;letter-spacing:2px;text-transform:uppercase;color:var(--leather);margin-top:.2rem}
.section{margin-bottom:2rem}
.section-title{font-family:var(--font-mono);font-size:.68rem;letter-spacing:3px;text-transform:uppercase;color:var(--rust-light);margin-bottom:.8rem;padding-bottom:.5rem;border-bottom:1px solid var(--bg3)}
table{width:100%;border-collapse:collapse;font-family:var(--font-mono);font-size:.75rem}
th{background:var(--bg3);padding:.4rem .8rem;text-align:left;color:var(--leather-light);font-weight:400;letter-spacing:1px;font-size:.62rem;text-transform:uppercase}
td{padding:.4rem .8rem;border-bottom:1px solid var(--bg3);color:var(--cream-dim);word-break:break-all}
tr:hover td{background:var(--bg2)}
.empty{color:var(--cream-muted);text-align:center;padding:2rem;font-style:italic}
.btn{font-family:var(--font-mono);font-size:.7rem;padding:.3rem .8rem;border:1px solid var(--leather);background:transparent;color:var(--cream);cursor:pointer;transition:all .2s}
.btn:hover{border-color:var(--rust-light);color:var(--rust-light)}
.btn-rust{border-color:var(--rust);color:var(--rust-light)}.btn-rust:hover{background:var(--rust);color:var(--cream)}
.btn-sm{font-size:.62rem;padding:.2rem .5rem}
.lbl{font-family:var(--font-mono);font-size:.62rem;letter-spacing:1px;text-transform:uppercase;color:var(--leather)}
input{font-family:var(--font-mono);font-size:.78rem;background:var(--bg3);border:1px solid var(--bg3);color:var(--cream);padding:.4rem .7rem;outline:none}
input:focus{border-color:var(--leather)}
.row{display:flex;gap:.8rem;align-items:flex-end;flex-wrap:wrap;margin-bottom:1rem}
.field{display:flex;flex-direction:column;gap:.3rem}
.tabs{display:flex;gap:0;margin-bottom:1.5rem;border-bottom:1px solid var(--bg3)}
.tab{font-family:var(--font-mono);font-size:.72rem;padding:.6rem 1.2rem;color:var(--cream-muted);cursor:pointer;border-bottom:2px solid transparent;letter-spacing:1px;text-transform:uppercase}
.tab:hover{color:var(--cream-dim)}.tab.active{color:var(--rust-light);border-bottom-color:var(--rust-light)}
.tab-content{display:none}.tab-content.active{display:block}
pre{background:var(--bg3);padding:.8rem 1rem;font-family:var(--font-mono);font-size:.72rem;color:var(--cream-dim);overflow-x:auto}
</style></head><body>
<div class="hdr">
  <div class="hdr-left">
    <svg viewBox="0 0 64 64" width="22" height="22" fill="none"><rect x="8" y="8" width="8" height="48" rx="2.5" fill="#e8753a"/><rect x="28" y="8" width="8" height="48" rx="2.5" fill="#e8753a"/><rect x="48" y="8" width="8" height="48" rx="2.5" fill="#e8753a"/><rect x="8" y="27" width="48" height="7" rx="2.5" fill="#c4a87a"/></svg>
    <span class="hdr-brand">Stockyard</span>
    <span class="hdr-title">Cipher</span>
  </div>
  <div style="display:flex;gap:.8rem;align-items:center"><span class="badge">Free</span><a href="/api/status" class="lbl" style="color:var(--leather)">API</a></div>
</div>
<div class="main">
<div class="cards">
  <div class="card"><span class="card-val" id="s-projects">—</span><span class="card-lbl">Projects</span></div>
  <div class="card"><span class="card-val" id="s-secrets">—</span><span class="card-lbl">Secrets</span></div>
  <div class="card"><span class="card-val" id="s-tokens">—</span><span class="card-lbl">Tokens</span></div>
  <div class="card"><span class="card-val" id="s-accesses">—</span><span class="card-lbl">Accesses</span></div>
</div>
<div class="tabs">
  <div class="tab active" onclick="switchTab('projects')">Projects</div>
  <div class="tab" onclick="switchTab('usage')">Usage</div>
</div>
<div id="tab-projects" class="tab-content active">
  <div class="section">
    <div class="section-title">Create Project</div>
    <div class="row">
      <div class="field"><span class="lbl">Name</span><input id="c-name" placeholder="my-api-prod" style="width:200px"></div>
      <button class="btn btn-rust" onclick="createProject()">Create</button>
    </div>
    <div id="c-result"></div>
  </div>
  <div class="section">
    <div class="section-title">Projects</div>
    <table><thead><tr><th>Name</th><th>ID</th><th>Secrets</th><th>Created</th><th></th></tr></thead>
    <tbody id="projects-body"></tbody></table>
  </div>
</div>
<div id="tab-usage" class="tab-content">
  <div class="section">
    <div class="section-title">Quick Start</div>
    <pre>
# Create a project
curl -X POST http://localhost:8870/api/projects \
  -H "Content-Type: application/json" \
  -d '{"name":"my-api-prod"}'

# Store a secret
curl -X PUT http://localhost:8870/api/projects/{id}/secrets/DATABASE_URL \
  -H "Content-Type: application/json" \
  -d '{"value":"postgres://user:pass@db:5432/myapp"}'

# Create a short-lived token (for CI/CD)
curl -X POST http://localhost:8870/api/projects/{id}/tokens \
  -H "Content-Type: application/json" \
  -d '{"name":"github-actions","ttl_minutes":15}'

# Fetch all secrets with token (in your CI pipeline)
curl -H "Authorization: Bearer {token}" \
  http://localhost:8870/api/secrets

# Or get them as .env format
curl -H "Authorization: Bearer {token}" \
  "http://localhost:8870/api/secrets?format=env"
    </pre>
  </div>
</div>
</div>
<script>
function switchTab(n){
  document.querySelectorAll('.tab').forEach(t=>t.classList.toggle('active',t.textContent.toLowerCase()===n));
  document.querySelectorAll('.tab-content').forEach(t=>t.classList.toggle('active',t.id==='tab-'+n));
}
async function refresh(){
  try{
    const s=await(await fetch('/api/status')).json();
    document.getElementById('s-projects').textContent=s.projects||0;
    document.getElementById('s-secrets').textContent=s.secrets||0;
    document.getElementById('s-tokens').textContent=s.tokens||0;
    document.getElementById('s-accesses').textContent=fmt(s.accesses||0);
  }catch(e){}
  try{
    const d=await(await fetch('/api/projects')).json();
    const ps=d.projects||[];
    const tb=document.getElementById('projects-body');
    if(!ps.length){tb.innerHTML='<tr><td colspan="5" class="empty">No projects yet.</td></tr>';return;}
    tb.innerHTML=ps.map(p=>
      '<tr><td style="color:var(--cream);font-weight:600">'+esc(p.name)+'</td>'+
      '<td style="font-size:.62rem">'+p.id+'</td>'+
      '<td>'+p.secret_count+'</td>'+
      '<td style="font-size:.65rem;color:var(--cream-muted)">'+timeAgo(p.created_at)+'</td>'+
      '<td><button class="btn btn-sm" onclick="deleteProject(\''+p.id+'\')">Delete</button></td></tr>'
    ).join('');
  }catch(e){}
}
async function createProject(){
  const name=document.getElementById('c-name').value.trim();
  if(!name)return;
  const r=await fetch('/api/projects',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({name})});
  const d=await r.json();
  if(r.ok){document.getElementById('c-result').innerHTML='<span style="color:var(--green)">Created: '+d.project.id+'</span>';document.getElementById('c-name').value='';refresh();}
  else{document.getElementById('c-result').innerHTML='<span style="color:var(--red)">'+esc(d.error)+'</span>';}
}
async function deleteProject(id){
  if(!confirm('Delete project and all secrets?'))return;
  await fetch('/api/projects/'+id,{method:'DELETE'});
  refresh();
}
function fmt(n){if(n>=1e6)return(n/1e6).toFixed(1)+'M';if(n>=1e3)return(n/1e3).toFixed(1)+'K';return n;}
function esc(s){const d=document.createElement('div');d.textContent=s||'';return d.innerHTML;}
function timeAgo(s){if(!s)return'—';const d=new Date(s);const diff=Date.now()-d.getTime();if(diff<60000)return'now';if(diff<3600000)return Math.floor(diff/60000)+'m';if(diff<86400000)return Math.floor(diff/3600000)+'h';return Math.floor(diff/86400000)+'d';}
refresh();setInterval(refresh,8000);
</script></body></html>`

func (s *Server) handleUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(uiHTML))
}
