export type StageTypeDefinition = {
  id: string
  name?: string
  description?: string
  dependencies?: string[] | null
  can_run_alone?: boolean
  parameters?: unknown
}

export type PipelineContext = {
  stageOrder: string[]
  totalStages: number
  stageNumberById: Record<string, number>
  dependsOnById: Record<string, string[]>
  stageById: Record<string, StageTypeDefinition>
}

const normalizeDependencies = (
  dependencies: unknown,
): string[] => {
  if (!Array.isArray(dependencies)) return []
  return dependencies.filter((dep): dep is string => typeof dep === 'string' && dep.length > 0)
}

const stableTopologicalSort = (
  nodes: string[],
  edges: Array<[from: string, to: string]>,
  originalIndexById: Record<string, number>,
): string[] | null => {
  const nodeSet = new Set(nodes)
  const indegree: Record<string, number> = {}
  const adjacency: Record<string, string[]> = {}

  for (const id of nodes) {
    indegree[id] = 0
    adjacency[id] = []
  }

  for (const [from, to] of edges) {
    if (!nodeSet.has(from) || !nodeSet.has(to)) continue
    adjacency[from]?.push(to)
    indegree[to] = (indegree[to] ?? 0) + 1
  }

  const zero: string[] = nodes.filter((id) => (indegree[id] ?? 0) === 0)
  zero.sort((a, b) => (originalIndexById[a] ?? 0) - (originalIndexById[b] ?? 0))

  const result: string[] = []
  while (zero.length) {
    const next = zero.shift()!
    result.push(next)

    for (const neighbor of adjacency[next] ?? []) {
      indegree[neighbor] = (indegree[neighbor] ?? 0) - 1
      if ((indegree[neighbor] ?? 0) === 0) {
        zero.push(neighbor)
      }
    }
    zero.sort((a, b) => (originalIndexById[a] ?? 0) - (originalIndexById[b] ?? 0))
  }

  if (result.length !== nodes.length) return null
  return result
}

/**
 * Build a future-proof pipeline context from `/api/operations/types`.
 *
 * - Orders stages based on dependencies (stable topo sort).
 * - Falls back to the backend list order if a cycle is detected.
 */
export function buildPipelineContext(
  stageTypes: StageTypeDefinition[] | undefined,
): PipelineContext | null {
  if (!Array.isArray(stageTypes) || stageTypes.length === 0) return null

  const filtered = stageTypes.filter((t) => t && typeof t.id === 'string' && t.id.length > 0 && t.id !== 'full_pipeline')
  if (filtered.length === 0) return null

  const stageById: Record<string, StageTypeDefinition> = {}
  const originalIndexById: Record<string, number> = {}
  const dependsOnById: Record<string, string[]> = {}

  filtered.forEach((t, index) => {
    stageById[t.id] = t
    originalIndexById[t.id] = index
  })

  const nodes = filtered.map((t) => t.id)
  const edges: Array<[string, string]> = []
  for (const t of filtered) {
    const deps = normalizeDependencies(t.dependencies).filter((dep) => dep in stageById)
    dependsOnById[t.id] = deps
    for (const dep of deps) {
      edges.push([dep, t.id])
    }
  }

  const sorted =
    stableTopologicalSort(nodes, edges, originalIndexById) ??
    nodes

  const stageNumberById: Record<string, number> = {}
  sorted.forEach((id, index) => {
    stageNumberById[id] = index + 1
  })

  return {
    stageOrder: sorted,
    totalStages: sorted.length,
    stageNumberById,
    dependsOnById,
    stageById,
  }
}

