package server
import "net/http"
func(s *Server)dashboard(w http.ResponseWriter,r *http.Request){w.Header().Set("Content-Type","text/html");w.Write([]byte(dashHTML))}
const dashHTML=`<!DOCTYPE html><html><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>Cipher</title>
<style>:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#e8753a;--leather:#a0845c;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--mono:'JetBrains Mono',monospace}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--mono);line-height:1.5}
.hdr{padding:1rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}.hdr h1{font-size:.9rem;letter-spacing:2px}
.main{padding:1.5rem;max-width:900px;margin:0 auto}
.search{width:100%;padding:.5rem .8rem;background:var(--bg2);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.78rem;margin-bottom:1rem}
.cat-bar{display:flex;gap:.3rem;margin-bottom:1rem;flex-wrap:wrap}
.cat-btn{font-size:.6rem;padding:.2rem .5rem;border:1px solid var(--bg3);background:var(--bg);color:var(--cm);cursor:pointer}.cat-btn:hover{border-color:var(--leather)}.cat-btn.active{border-color:var(--rust);color:var(--rust)}
.entry{background:var(--bg2);border:1px solid var(--bg3);padding:.8rem 1rem;margin-bottom:.5rem}
.entry-name{font-size:.85rem;color:var(--cream);margin-bottom:.2rem}
.entry-url{font-size:.65rem;color:var(--rust)}
.entry-user{font-size:.7rem;color:var(--cd);margin-top:.3rem}
.entry-pass{font-family:var(--mono);font-size:.7rem;color:var(--cm);background:var(--bg);padding:.2rem .4rem;border:1px solid var(--bg3);cursor:pointer;display:inline-block;margin-top:.2rem}
.entry-pass:hover{color:var(--cream)}
.entry-meta{font-size:.6rem;color:var(--cm);margin-top:.3rem}
.btn{font-size:.6rem;padding:.25rem .6rem;cursor:pointer;border:1px solid var(--bg3);background:var(--bg);color:var(--cd)}.btn:hover{border-color:var(--leather);color:var(--cream)}
.btn-p{background:var(--rust);border-color:var(--rust);color:var(--bg)}
.modal-bg{display:none;position:fixed;inset:0;background:rgba(0,0,0,.6);z-index:100;align-items:center;justify-content:center}.modal-bg.open{display:flex}
.modal{background:var(--bg2);border:1px solid var(--bg3);padding:1.5rem;width:400px;max-width:90vw}
.modal h2{font-size:.8rem;margin-bottom:1rem;color:var(--rust)}
.fr{margin-bottom:.5rem}.fr label{display:block;font-size:.55rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-bottom:.15rem}
.fr input,.fr select,.fr textarea{width:100%;padding:.35rem .5rem;background:var(--bg);border:1px solid var(--bg3);color:var(--cream);font-family:var(--mono);font-size:.7rem}
.acts{display:flex;gap:.4rem;justify-content:flex-end;margin-top:.8rem}
.empty{text-align:center;padding:3rem;color:var(--cm);font-style:italic;font-size:.75rem}
</style></head><body>
<div class="hdr"><h1>CIPHER</h1><button class="btn btn-p" onclick="openForm()">+ Add Entry</button></div>
<div class="main">
<input class="search" id="search" placeholder="Search passwords..." oninput="render()">
<div class="cat-bar" id="cats"></div>
<div id="entries"></div>
</div>
<div class="modal-bg" id="mbg" onclick="if(event.target===this)cm()"><div class="modal" id="mdl"></div></div>
<script>
const A='/api';let secrets=[],filterCat='';
async function load(){const r=await fetch(A+'/secrets').then(r=>r.json());secrets=r.secrets||[];
const cats=[...new Set(secrets.map(s=>s.category).filter(c=>c))];
let h='<button class="cat-btn'+(filterCat===''?' active':'')+'" onclick="setCat(\'\')">All ('+secrets.length+')</button>';
cats.forEach(c=>{h+='<button class="cat-btn'+(filterCat===c?' active':'')+'" onclick="setCat(\''+c+'\')">'+esc(c)+'</button>';});
document.getElementById('cats').innerHTML=h;render();}
function setCat(c){filterCat=c;load();}
function render(){const q=(document.getElementById('search').value||'').toLowerCase();
let filtered=secrets.filter(s=>{if(filterCat&&s.category!==filterCat)return false;if(q&&!(s.name+s.username+s.url+s.category).toLowerCase().includes(q))return false;return true;});
if(!filtered.length){document.getElementById('entries').innerHTML='<div class="empty">No passwords stored.</div>';return;}
let h='';filtered.forEach(s=>{
h+='<div class="entry"><div style="display:flex;justify-content:space-between"><div class="entry-name">'+esc(s.name)+'</div><button class="btn" onclick="del(\''+s.id+'\')" style="font-size:.5rem;color:var(--cm)">✕</button></div>';
if(s.url)h+='<div class="entry-url"><a href="'+esc(s.url)+'" target="_blank">'+esc(s.url)+'</a></div>';
if(s.username)h+='<div class="entry-user">User: <span style="cursor:pointer" onclick="navigator.clipboard.writeText(\''+esc(s.username)+'\')">'+esc(s.username)+' 📋</span></div>';
if(s.password){const masked='•'.repeat(Math.min(s.password.length,16));h+='<span class="entry-pass" onclick="reveal(this,\''+esc(s.password)+'\')">'+masked+'</span>';}
if(s.category)h+='<div class="entry-meta">Category: '+esc(s.category)+'</div>';
h+='</div>';});
document.getElementById('entries').innerHTML=h;}
function reveal(el,val){if(el.dataset.r){el.textContent='•'.repeat(val.length);el.dataset.r=''}else{el.textContent=val;el.dataset.r='1'}}
async function del(id){if(confirm('Delete?')){await fetch(A+'/secrets/'+id,{method:'DELETE'});load();}}
function openForm(){document.getElementById('mdl').innerHTML='<h2>Add Password</h2><div class="fr"><label>Name</label><input id="f-n" placeholder="e.g. GitHub"></div><div class="fr"><label>Username</label><input id="f-u"></div><div class="fr"><label>Password</label><input id="f-p" type="password"></div><div class="fr"><label>URL</label><input id="f-url" placeholder="https://"></div><div class="fr"><label>Category</label><input id="f-c" placeholder="e.g. work, personal, social"></div><div class="fr"><label>Notes</label><textarea id="f-nt" rows="2"></textarea></div><div class="acts"><button class="btn" onclick="cm()">Cancel</button><button class="btn btn-p" onclick="sub()">Save</button></div>';document.getElementById('mbg').classList.add('open');}
async function sub(){await fetch(A+'/secrets',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({name:document.getElementById('f-n').value,username:document.getElementById('f-u').value,password:document.getElementById('f-p').value,url:document.getElementById('f-url').value,category:document.getElementById('f-c').value,notes:document.getElementById('f-nt').value})});cm();load();}
function cm(){document.getElementById('mbg').classList.remove('open');}
function esc(s){if(!s)return'';const d=document.createElement('div');d.textContent=s;return d.innerHTML;}
load();
</script></body></html>`
