import React from 'react'
import type { Score } from './api'

export default function Leaderboard({ scores }: { scores: Score[] }) {
  if (!scores?.length) return <div>No scores yet. Be the first!</div>
  return (
    <table className="leaderboard">
      <thead>
        <tr>
          <th>#</th>
          <th>Name</th>
          <th>Score</th>
        </tr>
      </thead>
      <tbody>
        {scores.map((s, i) => (
          <tr key={s.id ?? i}>
            <td>{i + 1}</td>
            <td>{s.name}</td>
            <td>{s.score}</td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}
