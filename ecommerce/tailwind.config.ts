import type { Config } from 'tailwindcss';
export default {content:['./app/**/*.{ts,tsx}','./components/**/*.{ts,tsx}'],theme:{extend:{colors:{wine:'#7A1E48',gold:'#CFA15C',blush:'#FFF7F2'},fontFamily:{sans:['var(--font-body)'],serif:['var(--font-display)']},boxShadow:{soft:'0 18px 60px rgba(65,32,44,.10)'}}},plugins:[]} satisfies Config;
