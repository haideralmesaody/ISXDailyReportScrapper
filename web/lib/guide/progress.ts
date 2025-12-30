'use client'

import { useState, useEffect, useCallback } from 'react'
import type { GuideProgress } from '@/types/guide'

const STORAGE_KEY = 'isx-guide-progress'

/**
 * Default progress state for new users
 */
const defaultProgress: GuideProgress = {
  sectionsCompleted: [],
  lastVisited: 'welcome',
  startedAt: new Date().toISOString(),
  completedDemos: [],
  quizScores: {},
  bookmarks: []
}

/**
 * Hook for managing guide progress in localStorage
 * Follows React Hydration Best Practices from CLAUDE.md
 */
export function useGuideProgress() {
  const [progress, setProgress] = useState<GuideProgress>(defaultProgress)
  const [isHydrated, setIsHydrated] = useState(false)

  // Load progress from localStorage on mount (client-side only)
  useEffect(() => {
    const loadProgress = () => {
      try {
        const stored = localStorage.getItem(STORAGE_KEY)
        if (stored) {
          const parsed = JSON.parse(stored) as GuideProgress
          setProgress(parsed)
        }
      } catch (error) {
        console.error('Failed to load guide progress:', error)
      } finally {
        setIsHydrated(true)
      }
    }

    loadProgress()
  }, [])

  // Save progress to localStorage whenever it changes
  useEffect(() => {
    if (!isHydrated) return

    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(progress))
    } catch (error) {
      console.error('Failed to save guide progress:', error)
    }
  }, [progress, isHydrated])

  /**
   * Mark a section as completed
   */
  const markSectionComplete = useCallback((sectionId: string) => {
    setProgress(prev => ({
      ...prev,
      sectionsCompleted: prev.sectionsCompleted.includes(sectionId)
        ? prev.sectionsCompleted
        : [...prev.sectionsCompleted, sectionId]
    }))
  }, [])

  /**
   * Mark a section as incomplete (for re-visiting)
   */
  const markSectionIncomplete = useCallback((sectionId: string) => {
    setProgress(prev => ({
      ...prev,
      sectionsCompleted: prev.sectionsCompleted.filter(id => id !== sectionId)
    }))
  }, [])

  /**
   * Update last visited section
   */
  const setLastVisited = useCallback((sectionId: string) => {
    setProgress(prev => ({
      ...prev,
      lastVisited: sectionId
    }))
  }, [])

  /**
   * Mark a demo as completed
   */
  const markDemoComplete = useCallback((demoId: string) => {
    setProgress(prev => ({
      ...prev,
      completedDemos: prev.completedDemos.includes(demoId)
        ? prev.completedDemos
        : [...prev.completedDemos, demoId]
    }))
  }, [])

  /**
   * Toggle bookmark for a section
   */
  const toggleBookmark = useCallback((sectionId: string) => {
    setProgress(prev => ({
      ...prev,
      bookmarks: prev.bookmarks.includes(sectionId)
        ? prev.bookmarks.filter(id => id !== sectionId)
        : [...prev.bookmarks, sectionId]
    }))
  }, [])

  /**
   * Save quiz score for a section
   */
  const saveQuizScore = useCallback((sectionId: string, score: number) => {
    setProgress(prev => ({
      ...prev,
      quizScores: {
        ...prev.quizScores,
        [sectionId]: score
      }
    }))
  }, [])

  /**
   * Calculate overall progress percentage
   * Default total: 8 sections (current guide navigation)
   */
  const getProgressPercentage = useCallback((totalSections: number = 8) => {
    return Math.round((progress.sectionsCompleted.length / totalSections) * 100)
  }, [progress.sectionsCompleted])

  /**
   * Reset all progress (for testing or user request)
   */
  const resetProgress = useCallback(() => {
    setProgress({
      ...defaultProgress,
      startedAt: new Date().toISOString()
    })
  }, [])

  /**
   * Export progress as JSON string
   */
  const exportProgress = useCallback(() => {
    return JSON.stringify(progress, null, 2)
  }, [progress])

  /**
   * Import progress from JSON string
   */
  const importProgress = useCallback((jsonString: string) => {
    try {
      const imported = JSON.parse(jsonString) as GuideProgress
      setProgress(imported)
      return true
    } catch (error) {
      console.error('Failed to import progress:', error)
      return false
    }
  }, [])

  return {
    progress,
    isHydrated,
    markSectionComplete,
    markSectionIncomplete,
    setLastVisited,
    markDemoComplete,
    toggleBookmark,
    saveQuizScore,
    getProgressPercentage,
    resetProgress,
    exportProgress,
    importProgress
  }
}

/**
 * Check if a section is completed
 */
export function isSectionCompleted(progress: GuideProgress, sectionId: string): boolean {
  return progress.sectionsCompleted.includes(sectionId)
}

/**
 * Check if a demo is completed
 */
export function isDemoCompleted(progress: GuideProgress, demoId: string): boolean {
  return progress.completedDemos.includes(demoId)
}

/**
 * Check if a section is bookmarked
 */
export function isBookmarked(progress: GuideProgress, sectionId: string): boolean {
  return progress.bookmarks.includes(sectionId)
}

/**
 * Get quiz score for a section
 */
export function getQuizScore(progress: GuideProgress, sectionId: string): number | undefined {
  return progress.quizScores[sectionId]
}
