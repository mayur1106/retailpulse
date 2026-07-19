"use client";

import Link from "next/link";
import { useMutation } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { apiRequest, AuthResponse, saveTokens } from "@/lib/api";

const schema = z.object({
  organizationName: z.string().min(1),
  name: z.string().min(1),
  email: z.string().email(),
  password: z.string().min(12),
  accountType: z.enum(["seller", "owner"]),
});

type FormValues = z.infer<typeof schema>;

export default function RegisterPage() {
  const form = useForm<FormValues>({ resolver: zodResolver(schema), defaultValues: { accountType: "seller" } });
  const mutation = useMutation({
    mutationFn: (values: FormValues) => apiRequest<AuthResponse>("/v1/auth/register", values),
    onSuccess: (data) => {
      saveTokens(data.tokens);
      window.location.href = "/dashboard";
    },
  });

  return (
    <main className="flex min-h-screen items-center justify-center bg-[#f6f8fb] px-6 text-[#17202a]">
      <form className="w-full max-w-md rounded-2xl border border-[#e4e7ec] bg-white p-6 shadow-xl shadow-slate-200/70" onSubmit={form.handleSubmit((values) => mutation.mutate(values))}>
        <h1 className="text-2xl font-semibold tracking-tight text-[#101828]">Create your workspace</h1>
        <div className="mt-6 grid gap-4">
          <div className="grid grid-cols-2 gap-2 rounded-xl border border-[#d0d5dd] bg-[#f9fafb] p-1">
            {(["seller", "owner"] as const).map((type) => (
              <label key={type} className={`cursor-pointer rounded-lg px-3 py-2 text-center text-sm font-semibold ${form.watch("accountType") === type ? "bg-[#0f766e] text-white shadow-sm" : "text-[#667085]"}`}>
                <input className="sr-only" type="radio" value={type} {...form.register("accountType")} />
                {type === "seller" ? "Seller" : "Owner"}
              </label>
            ))}
          </div>
          <input className="rounded-xl border border-[#d0d5dd] bg-white px-3 py-3 text-[#344054] outline-none placeholder:text-[#98a2b3] focus:border-[#14b8a6] focus:ring-4 focus:ring-[#ccfbf1]" placeholder="Organization name" {...form.register("organizationName")} />
          <input className="rounded-xl border border-[#d0d5dd] bg-white px-3 py-3 text-[#344054] outline-none placeholder:text-[#98a2b3] focus:border-[#14b8a6] focus:ring-4 focus:ring-[#ccfbf1]" placeholder="Your name" {...form.register("name")} />
          <input className="rounded-xl border border-[#d0d5dd] bg-white px-3 py-3 text-[#344054] outline-none placeholder:text-[#98a2b3] focus:border-[#14b8a6] focus:ring-4 focus:ring-[#ccfbf1]" placeholder="Email" type="email" {...form.register("email")} />
          <input className="rounded-xl border border-[#d0d5dd] bg-white px-3 py-3 text-[#344054] outline-none placeholder:text-[#98a2b3] focus:border-[#14b8a6] focus:ring-4 focus:ring-[#ccfbf1]" placeholder="Password" type="password" {...form.register("password")} />
        </div>
        {mutation.error ? <p className="mt-4 text-sm text-[#b91c1c]">{mutation.error.message}</p> : null}
        <button className="mt-6 w-full rounded-md bg-[#0f766e] px-4 py-3 font-semibold text-white" disabled={mutation.isPending}>
          {mutation.isPending ? "Creating..." : "Create account"}
        </button>
        <Link className="mt-4 block text-sm text-[#0f766e]" href="/login">
          Already have an account?
        </Link>
      </form>
    </main>
  );
}
