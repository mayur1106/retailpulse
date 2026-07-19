import type { Metadata } from 'next';
import { Cormorant_Garamond, Manrope } from 'next/font/google';
import './globals.css';
import { StoreProvider } from '../components/store-provider';
const display=Cormorant_Garamond({subsets:['latin'],variable:'--font-display',weight:['500','600','700']});
const body=Manrope({subsets:['latin'],variable:'--font-body'});
export const metadata:Metadata={title:{default:'Rangavali — Modern Indian Luxury',template:'%s | Rangavali'},description:'Contemporary Indian womenswear, crafted with intention. Discover kurtas, occasion wear and modern classics.',metadataBase:new URL('https://rangavali.example')};
export default function Layout({children}:{children:React.ReactNode}){return <html lang="en"><body className={`${display.variable} ${body.variable}`}><StoreProvider>{children}</StoreProvider></body></html>}
