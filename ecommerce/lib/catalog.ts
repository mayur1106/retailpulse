import { apiBaseUrl, storeSlug } from './env';

export type Product={id:string;variant_id?:string;slug?:string;name:string;category:string;price:number;original:number;rating:number;reviews:number;image:string;images?:string[];badge?:string;color:string;description?:string;variants?:Array<{id:string;title:string;sku:string;color:string;size:string;price:number;compareAtPrice?:number;costPrice?:number;stockQuantity:number}>};
export type StorefrontCategory={id:string;name:string;slug:string;description?:string;image?:string;sort_order?:number};
export type StorefrontSection={id:string;section_key:string;section_type:string;title:string;subtitle:string;layout:string;image_url:string;cta_label:string;cta_href:string;category_slug:string;product_source:string;max_items:number;content:Record<string,any>;sort_order:number};
export type StorefrontPayment={id:string;code:string;name:string;provider:string;instructions:string;sort_order:number;settings?:Record<string,any>};
export type StorefrontShipping={id:string;name:string;country_code:string;region_codes:string[];rate_type:string;rate:number;free_shipping_threshold:number;estimated_days_min:number;estimated_days_max:number;cod_enabled:boolean};
export type StorefrontConfig={store:{id:string;name:string;slug:string;domain?:string;logo_url?:string;currency_code:string;country_code:string;settings:Record<string,any>};categories:StorefrontCategory[];sections:StorefrontSection[];paymentMethods:StorefrontPayment[];shippingOptions:StorefrontShipping[]};
export const money=(n:number)=>new Intl.NumberFormat('en-IN',{style:'currency',currency:'INR',maximumFractionDigits:0}).format(n);

export const STOREFRONT_BASE_URL=process.env.NEXT_PUBLIC_STOREFRONT_BASE_URL??'http://localhost:3006';

function normalizeProduct(raw:any):Product{
  const images=Array.isArray(raw.images)?raw.images:(raw.images?JSON.parse(String(raw.images)):[raw.image]).filter(Boolean);
  return {
    id:String(raw.id),
    variant_id:raw.variant_id??raw.variantId,
    slug:String(raw.slug??raw.id),
    name:String(raw.name??raw.title),
    category:String(raw.category??''),
    price:Number(raw.price??0),
    original:Number(raw.original??raw.compare_at_price??raw.price??0),
    rating:Number(raw.rating??4.5),
    reviews:Number(raw.reviews??0),
    image:String(raw.image??images?.[0]??'/images/catalog/look-1.jpg'),
    images,
    badge:String(raw.badge??''),
    color:String(raw.color??'Classic'),
    description:String(raw.description??'Premium women’s apparel crafted for everyday elegance.'),
    variants:Array.isArray(raw.variants)?raw.variants:[],
  };
}

export async function fetchProducts(category?:string,limit=60):Promise<Product[]>{
  try{
    const params=new URLSearchParams({limit:String(limit)});
    if(category)params.set('category',category);
    const res=await fetch(`${apiBaseUrl()}/v1/storefront/${storeSlug}/products?${params}`,{cache:'no-store'});
    if(!res.ok)throw new Error('storefront unavailable');
    const data=await res.json();
    const items=(data.items??[]).map(normalizeProduct);
    return items;
  }catch(error){console.error('Storefront API failed; not showing static fallback catalog',error);return []}
}

export async function fetchProduct(slug:string):Promise<Product>{
  const res=await fetch(`${apiBaseUrl()}/v1/storefront/${storeSlug}/products/${slug}`,{cache:'no-store'});
    if(!res.ok)throw new Error('product unavailable');
    return normalizeProduct(await res.json());
}

export async function fetchStorefrontConfig():Promise<StorefrontConfig>{
  try{
    const res=await fetch(`${apiBaseUrl()}/v1/storefront/${storeSlug}/config`,{cache:'no-store'});
    if(!res.ok)throw new Error('storefront config unavailable');
    const data=await res.json();
    return {
      store:data.store??fallbackConfig.store,
      categories:Array.isArray(data.categories)?data.categories:fallbackConfig.categories,
      sections:Array.isArray(data.sections)?data.sections:fallbackConfig.sections,
      paymentMethods:Array.isArray(data.paymentMethods)?data.paymentMethods:[],
      shippingOptions:Array.isArray(data.shippingOptions)?data.shippingOptions:[],
    };
  }catch(error){console.error('Storefront CMS config failed; using safe defaults',error);return fallbackConfig}
}

export const fallbackConfig:StorefrontConfig={
  store:{id:'fallback',name:'Rangavali',slug:'rangavali',logo_url:'',currency_code:'INR',country_code:'IN',settings:{brandName:'Rangavali',tagline:'Modern Indian clothing, thoughtfully made for women who collect moments, not trends.',announcement:'10% off first order · Free shipping above ₹1,999',returnPolicy:'Easy 7-day returns on unworn styles.',newsletterTitle:'₹300 off your first order'}},
  categories:[
    {name:'Kurtas',image:'/images/products/kurta-beige-chanderi-anarkali.png'},
    {name:'Salwar Suits',image:'/images/products/suit-turquoise-printed-dupatta.png'}
  ].map((category,index)=>({id:category.name, name:category.name, slug:category.name.toLowerCase().replaceAll(' ','-'), image:category.image, sort_order:index+1})),
  sections:[
    {id:'hero',section_key:'hero',section_type:'hero',title:'Kurtas and suit sets for every day of celebration.',subtitle:'Printed cottons, chanderi textures, and easy salwar suit sets inspired by modern Indian wardrobes.',layout:'split',image_url:'/images/rangavali-hero.png',cta_label:'Shop salwar suits',cta_href:'/category/salwar-suits',category_slug:'salwar-suits',product_source:'featured',max_items:8,content:{eyebrow:'THE ETHNICWEAR CHAPTER · 2026'},sort_order:1},
    {id:'category_tiles',section_key:'category_tiles',section_type:'category_tiles',title:'Shop by category',subtitle:'',layout:'tiles',image_url:'',cta_label:'View all',cta_href:'/category/all',category_slug:'',product_source:'categories',max_items:8,content:{},sort_order:2},
    {id:'trending',section_key:'trending',section_type:'product_grid',title:'Trending kurtas and suits',subtitle:'Styles everyone is adding to bag',layout:'grid',image_url:'',cta_label:'View all',cta_href:'/category/all',category_slug:'all',product_source:'trending',max_items:12,content:{},sort_order:3},
    {id:'bestsellers',section_key:'bestsellers',section_type:'product_grid',title:'Bestsellers',subtitle:'Most-loved styles',layout:'grid',image_url:'',cta_label:'View all',cta_href:'/category/all',category_slug:'all',product_source:'bestsellers',max_items:12,content:{},sort_order:4},
    {id:'newsletter',section_key:'newsletter',section_type:'newsletter',title:'₹300 off your first order',subtitle:'Join the store letter for launches and offers.',layout:'centered',image_url:'',cta_label:'Subscribe',cta_href:'',category_slug:'',product_source:'manual',max_items:1,content:{},sort_order:5}
  ],
  paymentMethods:[],
  shippingOptions:[]
};
