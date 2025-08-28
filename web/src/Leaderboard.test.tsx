import { render, screen } from '@testing-library/react'
import React from 'react'
import { describe, it, expect } from 'vitest'
import Leaderboard from './Leaderboard'

describe('Leaderboard', () => {
  it('renders empty state', () => {
    render(<Leaderboard scores={[]} />)
    expect(screen.getByText(/No scores yet/i)).toBeInTheDocument()
  })

  it('renders rows without when column', () => {
    render(<Leaderboard scores={[{ id: 1, name: 'Alice', score: 100 }]} />)
    expect(screen.getByText('#')).toBeInTheDocument()
    expect(screen.getByText('Name')).toBeInTheDocument()
    expect(screen.getByText('Score')).toBeInTheDocument()
    expect(screen.queryByText(/When/i)).not.toBeInTheDocument()
    expect(screen.getByText('Alice')).toBeInTheDocument()
    expect(screen.getByText('100')).toBeInTheDocument()
  })
})
