import {
  BookOpen,
  ShieldCheck,
  GitBranch,
  Settings,
  FileText,
  LayoutGrid,
  CandlestickChart,
  Target,
  BookMarked
} from 'lucide-react'
import type { GuideSection } from '@/types/guide'

/**
 * All guide sections configuration
 * Following the 10-section structure from CLAUDE.md Section 26
 */
export const guideSections: GuideSection[] = [
  {
    id: 'welcome',
    title: 'Welcome',
    icon: BookOpen,
    category: 'basics',
    order: 1,
    description: 'Introduction to ISX Pulse and guide navigation',
    estimatedMinutes: 5,
    prerequisites: []
  },
  {
    id: 'license',
    title: 'License System',
    icon: ShieldCheck,
    category: 'basics',
    order: 2,
    description: 'Understand licensing, activation, and scratch cards',
    estimatedMinutes: 10,
    prerequisites: ['welcome']
  },
  {
    id: 'pipeline',
    title: 'Pipeline Architecture',
    icon: GitBranch,
    category: 'basics',
    order: 3,
    description: '4-stage data processing pipeline explained',
    estimatedMinutes: 15,
    prerequisites: ['welcome']
  },
  {
    id: 'reports',
    title: 'Reports System',
    icon: FileText,
    category: 'features',
    order: 4,
    description: 'Generated reports and data access',
    estimatedMinutes: 15,
    prerequisites: ['pipeline']
  },
  {
    id: 'liquidity-calculations',
    title: 'Liquidity Calculations',
    icon: Settings,
    category: 'features',
    order: 5,
    description: 'Step-by-step liquidity scoring methodology with pros/cons',
    estimatedMinutes: 20,
    prerequisites: ['reports']
  },
  {
    id: 'market-overview',
    title: 'Market Overview',
    icon: LayoutGrid,
    category: 'features',
    order: 6,
    description: 'Interactive treemap visualization',
    estimatedMinutes: 10,
    prerequisites: ['reports']
  },
  {
    id: 'charts',
    title: 'Charts & Analysis',
    icon: CandlestickChart,
    category: 'features',
    order: 7,
    description: 'TradingView-style charting with indicators and drawing tools',
    estimatedMinutes: 20,
    prerequisites: ['reports']
  },
  {
    id: 'strategy',
    title: 'Trading Strategies',
    icon: Target,
    category: 'advanced',
    order: 8,
    description: 'Automated strategy system guide',
    estimatedMinutes: 25,
    prerequisites: ['reports']
  },
  {
    id: 'quick-reference',
    title: 'Quick Reference',
    icon: BookMarked,
    category: 'reference',
    order: 9,
    description: 'Keyboard shortcuts and cheat sheet',
    estimatedMinutes: 5,
    prerequisites: []
  }
]

/**
 * Get section by ID
 */
export function getSectionById(id: string): GuideSection | undefined {
  return guideSections.find(section => section.id === id)
}

/**
 * Get sections by category
 */
export function getSectionsByCategory(category: GuideSection['category']): GuideSection[] {
  return guideSections.filter(section => section.category === category)
}

/**
 * Get next section in sequence
 */
export function getNextSection(currentId: string): GuideSection | undefined {
  const currentSection = getSectionById(currentId)
  if (!currentSection) return undefined

  const nextOrder = currentSection.order + 1
  return guideSections.find(section => section.order === nextOrder)
}

/**
 * Get previous section in sequence
 */
export function getPreviousSection(currentId: string): GuideSection | undefined {
  const currentSection = getSectionById(currentId)
  if (!currentSection) return undefined

  const previousOrder = currentSection.order - 1
  return guideSections.find(section => section.order === previousOrder)
}

/**
 * Check if prerequisites are met
 */
export function arePrerequisitesMet(
  sectionId: string,
  completedSections: string[]
): boolean {
  const section = getSectionById(sectionId)
  if (!section || !section.prerequisites) return true

  return section.prerequisites.every(prereq => completedSections.includes(prereq))
}

/**
 * Get total estimated time for all sections
 */
export function getTotalEstimatedMinutes(): number {
  return guideSections.reduce((total, section) => {
    return total + (section.estimatedMinutes || 0)
  }, 0)
}

/**
 * Get sections grouped by category
 */
export function getSectionsGroupedByCategory() {
  return {
    basics: getSectionsByCategory('basics'),
    features: getSectionsByCategory('features'),
    advanced: getSectionsByCategory('advanced'),
    reference: getSectionsByCategory('reference')
  }
}
