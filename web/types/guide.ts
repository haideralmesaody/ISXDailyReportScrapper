import { LucideIcon } from 'lucide-react'

/**
 * Progress tracking for guide sections and demos
 * Stored in localStorage for persistence
 */
export interface GuideProgress {
  sectionsCompleted: string[]      // Section IDs that user has completed
  lastVisited: string               // Last visited section ID
  startedAt: string                 // ISO timestamp when guide was first accessed
  completedDemos: string[]          // Demo IDs that user has completed
  quizScores: Record<string, number> // Section ID → score (for future quiz feature)
  bookmarks: string[]               // Bookmarked section IDs
}

/**
 * Section metadata and configuration
 */
export interface GuideSection {
  id: string                        // Unique identifier (e.g., 'welcome', 'license')
  title: string                     // Display title
  icon: LucideIcon                  // Lucide icon component
  category: 'basics' | 'features' | 'advanced' | 'reference'
  order: number                     // Display order in navigation
  description: string               // Short description for tooltips
  estimatedMinutes?: number         // Estimated completion time
  prerequisites?: string[]          // Section IDs that should be completed first
  hidden?: boolean                  // Hide from navigation (for sub-pages like individual indicators)
}

/**
 * Interactive demo configuration
 */
export interface InteractiveDemo {
  id: string                        // Unique identifier
  title: string                     // Demo title
  description: string               // Demo description
  sectionId: string                 // Parent section ID
  interactive: boolean              // Whether demo requires user interaction
  codeExample?: string              // Optional code snippet
}

/**
 * Code example configuration for syntax highlighting
 */
export interface CodeExample {
  language: 'typescript' | 'javascript' | 'bash' | 'json' | 'go'
  code: string
  title?: string
  highlightLines?: number[]         // Line numbers to highlight
  filename?: string                 // Optional filename to display
}

/**
 * Navigation state for guide
 */
export interface GuideNavigationState {
  currentSection: string
  isMobileMenuOpen: boolean
  isCollapsed: boolean              // For tablet sidebar
}
