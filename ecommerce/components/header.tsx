'use client';
import Link from 'next/link';
import {Heart,Menu,Search,ShoppingBag,User,X} from 'lucide-react';
import {useEffect,useState} from 'react';
import {useStore} from './store-provider';
import {Product,StorefrontConfig,fallbackConfig,fetchProducts,fetchStorefrontConfig,money} from '../lib/catalog';

export function Header(){
  const {cart,wish}=useStore();
  const [config,setConfig]=useState<StorefrontConfig>(fallbackConfig);
  const [menu,setMenu]=useState(false),[search,setSearch]=useState(false),[q,setQ]=useState(''),[searchItems,setSearchItems]=useState<Product[]>([]);
  useEffect(()=>{fetchStorefrontConfig().then(setConfig)},[]);
  useEffect(()=>{if(search&&searchItems.length===0)fetchProducts('all',40).then(setSearchItems)},[search,searchItems.length]);
  const settings=config.store.settings??{};
  const brand=String(settings.brandName??config.store.name??'Rangavali');
  const announcement=String(settings.announcement??'');
  const nav=(config.categories?.length?config.categories:fallbackConfig.categories).slice(0,8).map(category=>({label:category.name,href:`/category/${category.slug}`}));
  const actions=[
    {Icon:User,label:'Profile',href:'/account',count:0},
    {Icon:Heart,label:'Wishlist',href:'/wishlist',count:wish.length},
    {Icon:ShoppingBag,label:'Bag',href:'/cart',count:cart.length}
  ];
  return <>
    {announcement?<div className="bg-wine text-white text-center py-1.5 text-[9px] md:text-[10px] tracking-[.14em]">{announcement}</div>:null}
    <header className="sticky top-0 z-40 bg-white/95 backdrop-blur border-b border-black/5">
      <div className="container h-[66px] flex items-center gap-5 xl:gap-8">
        <button onClick={()=>setMenu(true)} className="md:hidden" aria-label="Open menu"><Menu/></button>
        <Link href="/" className="serif text-[26px] xl:text-[30px] font-bold tracking-[.07em] text-wine shrink-0">{config.store.logo_url?<img src={config.store.logo_url} alt={brand} className="h-9 max-w-[180px] object-contain"/>:brand.toUpperCase()}</Link>
        <nav className="desktop flex items-stretch self-stretch gap-4 xl:gap-6 text-[11px] font-extrabold uppercase tracking-[.03em] whitespace-nowrap">
          {nav.map(n=><Link key={n.href} href={n.href} className={`flex items-center border-b-2 border-transparent hover:border-wine hover:text-wine ${n.label.toLowerCase().includes('sale')?'text-wine':''}`}>{n.label}</Link>)}
        </nav>
        <button onClick={()=>setSearch(true)} className="desktop ml-auto max-w-[320px] min-w-[190px] flex-1 h-10 rounded-md bg-[#f6f5f5] text-black/45 px-3 flex items-center gap-3 text-left" aria-label="Search">
          <Search size={17}/><span className="truncate">Search for styles and collections</span>
        </button>
        <div className="ml-auto md:ml-0 flex items-center gap-4">
          <button onClick={()=>setSearch(true)} className="md:hidden" aria-label="Search"><Search size={20}/></button>
          {actions.map(({Icon,label,href,count})=><Link key={label} className={`relative flex flex-col items-center gap-0.5 text-[9px] font-bold ${label==='Profile'?'desktop':''}`} href={href}>
            <Icon size={19}/><span className="desktop">{label}</span>
            {count>0&&<i className="absolute -top-2 right-0 not-italic text-[8px] bg-wine text-white rounded-full w-4 h-4 grid place-items-center">{count}</i>}
          </Link>)}
        </div>
      </div>
    </header>
    {menu&&<div className="fixed inset-0 z-50 bg-white p-6"><div className="flex justify-between"><span className="serif text-2xl text-wine">{brand}</span><button onClick={()=>setMenu(false)}><X/></button></div><div className="mt-12 grid gap-6 text-2xl serif">{nav.map(n=><Link onClick={()=>setMenu(false)} key={n.href} href={n.href}>{n.label}</Link>)}<Link href="/account">My account</Link></div></div>}
    {search&&<div className="fixed inset-0 z-50 bg-black/30 backdrop-blur-sm"><div className="bg-white p-6 md:p-10 shadow-soft"><div className="max-w-3xl mx-auto"><div className="flex border-b-2 border-wine py-3 gap-3"><Search/><input autoFocus value={q} onChange={e=>setQ(e.target.value)} className="flex-1 outline-none text-lg" placeholder="Search styles, colours, occasions…"/><button onClick={()=>setSearch(false)}><X/></button></div><p className="eyebrow mt-7 mb-4">{q?'Suggestions':'Trending now'}</p><div className="grid grid-cols-2 md:grid-cols-4 gap-4">{searchItems.filter(p=>p.name.toLowerCase().includes(q.toLowerCase())).slice(0,4).map(p=><Link onClick={()=>setSearch(false)} href={`/product/${p.id}`} key={p.id}><img src={p.image} className="w-full aspect-[3/4] object-cover rounded-lg" alt={p.name}/><b className="block mt-2 text-xs">{p.name}</b><span className="text-xs">{money(p.price)}</span></Link>)}</div>{searchItems.length===0?<p className="mt-5 text-sm text-black/45">No backend products are listed to the Website Storefront channel yet.</p>:null}</div></div></div>}
  </>;
}
