import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { addRandomTile, canMove, moveDown, moveLeft, moveRight, moveUp, newGame, type Grid } from './game'
import { getTopScores, postScore, type Score } from './api'
import Leaderboard from './Leaderboard'

function useKeyboard(handler: (key: string) => void) {
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      const keys = ['ArrowUp', 'ArrowDown', 'ArrowLeft', 'ArrowRight']
      if (keys.includes(e.key)) {
        e.preventDefault()
        handler(e.key)
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [handler])
}

export default function App() {
  const [grid, setGrid] = useState<Grid>(newGame())
  const [score, setScore] = useState(0)
  const [gameOver, setGameOver] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [submitted, setSubmitted] = useState(false)
  const [name, setName] = useState('')
  const [scores, setScores] = useState<Score[]>([])
  const [error, setError] = useState<string | null>(null)

  const startNew = useCallback(() => {
    const g = newGame()
    setGrid(g)
    setScore(0)
    setGameOver(!canMove(g))
    setSubmitted(false)
    setName('')
  }, [])

  const handleMove = useCallback((key: string) => {
    if (gameOver) return
    let moved = false
    let gained = 0
    let next = grid
    switch (key) {
      case 'ArrowUp': ({ grid: next, moved, gained } = moveUp(grid)); break
      case 'ArrowDown': ({ grid: next, moved, gained } = moveDown(grid)); break
      case 'ArrowLeft': ({ grid: next, moved, gained } = moveLeft(grid)); break
      case 'ArrowRight': ({ grid: next, moved, gained } = moveRight(grid)); break
    }
    if (moved) {
      next = addRandomTile(next)
      setGrid(next)
      setScore(s => s + gained)
      if (!canMove(next)) {
        setGameOver(true)
      }
    } else {
      // No tiles moved; if there are no possible moves, the game is over.
      if (!canMove(grid)) {
        setGameOver(true)
      }
    }
  }, [grid, gameOver])

  useKeyboard(handleMove)

  useEffect(() => {
    getTopScores(10).then(setScores).catch(() => {})
  }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (submitted || submitting) return
    const trimmed = name.trim()
    if (!trimmed) return
    setSubmitting(true)
    setError(null)
    try {
      await postScore({ name: trimmed, score })
      setSubmitted(true)
      const list = await getTopScores(10)
      setScores(list)
    } catch (err: any) {
      setError(err?.message ?? 'Failed to submit score')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="app">
      <header>
        <h1>2048</h1>
        <div className="score">Score: {score}</div>
        <button onClick={startNew}>New Game</button>
      </header>
      <Board grid={grid} />
      {gameOver && (
        <div className="game-over">
          <div>Game Over! Your score: {score}</div>
          {!submitted ? (
            <form onSubmit={handleSubmit} className="submit-form">
              <input
                type="text"
                placeholder="Your name"
                value={name}
                onChange={e => setName(e.target.value)}
                maxLength={50}
                required
              />
              <button type="submit" disabled={submitting}>Submit Score</button>
            </form>
          ) : (
            <div className="submitted-msg">Score submitted! Thank you.</div>
          )}
          {error && <div className="error">{error}</div>}
        </div>
      )}

      <section>
        <h2>Leaderboard</h2>
        <Leaderboard scores={scores} />
      </section>
      <footer>
        <small>Use arrow keys to play.</small>
      </footer>
    </div>
  )
}

function Board({ grid }: { grid: Grid }) {
  const tiles = useMemo(() => grid.flat(), [grid])
  return (
    <div className="board">
      {tiles.map((v, i) => (
        <div key={i} className={`tile tile-${v || 'empty'}`}>{v || ''}</div>
      ))}
    </div>
  )
}
