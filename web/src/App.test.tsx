import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach } from 'vitest'

// Mock API module
vi.mock('./api', () => ({
  getTopScores: vi.fn().mockResolvedValue([{ id: 1, name: 'Alice', score: 100 }]),
  postScore: vi.fn().mockResolvedValue({ id: 2, name: 'Bob', score: 50 })
}))

// Prepare a dead-end grid with no possible moves
const dead: number[][] = [
  [2, 4, 2, 4],
  [4, 2, 4, 2],
  [2, 4, 2, 4],
  [4, 2, 4, 2],
]

// Mock game logic so that any move does not change the grid and there are no moves available
vi.mock('./game', () => ({
  newGame: () => dead,
  addRandomTile: (g: number[][]) => g,
  canMove: () => false,
  moveUp: (_g: number[][]) => ({ grid: dead, moved: true, gained: 0 }),
  moveDown: (_g: number[][]) => ({ grid: dead, moved: true, gained: 0 }),
  moveLeft: (_g: number[][]) => ({ grid: dead, moved: true, gained: 0 }),
  moveRight: (_g: number[][]) => ({ grid: dead, moved: true, gained: 0 })
}))

import App from './App'
import { getTopScores, postScore } from './api'

describe('App integration', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('loads leaderboard on mount and shows rows', async () => {
    render(<App />)
    expect(await screen.findByText('Leaderboard')).toBeInTheDocument()
    await waitFor(() => {
      expect(getTopScores).toHaveBeenCalled()
    })
    expect(screen.getByText('Alice')).toBeInTheDocument()
    expect(screen.getByText('100')).toBeInTheDocument()
  })

  it('shows Game Over and submit form when a move is not possible', async () => {
    render(<App />)
    await waitFor(() => {
      expect(getTopScores).toHaveBeenCalled()
    })
    // Start new game with a dead grid (mocked), which should set gameOver immediately
    const user = userEvent.setup()
    const btn = screen.getAllByRole('button', { name: /New Game/i })[0]
    await user.click(btn)
    expect(await screen.findByText(/Game Over!/i)).toBeInTheDocument()
    expect(screen.getByPlaceholderText('Your name')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Submit Score/i })).toBeInTheDocument()
  })
})


it('submits score only once and hides the form after submission', async () => {
  render(<App />)
  await waitFor(() => {
    expect(getTopScores).toHaveBeenCalled()
  })
  const user = userEvent.setup()
  const btn = screen.getAllByRole('button', { name: /New Game/i })[0]
  await user.click(btn)

  // Form is visible
  const input = await screen.findByPlaceholderText('Your name')
  await user.type(input, 'Charlie')
  const submit = screen.getByRole('button', { name: /Submit Score/i })
  await user.click(submit)

  // After submission, confirmation is shown and form is hidden
  expect(await screen.findByText(/Score submitted! Thank you\./i)).toBeInTheDocument()
  expect(screen.queryByPlaceholderText('Your name')).not.toBeInTheDocument()
  expect(screen.queryByRole('button', { name: /Submit Score/i })).not.toBeInTheDocument()

  // Ensure API was called exactly once
  expect(postScore).toHaveBeenCalledTimes(1)
})
