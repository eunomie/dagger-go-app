import { describe, it, expect } from 'vitest'
import { newGame, canMove, moveLeft, type Grid } from './game'

function countNonZero(g: Grid): number {
  return g.flat().filter(v => v !== 0).length
}

describe('game logic', () => {
  it('newGame starts with exactly two tiles', () => {
    const g = newGame()
    const nz = countNonZero(g)
    expect(nz).toBe(2)
  })

  it('canMove false when board is full with no merges', () => {
    // pattern without adjacent equals
    const g: Grid = [
      [2, 4, 2, 4],
      [4, 2, 4, 2],
      [2, 4, 2, 4],
      [4, 2, 4, 2],
    ]
    expect(canMove(g)).toBe(false)
  })

  it('moveLeft merges pairs correctly', () => {
    const g: Grid = [
      [2, 2, 4, 4],
      [0, 0, 0, 0],
      [0, 0, 0, 0],
      [0, 0, 0, 0],
    ]
    const { grid: next, moved, gained } = moveLeft(g)
    expect(moved).toBe(true)
    expect(gained).toBe(4 + 8)
    expect(next[0]).toEqual([4, 8, 0, 0])
  })
})
