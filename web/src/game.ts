export type Grid = number[][]

const SIZE = 4

export function newGame(): Grid {
  let g = emptyGrid()
  g = addRandomTile(g)
  g = addRandomTile(g)
  return g
}

export function emptyGrid(): Grid {
  return Array.from({ length: SIZE }, () => Array.from({ length: SIZE }, () => 0))
}

export function addRandomTile(grid: Grid): Grid {
  const empties: [number, number][] = []
  for (let r = 0; r < SIZE; r++) {
    for (let c = 0; c < SIZE; c++) {
      if (grid[r][c] === 0) empties.push([r, c])
    }
  }
  if (empties.length === 0) return grid
  const [r, c] = empties[Math.floor(Math.random() * empties.length)]
  const v = Math.random() < 0.9 ? 2 : 4
  const next = clone(grid)
  next[r][c] = v
  return next
}

export function canMove(grid: Grid): boolean {
  for (let r = 0; r < SIZE; r++) {
    for (let c = 0; c < SIZE; c++) {
      if (grid[r][c] === 0) return true
      if (r + 1 < SIZE && grid[r][c] === grid[r + 1][c]) return true
      if (c + 1 < SIZE && grid[r][c] === grid[r][c + 1]) return true
    }
  }
  return false
}

function clone(grid: Grid): Grid { return grid.map(row => row.slice()) }

function slideAndMerge(line: number[]): { line: number[]; gained: number; moved: boolean } {
  const filtered = line.filter(v => v !== 0)
  const merged: number[] = []
  let gained = 0
  for (let i = 0; i < filtered.length; i++) {
    if (i + 1 < filtered.length && filtered[i] === filtered[i + 1]) {
      const sum = filtered[i] * 2
      merged.push(sum)
      gained += sum
      i++
    } else {
      merged.push(filtered[i])
    }
  }
  while (merged.length < SIZE) merged.push(0)
  const moved = !arraysEqual(line, merged)
  return { line: merged, gained, moved }
}

function arraysEqual(a: number[], b: number[]): boolean {
  if (a.length !== b.length) return false
  for (let i = 0; i < a.length; i++) if (a[i] !== b[i]) return false
  return true
}

export function moveLeft(grid: Grid): { grid: Grid; moved: boolean; gained: number } {
  const next = emptyGrid()
  let moved = false
  let gained = 0
  for (let r = 0; r < SIZE; r++) {
    const { line, gained: g, moved: m } = slideAndMerge(grid[r])
    next[r] = line
    moved = moved || m
    gained += g
  }
  return { grid: next, moved, gained }
}

export function moveRight(grid: Grid): { grid: Grid; moved: boolean; gained: number } {
  const next = emptyGrid()
  let moved = false
  let gained = 0
  for (let r = 0; r < SIZE; r++) {
    const reversed = grid[r].slice().reverse()
    const { line, gained: g, moved: m } = slideAndMerge(reversed)
    next[r] = line.reverse()
    moved = moved || m
    gained += g
  }
  return { grid: next, moved, gained }
}

export function moveUp(grid: Grid): { grid: Grid; moved: boolean; gained: number } {
  const next = clone(grid)
  let moved = false
  let gained = 0
  for (let c = 0; c < SIZE; c++) {
    const col = Array.from({ length: SIZE }, (_, r) => grid[r][c])
    const { line, gained: g, moved: m } = slideAndMerge(col)
    for (let r = 0; r < SIZE; r++) next[r][c] = line[r]
    moved = moved || m
    gained += g
  }
  return { grid: next, moved, gained }
}

export function moveDown(grid: Grid): { grid: Grid; moved: boolean; gained: number } {
  const next = clone(grid)
  let moved = false
  let gained = 0
  for (let c = 0; c < SIZE; c++) {
    const col = Array.from({ length: SIZE }, (_, r) => grid[r][c]).reverse()
    const { line, gained: g, moved: m } = slideAndMerge(col)
    const rev = line.reverse()
    for (let r = 0; r < SIZE; r++) next[r][c] = rev[r]
    moved = moved || m
    gained += g
  }
  return { grid: next, moved, gained }
}
