const qs=(s,el=document)=>el.querySelector(s), qsa=(s,el=document)=>[...el.querySelectorAll(s)];
const cart=[]; const drawer=qs('.cart-drawer'), overlay=qs('.overlay'), toast=qs('.toast');
const money=n=>`₹${n.toLocaleString('en-IN')}`;
function notify(msg){toast.textContent=msg;toast.classList.add('show');setTimeout(()=>toast.classList.remove('show'),1800)}
function openCart(){drawer.classList.add('open');overlay.classList.add('show');drawer.setAttribute('aria-hidden','false')}
function closeCart(){drawer.classList.remove('open');overlay.classList.remove('show');drawer.setAttribute('aria-hidden','true')}
function renderCart(){qs('.cart-count').textContent=cart.length;const area=qs('.cart-items');if(!cart.length){area.innerHTML='<p>Your bag is waiting for something beautiful.</p>'}else area.innerHTML=cart.map(x=>`<div class="cart-item"><img src="${x.img}" alt=""><div><h4>${x.name}</h4><small>Size M · Qty 1</small><p>${money(x.price)}</p></div></div>`).join('');qs('.cart-total').textContent=money(cart.reduce((a,b)=>a+b.price,0))}
qsa('.quick-add').forEach(btn=>btn.addEventListener('click',()=>{const card=btn.closest('.product-card');cart.push({name:qs('h3',card).textContent,price:+qs('strong',card).textContent.replace(/\D/g,''),img:qs('img',card).src});renderCart();notify('Added to your bag');openCart()}));
qs('.bag-btn').addEventListener('click',openCart);qs('.cart-close').addEventListener('click',closeCart);overlay.addEventListener('click',closeCart);
qsa('.heart').forEach(b=>b.addEventListener('click',()=>{b.classList.toggle('liked');b.textContent=b.classList.contains('liked')?'♥':'♡';notify(b.classList.contains('liked')?'Saved to your wishlist':'Removed from wishlist')}));
qsa('.filters button').forEach(btn=>btn.addEventListener('click',()=>{qsa('.filters button').forEach(b=>b.classList.remove('active'));btn.classList.add('active');qsa('.product-card').forEach(c=>c.style.display=btn.dataset.filter==='all'||c.dataset.category===btn.dataset.filter?'block':'none')}));
qs('.menu-btn').addEventListener('click',()=>qs('.nav-links').classList.toggle('open'));qsa('.nav-links a').forEach(a=>a.addEventListener('click',()=>qs('.nav-links').classList.remove('open')));
qs('#newsletter').addEventListener('submit',e=>{e.preventDefault();notify('Welcome to the Rangavali circle ✦');e.target.reset()});
const io=new IntersectionObserver(es=>es.forEach(e=>{if(e.isIntersecting)e.target.classList.add('visible')}),{threshold:.12});qsa('.reveal').forEach(el=>io.observe(el));
