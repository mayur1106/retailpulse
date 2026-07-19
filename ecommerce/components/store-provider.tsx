'use client';
import {createContext,useContext,useEffect,useMemo,useState} from 'react';
import {Toaster,toast} from 'sonner';

const API_BASE=process.env.NEXT_PUBLIC_API_BASE_URL??'http://localhost:4005';
const STORE_SLUG=process.env.NEXT_PUBLIC_STORE_SLUG??'rangavali';
const CART_TOKEN_KEY='retailpulseCartToken';
const VISITOR_ID_KEY='retailpulseVisitorId';

export type CartItem={productId:string;variantId?:string;cartItemId?:string;quantity?:number};
type Ctx={cart:CartItem[];wish:string[];cartToken:string;add:(id:string,variantId?:string)=>void;toggleWish:(id:string)=>void;remove:(id:string,variantId?:string)=>void;clear:()=>void;markCheckoutStarted:(payload?:{name?:string;email?:string;phone?:string})=>void};
const Store=createContext<Ctx>({cart:[],wish:[],cartToken:'',add:()=>{},toggleWish:()=>{},remove:()=>{},clear:()=>{},markCheckoutStarted:()=>{}});
export const useStore=()=>useContext(Store);

export function StoreProvider({children}:{children:React.ReactNode}){
  const [cart,setCart]=useState<CartItem[]>([]);
  const [wish,setWish]=useState<string[]>([]);
  const [cartToken,setCartToken]=useState('');
  const [visitorId,setVisitorId]=useState('');
  const [ready,setReady]=useState(false);

  useEffect(()=>{bootstrapCart()},[]);
  useEffect(()=>{try{localStorage.rvCart=JSON.stringify(cart);localStorage.rvWish=JSON.stringify(wish)}catch{}},[cart,wish]);

  async function bootstrapCart(){
    try{
      const token=localStorage.getItem(CART_TOKEN_KEY)||`cart_${crypto.randomUUID()}`;
      const visitor=localStorage.getItem(VISITOR_ID_KEY)||`visitor_${crypto.randomUUID()}`;
      localStorage.setItem(CART_TOKEN_KEY,token);
      localStorage.setItem(VISITOR_ID_KEY,visitor);
      setCartToken(token);setVisitorId(visitor);
      const oldCart=readLocalCart();
      setCart(oldCart);
      setWish(JSON.parse(localStorage.rvWish||'[]').map(String));
      const backend=await cartRequest('',{method:'POST',body:JSON.stringify({cartToken:token,visitorId:visitor})});
      if(backend.cart_token&&backend.cart_token!==token){localStorage.setItem(CART_TOKEN_KEY,String(backend.cart_token));setCartToken(String(backend.cart_token))}
      const backendItems=cartFromBackend(backend);
      if(backendItems.length){setCart(backendItems)}
      else if(oldCart.length){
        let latest=backend;
        for(const item of oldCart){latest=await cartRequest('/items',{method:'POST',body:JSON.stringify({cartToken:token,visitorId:visitor,productId:item.productId,variantId:item.variantId,quantity:1})})}
        setCart(cartFromBackend(latest));
      }
    }catch(error){console.error('Backend cart unavailable; using local cart fallback',error);try{setCart(readLocalCart());setWish(JSON.parse(localStorage.rvWish||'[]').map(String))}catch{}}
    finally{setReady(true)}
  }

  async function add(id:string,variantId?:string){
    const optimistic=[...cart,{productId:String(id),variantId}];
    setCart(optimistic);
    toast.success('Added to your bag');
    if(!ready||!cartToken)return;
    try{
      const backend=await cartRequest('/items',{method:'POST',body:JSON.stringify({cartToken,visitorId,productId:String(id),variantId,quantity:1})});
      setCart(cartFromBackend(backend));
    }catch(error){console.error('Could not persist cart item',error)}
  }

  async function remove(id:string,variantId?:string){
    const target=cart.find(item=>item.productId===id&&(variantId?item.variantId===variantId:true));
    setCart(items=>{const index=items.findIndex(item=>item.productId===id&&(variantId?item.variantId===variantId:true));return index<0?items:[...items.slice(0,index),...items.slice(index+1)]});
    if(!target?.cartItemId||!cartToken)return;
    try{
      const backend=await cartRequest(`/items/${target.cartItemId}?cartToken=${encodeURIComponent(cartToken)}`,{method:'DELETE'});
      setCart(cartFromBackend(backend));
    }catch(error){console.error('Could not remove backend cart item',error)}
  }

  async function clear(){
    setCart([]);
    try{
      if(cartToken)await cartRequest(`?cartToken=${encodeURIComponent(cartToken)}`,{method:'DELETE'});
      const next=`cart_${crypto.randomUUID()}`;
      localStorage.setItem(CART_TOKEN_KEY,next);
      localStorage.rvCart='[]';
      setCartToken(next);
    }catch(error){console.error('Could not clear backend cart',error)}
  }

  async function markCheckoutStarted(payload:{name?:string;email?:string;phone?:string}={}){
    if(!cartToken)return;
    try{await cartRequest('/checkout-started',{method:'POST',body:JSON.stringify({cartToken,visitorId,...payload})})}catch(error){console.error('Could not mark checkout started',error)}
  }

  const toggleWish=(id:string)=>{setWish(x=>x.includes(id)?x.filter(a=>a!==id):[...x,id]);toast.success('Wishlist updated')};
  const value=useMemo(()=>({cart,wish,cartToken,add,toggleWish,clear,remove,markCheckoutStarted}),[cart,wish,cartToken,visitorId,ready]);
  return <Store.Provider value={value}>{children}<Toaster position="top-center" richColors/></Store.Provider>
}

function readLocalCart():CartItem[]{try{const raw=JSON.parse(localStorage.rvCart||'[]');return raw.map((x:any)=>typeof x==='string'||typeof x==='number'?{productId:String(x)}:{productId:String(x.productId),variantId:x.variantId?String(x.variantId):undefined,cartItemId:x.cartItemId?String(x.cartItemId):undefined}).filter((x:any)=>x.productId)}catch{return []}}
async function cartRequest(path:string,init:RequestInit){const res=await fetch(`${API_BASE}/v1/storefront/${STORE_SLUG}/cart${path}`,{...init,headers:{'Content-Type':'application/json',...(init.headers??{})}});const data=await res.json().catch(()=>null);if(!res.ok)throw new Error(data?.error?.message||'Cart request failed');return data}
function cartFromBackend(data:any):CartItem[]{const rows=Array.isArray(data?.items)?data.items:[];return rows.flatMap((item:any)=>Array.from({length:Math.max(1,Number(item.quantity??1))},()=>({productId:String(item.productId),variantId:item.variantId?String(item.variantId):undefined,cartItemId:String(item.id),quantity:Number(item.quantity??1)})))}
