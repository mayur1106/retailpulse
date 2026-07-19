import Link from "next/link";
import { Activity, BarChart3, ShieldCheck } from "lucide-react";

export default function HomePage() {
  return (
    <main className="min-h-screen bg-[#f6f8fb] text-[#17202a]">
      <section className="mx-auto flex min-h-screen max-w-6xl flex-col justify-between px-6 py-8">
        <nav className="flex items-center justify-between">
          <div className="flex items-center gap-3 text-lg font-semibold">
            <Activity className="h-6 w-6 text-[#0f766e]" aria-hidden />
            RetailPulse AI
          </div>
          <div className="flex gap-3">
            <Link className="rounded-md border border-[#c8ced8] px-4 py-2 text-sm" href="/login">
              Sign in
            </Link>
            <Link className="rounded-md bg-[#0f766e] px-4 py-2 text-sm font-semibold text-white" href="/register">
              Create account
            </Link>
          </div>
        </nav>
        <div className="grid gap-10 py-16 lg:grid-cols-[1.05fr_0.95fr] lg:items-center">
          <div>
            <h1 className="max-w-3xl text-5xl font-semibold leading-tight tracking-normal md:text-6xl">
              RetailPulse AI
            </h1>
            <p className="mt-5 max-w-2xl text-lg leading-8 text-[#667085]">
              A secure analytics operating system for Amazon sellers, starting with tenant-safe authentication and the data foundation for SP-API synchronization.
            </p>
            <div className="mt-8 flex flex-wrap gap-3">
              <Link className="rounded-md bg-[#0f766e] px-5 py-3 font-semibold text-white" href="/register">
                Start Phase 1
              </Link>
              <Link className="rounded-md border border-[#c8ced8] px-5 py-3 font-semibold" href="/login">
                Sign in
              </Link>
            </div>
          </div>
          <div className="grid gap-4">
            {[
              ["Tenant isolation", ShieldCheck],
              ["Analytics-ready schema", BarChart3],
              ["OAuth-only Amazon path", Activity],
            ].map(([label, Icon]) => (
              <div key={label as string} className="rounded-2xl border border-[#e4e7ec] bg-white p-5 shadow-sm shadow-slate-200/70">
                <Icon className="h-6 w-6 text-[#b45309]" aria-hidden />
                <div className="mt-3 text-lg font-semibold">{label as string}</div>
              </div>
            ))}
          </div>
        </div>
      </section>
    </main>
  );
}
