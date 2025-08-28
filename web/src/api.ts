export type Score = {
  id?: number
  name: string
  score: number
  created_at?: string
}

const API_BASE = '' // same origin; Vite dev proxies /api to :8080

export async function getTopScores(limit = 10): Promise<Score[]> {
  const res = await fetch(`/api/scores?limit=${encodeURIComponent(String(limit))}`)
  if (!res.ok) throw new Error(`Failed to fetch scores: ${res.status}`)
  const data = await res.json()
  return data.scores ?? []
}

export async function postScore(s: { name: string; score: number }): Promise<Score> {
  const res = await fetch('/api/scores', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(s)
  })
  if (!res.ok) {
    let msg = `Failed to submit: ${res.status}`
    try { const e = await res.json(); if (e?.error) msg = e.error } catch {}
    throw new Error(msg)
  }
  return res.json()
}
