import { describe, it, expect, vi, afterEach } from 'vitest'
import { getTopScores, postScore } from './api'

declare const global: any

afterEach(() => {
  vi.restoreAllMocks()
})

describe('api', () => {
  it('getTopScores returns scores array', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValueOnce({
      ok: true,
      status: 200,
      json: async () => ({ scores: [{ id: 1, name: 'Alice', score: 123 }] })
    })
    const res = await getTopScores(10)
    expect(res).toEqual([{ id: 1, name: 'Alice', score: 123 }])
  })

  it('getTopScores throws on http error', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValueOnce({ ok: false, status: 500 })
    await expect(getTopScores(5)).rejects.toThrow('Failed to fetch scores: 500')
  })

  it('postScore returns created score', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValueOnce({
      ok: true,
      status: 201,
      json: async () => ({ id: 1, name: 'Bob', score: 42 })
    })
    const out = await postScore({ name: 'Bob', score: 42 })
    expect(out).toEqual({ id: 1, name: 'Bob', score: 42 })
  })

  it('postScore throws with API error message', async () => {
    vi.spyOn(global, 'fetch').mockResolvedValueOnce({
      ok: false,
      status: 400,
      json: async () => ({ error: 'invalid' })
    })
    await expect(postScore({ name: 'X', score: 0 })).rejects.toThrow('invalid')
  })
})
