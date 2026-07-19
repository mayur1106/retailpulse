'use client';
import Link from 'next/link';
import {useEffect,useMemo,useState} from 'react';
import {Check,ChevronDown,ChevronRight,Heart,Minus,Plus,ShoppingBag,Star,Truck} from 'lucide-react';
import {StorefrontConfig,fallbackConfig,Product,fetchProduct,fetchProducts,fetchStorefrontConfig,money} from '../lib/catalog';
import {ProductCard} from './product-card';
import {useStore} from './store-provider';

const API_BASE=process.env.NEXT_PUBLIC_API_BASE_URL??'http://localhost:4005';
const STORE_SLUG=process.env.NEXT_PUBLIC_STORE_SLUG??'rangavali';

export function CategoryPage({name}:{name:string}){const [sort,setSort]=useState('Featured'),[limit,setLimit]=useState(18),[items,setItems]=useState<Product[]>([]);const category=name==='all styles'?'all':name;useEffect(()=>{fetchProducts(category,100).then(setItems)},[category]);const shown=useMemo(()=>{let x=[...items];if(sort==='Price: Low to high')x.sort((a,b)=>a.price-b.price);if(sort==='Top rated')x.sort((a,b)=>b.rating-a.rating);return x.slice(0,limit)},[sort,limit,items]);return <main><div className="bg-blush py-8 md:py-10 text-center"><p className="eyebrow">The Rangavali edit</p><h1 className="title mt-2 capitalize">{name.replaceAll('-',' ')}</h1><p className="text-black/50 mt-2">Modern silhouettes, rooted in craft and made for your every day.</p></div><div className="container py-6"><div className="flex justify-between items-center pb-4 border-b"><button className="flex items-center gap-2 font-bold"><Plus size={15}/> Filters <span className="text-black/40 font-normal">{items.length} styles</span></button><label className="flex items-center gap-2 text-xs">Sort by <select value={sort} onChange={e=>setSort(e.target.value)} className="font-bold bg-transparent outline-none"><option>Featured</option><option>Price: Low to high</option><option>Top rated</option></select></label></div>{items.length===0?<div className="py-20 text-center text-sm text-black/45">No products are currently listed to the Website Storefront channel.</div>:<div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 2xl:grid-cols-6 gap-x-3 md:gap-x-4 gap-y-6 mt-5">{shown.map(p=><ProductCard key={p.id} p={p}/>)}</div>}{limit<items.length&&<div className="text-center mt-9"><p className="text-xs text-black/50 mb-3">Showing {limit} of {items.length} styles</p><button onClick={()=>setLimit(x=>Math.min(items.length,x+18))} className="btn border border-wine text-wine">Load more</button></div>}</div></main>}

export function ProductPage({slug}:{slug:string}){
  const [p,setProduct]=useState<Product|null>(null),[loading,setLoading]=useState(true),[error,setError]=useState('');
  const {add,toggleWish,wish}=useStore();
  const [color,setColor]=useState(''),[size,setSize]=useState(''),[qty,setQty]=useState(1),[activeImage,setActiveImage]=useState(0);

  useEffect(()=>{
    let alive=true;
    setLoading(true);setError('');setProduct(null);setActiveImage(0);
    fetchProduct(slug).then(product=>{
      if(!alive)return;
      setProduct(product);
      const first=product.variants?.[0];
      if(first?.color)setColor(first.color);
      if(first?.size)setSize(first.size);
    }).catch(()=>{
      if(alive)setError('This product is not available on the Website Storefront channel.');
    }).finally(()=>{
      if(alive)setLoading(false);
    });
    return()=>{alive=false};
  },[slug]);

  useEffect(()=>{
    if(!p?.variants?.length)return;
    const variants=p.variants;
    const colors=uniqueValues(variants.map(v=>v.color));
    const currentColor=color&&colors.includes(color)?color:colors[0]??'';
    const sizesForColor=uniqueValues(variants.filter(v=>!currentColor||v.color===currentColor).map(v=>v.size));
    const currentSize=size&&sizesForColor.includes(size)?size:sizesForColor[0]??'';
    if(currentColor!==color)setColor(currentColor);
    if(currentSize!==size)setSize(currentSize);
  },[p,color,size]);

  if(loading)return <main className="container py-24 text-center text-sm text-black/45">Loading product from backend...</main>;
  if(error||!p)return <main className="container py-24 text-center"><h1 className="serif text-3xl">Product unavailable</h1><p className="mt-3 text-sm text-black/50">{error||'No backend product was found for this page.'}</p><Link href="/category/all" className="btn btn-dark mt-7">View listed products</Link></main>;

  const variants=p.variants??[];
  const colors=uniqueValues(variants.map(v=>v.color));
  const sizesForColor=uniqueValues(variants.filter(v=>!color||v.color===color).map(v=>v.size));
  const selectedVariant=variants.find(v=>v.color===color&&v.size===size)||variants.find(v=>v.color===color)||variants.find(v=>v.size===size)||variants[0];
  const gallery=(p.images?.length?p.images:[p.image]).filter(Boolean);
  const currentImage=gallery[Math.min(activeImage,gallery.length-1)]||p.image;
  const storefrontPrice=selectedVariant?.price??p.price;
  const storefrontOriginal=selectedVariant?.compareAtPrice??p.original;
  const off=Math.max(0,Math.round((1-storefrontPrice/Math.max(storefrontOriginal,storefrontPrice))*100));

  return <main className="container py-5 md:py-7">
    <div className="text-xs text-black/45 mb-4"><Link href="/">Home</Link> / <Link href={`/category/${p.category.toLowerCase().replaceAll(' ','-')}`}>{p.category}</Link> / {p.name}</div>
    <div className="grid lg:grid-cols-[minmax(0,560px)_minmax(360px,1fr)] gap-6 lg:gap-10 items-start">
      <div className="lg:sticky lg:top-24">
        <div className="relative overflow-hidden rounded-2xl bg-blush border border-black/5">
          <img src={currentImage} alt={p.name} className="w-full max-h-[620px] aspect-[4/5] object-cover"/>
          {p.badge?<span className="absolute left-3 top-3 rounded-full bg-white/95 px-3 py-1 text-[9px] font-extrabold tracking-[.16em] text-wine shadow-sm">{p.badge}</span>:null}
        </div>
        {gallery.length>1?<div className="mt-3 flex gap-2 overflow-x-auto pb-1 hide-scroll">
          {gallery.map((img,i)=><button key={`${img}-${i}`} onClick={()=>setActiveImage(i)} className={`shrink-0 overflow-hidden rounded-xl border bg-blush ${activeImage===i?'border-wine ring-2 ring-wine/15':'border-black/10'}`} aria-label={`View product image ${i+1}`}>
            <img src={img} alt="" className="h-16 w-14 md:h-20 md:w-16 object-cover"/>
          </button>)}
        </div>:null}
      </div>

      <div className="lg:sticky lg:top-24 self-start rounded-3xl border border-black/5 bg-white p-5 md:p-7 shadow-[0_18px_50px_rgba(31,20,24,.06)]">
        <span className="eyebrow">{p.badge||'Backend storefront product'}</span>
        <h1 className="serif text-3xl md:text-5xl font-semibold mt-2 leading-tight">{p.name}</h1>
        <div className="mt-4 flex flex-wrap items-center gap-3 text-sm">
          <span className="inline-flex items-center gap-1 rounded-full bg-gold/15 px-3 py-1 font-bold"><Star size={14} fill="#CFA15C" color="#CFA15C"/>{p.rating.toFixed(1)}</span>
          <span className="text-black/45">{p.reviews} verified reviews</span>
          <span className="text-black/30">•</span>
          <span className="text-black/55">{selectedVariant?.color||p.color}</span>
        </div>
        <div className="mt-6 flex flex-wrap items-end gap-3">
          <span className="text-3xl font-extrabold tracking-tight">{money(storefrontPrice)}</span>
          {storefrontOriginal>storefrontPrice?<s className="pb-1 text-base text-black/35">{money(storefrontOriginal)}</s>:null}
          {off>0?<span className="mb-1 rounded-full bg-green-50 px-2.5 py-1 text-xs font-bold text-green-700">{off}% OFF</span>:null}
        </div>
        <p className="mt-1 text-xs text-black/45">Inclusive of all taxes. Price updates from the selected variant when an override exists.</p>

        <hr className="my-6"/>
        {colors.length>0?<div>
          <div className="flex items-center justify-between"><b>Select colour</b><span className="text-xs font-semibold text-black/45">{color||'Default'}</span></div>
          <div className="mt-3 flex flex-wrap gap-2.5">{colors.map(c=><button onClick={()=>{setColor(c);const firstSize=variants.find(v=>v.color===c)?.size;if(firstSize)setSize(firstSize)}} key={c} className={`flex min-h-11 items-center gap-2 rounded-full border px-3 text-sm font-bold transition ${color===c?'border-wine bg-blush text-wine ring-2 ring-wine/10':'hover:border-wine'}`} aria-label={`Select colour ${c}`}>
            <span className="h-5 w-5 rounded-full border border-black/10 shadow-inner" style={{background:colorSwatch(c)}}/>
            {c}
          </button>)}</div>
        </div>:null}
        {colors.length>0?<hr className="my-6"/>:null}
        <div className="flex justify-between"><b>Select size</b><button className="text-wine underline text-xs">Size guide</button></div>
        <div className="mt-3 flex flex-wrap gap-2.5">{sizesForColor.map(s=>{const v=variants.find(item=>item.color===color&&item.size===s)||variants.find(item=>item.size===s);const disabled=Number(v?.stockQuantity??0)<=0;return <button onClick={()=>setSize(s)} disabled={disabled} key={s} className={`h-11 min-w-11 rounded-full border px-3 text-sm font-bold transition disabled:cursor-not-allowed disabled:opacity-40 ${size===s?'bg-wine text-white border-wine':'hover:border-wine'}`}>{s}</button>})}</div>
        <p className={`mt-3 text-xs ${(selectedVariant?.stockQuantity??0)>0?'text-green-700':'text-red-700'}`}>{(selectedVariant?.stockQuantity??0)>0?`● In stock — ${selectedVariant?.stockQuantity??0} left for ${color||'selected colour'} / ${size||'selected size'}`:'● Out of stock for this colour / size'}</p>

        <div className="mt-6 grid grid-cols-[auto_1fr_auto] gap-3">
          <div className="h-14 rounded-full border flex items-center">
            <button onClick={()=>setQty(Math.max(1,qty-1))} className="px-3" aria-label="Decrease quantity"><Minus size={16}/></button>
            <span className="min-w-5 text-center font-bold">{qty}</span>
            <button onClick={()=>setQty(qty+1)} className="px-3" aria-label="Increase quantity"><Plus size={16}/></button>
          </div>
          <button onClick={()=>{for(let i=0;i<qty;i++)add(p.id,selectedVariant?.id??p.variant_id)}} disabled={!selectedVariant||(selectedVariant.stockQuantity??0)<=0} className="btn btn-dark rounded-full disabled:opacity-50"><ShoppingBag size={18}/> Add to bag</button>
          <button onClick={()=>toggleWish(p.id)} className={`h-14 w-14 rounded-full border grid place-items-center ${wish.includes(p.id)?'bg-blush text-wine':''}`} aria-label="Add to wishlist"><Heart fill={wish.includes(p.id)?'currentColor':'none'}/></button>
        </div>

        <div className="mt-6 rounded-2xl bg-blush p-4 flex gap-3">
          <Truck className="text-wine shrink-0"/>
          <div><b>Delivery in 3–6 business days</b><p className="text-xs text-black/50 mt-1">Shipping is calculated dynamically at checkout.</p></div>
        </div>
        {['Details & craftsmanship','Fit & size','Fabric & care','Shipping & returns'].map(x=><details key={x} className="border-b py-4">
          <summary className="flex justify-between font-bold cursor-pointer">{x}<ChevronDown size={17}/></summary>
          <p className="text-black/55 leading-7 pt-3">{p.description||'Designed in our studio and finished by skilled artisans. Soft, breathable fabric with thoughtful detailing for lasting comfort.'}</p>
        </details>)}
      </div>
    </div>
    <Recommendations exclude={p.id}/>
  </main>
}

function Recommendations({exclude}:{exclude:string}){const [items,setItems]=useState<Product[]>([]);useEffect(()=>{fetchProducts('all',12).then(list=>setItems(list.filter(p=>p.id!==exclude).slice(0,4)))},[exclude]);return <div className="section"><p className="eyebrow">Style it your way</p><h2 className="title mt-3">You may also love</h2><div className="grid grid-cols-2 lg:grid-cols-4 gap-5 mt-9">{items.map(x=><ProductCard p={x} key={x.id}/>)}</div></div>}

export function CartPage(){const {cart,remove}=useStore();const [catalog,setCatalog]=useState<Product[]>([]);useEffect(()=>{fetchProducts('all',100).then(setCatalog)},[]);const list=cart.map(item=>catalog.find(p=>p.id===item.productId)).filter(Boolean) as Product[];const subtotal=list.reduce((a,b)=>a+b.price,0);return <main className="container py-12"><h1 className="title">Your bag <span className="text-xl text-black/35">({cart.length})</span></h1>{!list.length?<Empty icon={<ShoppingBag/>} title="Your bag is waiting" text="Fall in love with something beautiful." action="Start shopping" href="/category/new-arrivals"/>:<div className="grid lg:grid-cols-[1fr_400px] gap-12 mt-10"><div>{list.map((p,i)=><div className="flex gap-5 py-5 border-b" key={`${p.id}-${i}`}><img src={p.image} className="w-28 md:w-36 aspect-[3/4] object-cover rounded-xl"/><div className="flex-1"><h2 className="font-bold">{p.name}</h2><p className="text-xs text-black/45 mt-2">{p.color} · Size M</p><p className="font-bold mt-4">{money(p.price)}</p><button onClick={()=>remove(p.id)} className="text-xs underline mt-6">Remove</button></div></div>)}</div><aside className="bg-blush rounded-3xl p-7 self-start lg:sticky lg:top-28"><h2 className="serif text-2xl font-bold">Order summary</h2><div className="flex justify-between mt-6"><span>Subtotal</span><b>{money(subtotal)}</b></div><div className="flex justify-between mt-3 text-green-700"><span>Delivery</span><b>Calculated at checkout</b></div><div className="border-t mt-6 pt-6 flex justify-between text-lg"><b>Total</b><b>{money(subtotal)}</b></div><Link href="/checkout" className="btn btn-dark w-full mt-6">Secure checkout <ChevronRight size={18}/></Link><p className="text-center text-[10px] text-black/45 mt-4">UPI · CARDS · NET BANKING · COD</p></aside></div>}</main>}

export function WishlistPage(){const {wish}=useStore();const [catalog,setCatalog]=useState<Product[]>([]);useEffect(()=>{fetchProducts('all',100).then(setCatalog)},[]);const list=wish.map(id=>catalog.find(p=>p.id===id)).filter(Boolean) as Product[];return <main className="container py-12"><h1 className="title">Your wishlist</h1>{!list.length?<Empty icon={<Heart/>} title="Save what you love" text="Tap the heart on any style to find it here." action="Discover styles" href="/category/all"/>:<div className="grid grid-cols-2 lg:grid-cols-4 gap-5 mt-10">{list.map(p=><ProductCard p={p} key={p.id}/>)}</div>}</main>}

function Empty({icon,title,text,action,href}:{icon:React.ReactNode,title:string;text:string;action:string;href:string}){return <div className="text-center py-28"><div className="w-16 h-16 rounded-full bg-blush grid place-items-center mx-auto text-wine">{icon}</div><h2 className="serif text-3xl mt-5">{title}</h2><p className="text-black/50 mt-2">{text}</p><Link href={href} className="btn btn-dark mt-7">{action}</Link></div>}

function uniqueValues(values:Array<string|undefined|null>){return Array.from(new Set(values.map(v=>String(v??'').trim()).filter(Boolean)))}
function colorSwatch(color:string){const key=color.toLowerCase().trim();const map:Record<string,string>={wine:'#7f1d38',ivory:'#f8f1df',sage:'#9aaa85',indigo:'#243b6b',rose:'#d7a2a8',black:'#111827',white:'#ffffff',red:'#b91c1c',blue:'#2563eb',green:'#15803d',gold:'#cfa15c',pink:'#ec4899',yellow:'#eab308',purple:'#7c3aed',orange:'#f97316'};return map[key]||'#e5e7eb'}

export function Checkout(){
  const {cart,clear,cartToken,markCheckoutStarted}=useStore();
  const [config,setConfig]=useState<StorefrontConfig>(fallbackConfig);
  const [step,setStep]=useState(1),[placing,setPlacing]=useState(false),[done,setDone]=useState<any>(null),[form,setForm]=useState({name:'',email:'',phone:'',pin:'',city:'Mumbai',state:'MH',address:'',couponCode:'WELCOME10'});
  useEffect(()=>{fetchStorefrontConfig().then(setConfig)},[]);
  const shipping=config.shippingOptions?.length?config.shippingOptions:[{id:'default',name:'Standard delivery',country_code:'IN',region_codes:[],rate_type:'flat',rate:0,free_shipping_threshold:0,estimated_days_min:3,estimated_days_max:7,cod_enabled:true}];
  const payments=config.paymentMethods?.length?config.paymentMethods:[{id:'cod',code:'cod',name:'Cash on Delivery',provider:'manual',instructions:'Pay when the order is delivered.',sort_order:1}];
  async function place(){setPlacing(true);try{const res=await fetch(`${API_BASE}/v1/storefront/${STORE_SLUG}/checkout`,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({cartToken,name:form.name||'Guest Customer',email:form.email||'guest@example.test',phone:form.phone,couponCode:form.couponCode,shippingAddress:{pin:form.pin,city:form.city,state:form.state,address:form.address},items:cart.map(item=>({productId:item.productId,variantId:item.variantId,quantity:1}))})});const data=await res.json();if(!res.ok)throw new Error(data?.error?.message||'Could not place order');setDone(data);clear()}catch(e:any){alert(e.message)}finally{setPlacing(false)}}
  if(done)return <main className="container py-20 max-w-3xl text-center"><div className="mx-auto grid h-16 w-16 place-items-center rounded-full bg-green-100 text-green-700"><Check/></div><h1 className="title mt-5">Order placed</h1><p className="mt-3 text-black/55">Your order <b>{done.orderNumber}</b> has been created from the SaaS commerce backend.</p><p className="mt-2 font-bold">{money(done.total)}</p><Link href="/category/all" className="btn btn-dark mt-8">Continue shopping</Link></main>;
  return <main className="container py-12 max-w-5xl"><h1 className="title">Checkout</h1><div className="flex mt-8 mb-12">{['Address','Delivery','Payment','Review'].map((x,i)=><div className="flex-1" key={x}><span className={`w-8 h-8 rounded-full inline-grid place-items-center ${step>=i+1?'bg-wine text-white':'bg-black/10'}`}>{step>i+1?<Check size={15}/>:i+1}</span><b className="ml-2 hidden md:inline">{x}</b><div className={`h-[2px] mt-3 ${step>i+1?'bg-wine':'bg-black/10'}`}/></div>)}</div><div className="max-w-2xl"><h2 className="serif text-3xl">{['','Where should we deliver?','Choose delivery','Payment method','Review your order'][step]}</h2>{step===1&&<div className="grid md:grid-cols-2 gap-4 mt-7">{[['name','Full name'],['email','Email'],['phone','Mobile number'],['pin','PIN code'],['city','City'],['state','State'],['address','Address'],['couponCode','Coupon code']].map(([key,label])=><input key={key} value={(form as any)[key]} onChange={e=>setForm(v=>({...v,[key]:e.target.value}))} placeholder={label} className="border rounded-xl p-4 outline-wine"/>)}</div>}{step===2&&<div className="grid gap-3 mt-7">{shipping.map(option=><label className="border rounded-xl p-5 flex gap-3" key={option.id}><input type="radio" name="ship" defaultChecked/><div><b>{option.name}</b><p className="text-sm text-black/50 mt-1">{money(Number(option.rate||0))} delivery · free above {money(Number(option.free_shipping_threshold||0))} · {option.estimated_days_min}-{option.estimated_days_max} days {option.cod_enabled?'· COD available':''}</p></div></label>)}</div>}{step===3&&<div className="grid gap-3 mt-7">{payments.map((option,index)=><label className="border rounded-xl p-5 flex gap-3" key={option.id||option.code}><input type="radio" name="pay" defaultChecked={index===0}/> <span><b>{option.name}</b>{option.instructions?<p className="text-sm text-black/50 mt-1">{option.instructions}</p>:null}</span></label>)}</div>}{step===4&&<div className="bg-blush p-7 rounded-xl mt-7"><b>Everything looks beautiful.</b><p className="mt-2 text-black/50">This will create a real ecommerce order record in the SaaS backend and reduce inventory.</p></div>}<button onClick={()=>{if(step===4)place();else{if(step===1)markCheckoutStarted({name:form.name,email:form.email,phone:form.phone});setStep(Math.min(4,step+1))}}} disabled={placing||cart.length===0} className="btn btn-dark mt-8 disabled:opacity-50">{step===4?(placing?'Placing...':'Place order'):'Continue'} <ChevronRight size={17}/></button>{cart.length===0?<p className="mt-3 text-xs text-red-700">Your bag is empty.</p>:null}</div></main>
}
