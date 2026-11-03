import { ssdUrl } from "./env";

export type Interval = { incoming: number | null; outgoing: number | null };
export type DaySpec = Interval[] | "empty" | "error";

export async function resetSsd(days: Record<string, DaySpec> = {}, fallback: DaySpec = "empty"): Promise<void> {
  await fetch(`${ssdUrl}/__control/reset`, { method: "POST" });
  await fetch(`${ssdUrl}/__control/scenario`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ days, fallback }),
  });
}

export async function ssdRequests(): Promise<{ date: string; pod: string; from: string; to: string }[]> {
  const res = await fetch(`${ssdUrl}/__control/requests`);
  return (await res.json()).requests;
}
