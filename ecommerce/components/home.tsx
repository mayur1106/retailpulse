'use client';
import Link from 'next/link';
import {useEffect,useMemo,useState} from 'react';
import {ArrowRight,Leaf,RotateCcw,ShieldCheck,Truck} from 'lucide-react';
import {Product,StorefrontConfig,StorefrontSection,fallbackConfig,fetchProducts,fetchStorefrontConfig} from '../lib/catalog';
import {ProductCard} from './product-card';

export function Home(){
  const [items,setItems]=useState<Product[]>([]);
  const [config,setConfig]=useState<StorefrontConfig>(fallbackConfig);
  useEffect(()=>{fetchProducts('all',80).then(setItems);fetchStorefrontConfig().then(setConfig)},[]);
  const sections=useMemo(()=>[...(config.sections?.length?config.sections:fallbackConfig.sections)].sort((a,b)=>Number(a.sort_order??0)-Number(b.sort_order??0)),[config.sections]);
  return <main>{sections.map(section=><Section key={section.section_key} section={section} items={items} config={config}/>)}</main>;
}

function Section({section,items,config}:{section:StorefrontSection;items:Product[];config:StorefrontConfig}){
  if(section.section_type==='hero')return <Hero section={section}/>;
  if(section.section_type==='category_tiles')return <CategoryTiles section={section} config={config}/>;
  if(section.section_type==='product_grid')return <ProductSection section={section} items={items}/>;
  if(section.section_type==='promo_tiles')return <PromoTiles section={section}/>;
  if(section.section_type==='banner')return <Banner section={section}/>;
  if(section.section_type==='benefits')return <Benefits section={section} config={config}/>;
  if(section.section_type==='newsletter')return <Newsletter section={section} config={config}/>;
  return null;
}

function Hero({section}:{section:StorefrontSection}){const eyebrow=String(section.content?.eyebrow??'');return <section className="relative h-[310px] sm:h-[380px] lg:h-[430px] overflow-hidden">
  <img src={section.image_url||'/images/rangavali-hero.png'} className="absolute inset-0 w-full h-full object-cover object-[68%_center]" alt={section.title}/>
  <div className="absolute inset-0 bg-gradient-to-r from-[#351306]/65 via-[#351306]/22 to-transparent"/>
  <div className="container relative h-full flex items-center text-white"><div className="max-w-[530px]">
    {eyebrow?<p className="text-[9px] md:text-[11px] tracking-[.22em] font-bold">{eyebrow}</p>:null}
    <h1 className="serif text-[42px] sm:text-[55px] lg:text-[68px] leading-[.88] font-semibold mt-3">{section.title}</h1>
    {section.subtitle?<p className="hidden sm:block mt-4 max-w-md text-white/85 leading-6">{section.subtitle}</p>:null}
    {section.cta_label?<Link href={section.cta_href||'/category/all'} className="inline-flex items-center gap-2 bg-white text-wine min-h-10 px-5 font-extrabold text-xs mt-5">{section.cta_label} <ArrowRight size={15}/></Link>:null}
  </div></div>
</section>}

function CategoryTiles({section,config}:{section:StorefrontSection;config:StorefrontConfig}){const categories=(config.categories?.length?config.categories:fallbackConfig.categories).slice(0,section.max_items||8);return <section className="shop-section"><div className="container">
  <SectionHead title={section.title||'Shop by category'} subtitle={section.subtitle} href={section.cta_href||'/category/all'}/>
  <div className="grid grid-cols-4 md:grid-cols-8 gap-2 md:gap-3 mt-4">{categories.map((category,index)=><Link href={`/category/${category.slug}`} key={category.slug} className="group text-center min-w-0">
    <div className="overflow-hidden bg-blush"><img src={category.image||`/images/catalog/look-${(index%8)+1}.jpg`} alt={category.name} className="w-full aspect-[3/4] object-cover group-hover:scale-105 transition duration-500"/></div>
    <b className="block mt-2 text-[10px] md:text-xs truncate">{category.name}</b>
  </Link>)}</div>
</div></section>}

function ProductSection({section,items}:{section:StorefrontSection;items:Product[]}){const filtered=items.filter(product=>!section.category_slug||['all','all-styles'].includes(section.category_slug)||slug(product.category)===section.category_slug);const source=section.product_source==='featured'?filtered.filter(product=>product.badge):filtered;return <section className={`shop-section ${section.layout==='muted'||section.section_key==='trending'?'bg-[#fafafa]':''}`}><div className="container">
  <SectionHead title={section.title} subtitle={section.subtitle} href={section.cta_href}/>
  <ProductGrid items={source.slice(0,section.max_items||12)}/>
</div></section>}

function PromoTiles({section}:{section:StorefrontSection}){const tiles=Array.isArray(section.content?.tiles)?section.content.tiles as Array<Record<string,string>>:[];if(!tiles.length)return null;return <section className="shop-section"><div className="container">
  <SectionHead title={section.title||'Offers'} subtitle={section.subtitle} href={section.cta_href}/>
  <div className="grid md:grid-cols-3 gap-2 md:gap-3 mt-4">{tiles.slice(0,section.max_items||3).map((tile,index)=><Link href={tile.href||section.cta_href||'/category/all'} key={`${tile.title}-${index}`} className="relative h-[180px] md:h-[230px] overflow-hidden group">
    <img src={tile.imageUrl||`/images/catalog/look-${index+1}.jpg`} alt={tile.title||section.title} className="absolute inset-0 w-full h-full object-cover group-hover:scale-105 transition duration-500"/>
    <div className="absolute inset-0 bg-gradient-to-r from-black/65 to-transparent"/>
    <div className="absolute inset-0 p-5 md:p-7 flex flex-col justify-end text-white"><b className="text-lg md:text-2xl">{tile.title}</b><span className="text-xs text-white/75 mt-1">{tile.subtitle}</span><span className="text-[10px] font-extrabold mt-3">{section.cta_label||'SHOP NOW'} →</span></div>
  </Link>)}</div>
</div></section>}

function Banner({section}:{section:StorefrontSection}){const eyebrow=String(section.content?.eyebrow??'');return <section className="container py-5">
  <Link href={section.cta_href||'/category/all'} className="relative min-h-[170px] md:min-h-[210px] overflow-hidden bg-wine text-white px-6 md:px-10 py-7 flex items-center">
    <img src={section.image_url||'/images/catalog/look-8.jpg'} alt="" className="absolute right-0 top-0 h-full w-[46%] object-cover opacity-75"/>
    <div className="absolute inset-0 bg-gradient-to-r from-wine via-wine/95 to-transparent"/>
    <div className="relative max-w-xl">{eyebrow?<p className="text-[9px] tracking-[.2em] text-gold font-bold">{eyebrow}</p>:null}<h2 className="serif text-3xl md:text-5xl mt-2">{section.title}</h2>{section.subtitle?<p className="hidden sm:block text-white/65 mt-2">{section.subtitle}</p>:null}<span className="inline-flex items-center gap-2 text-xs font-extrabold mt-4">{section.cta_label||'SHOP NOW'} <ArrowRight size={15}/></span></div>
  </Link>
</section>}

function Benefits({section,config}:{section:StorefrontSection;config:StorefrontConfig}){const items=Array.isArray(section.content?.items)?section.content.items as Array<Record<string,string>>:[];const fallback=[{icon:'leaf',label:'Premium fabrics'},{icon:'returns',label:String(config.store.settings?.returnPolicy??'Easy returns')},{icon:'truck',label:config.shippingOptions?.[0]?`${config.shippingOptions[0].estimated_days_min}-${config.shippingOptions[0].estimated_days_max} day delivery`:'Tracked delivery'},{icon:'shield',label:config.paymentMethods?.length?'Secure payments':'COD available'}];return <section className="bg-wine text-white py-7"><div className="container grid grid-cols-2 md:grid-cols-4 gap-5 text-center">{(items.length?items:fallback).slice(0,4).map(item=>{const I=iconFor(item.icon);return <div key={item.label} className="flex items-center justify-center gap-3"><I className="text-gold" size={19}/><b className="text-[10px] md:text-xs">{item.label}</b></div>})}</div></section>}

function Newsletter({section,config}:{section:StorefrontSection;config:StorefrontConfig}){const title=section.title||String(config.store.settings?.newsletterTitle??'Join our newsletter');return <section className="container py-9 text-center"><p className="text-[9px] tracking-[.2em] font-extrabold text-wine">THE {String(config.store.settings?.brandName??config.store.name??'STORE').toUpperCase()} LETTER</p><h2 className="serif text-3xl md:text-4xl mt-2">{title}</h2>{section.subtitle?<p className="mt-2 text-xs text-black/45">{section.subtitle}</p>:null}<form className="max-w-md mx-auto flex border mt-5"><input className="flex-1 px-4 h-11 outline-none min-w-0" type="email" placeholder="Enter your email address"/><button className="bg-wine text-white px-5 text-xs font-extrabold">{section.cta_label||'SUBSCRIBE'}</button></form></section>}

function SectionHead({title,subtitle,href}:{title:string;subtitle?:string;href?:string}){return <div className="flex items-end justify-between gap-4"><div><h2 className="text-lg md:text-2xl font-extrabold tracking-tight">{title}</h2>{subtitle&&<p className="text-[10px] md:text-xs text-black/45 mt-1">{subtitle}</p>}</div>{href&&<Link href={href} className="text-[10px] md:text-xs font-extrabold text-wine whitespace-nowrap">VIEW ALL →</Link>}</div>}
function ProductGrid({items}:{items:Product[]}){return <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-x-2 md:gap-x-3 gap-y-5 mt-4">{items.map(p=><ProductCard key={p.id} p={p}/>)}</div>}
function slug(value:string){return value.toLowerCase().replaceAll(' ','-')}
function iconFor(icon?:string){if(icon==='returns')return RotateCcw;if(icon==='truck')return Truck;if(icon==='shield')return ShieldCheck;return Leaf}
