'use client'

import { HelpCircle } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { useRouter } from 'next/navigation'
import { cn } from '@/lib/utils'

interface HelpButtonProps {
  /**
   * Guide section ID to navigate to
   * Examples: 'indicators', 'operations', 'reports', 'liquidity', 'market-overview'
   */
  section: string

  /**
   * Optional subsection or specific indicator to deep link to
   * Example: 'rsi', 'macd', 'bollinger-bands'
   */
  subsection?: string

  /**
   * Tooltip text to show on hover
   */
  tooltip?: string

  /**
   * Optional className for custom styling
   */
  className?: string

  /**
   * Button size variant
   */
  size?: 'default' | 'sm' | 'lg' | 'icon'

  /**
   * Button variant
   */
  variant?: 'default' | 'ghost' | 'outline' | 'secondary' | 'link'
}

/**
 * Context-Sensitive Help Button Component
 *
 * Provides quick access to relevant Interactive Guide sections from any page.
 * Features:
 * - Deep linking to specific guide sections and subsections
 * - Tooltip with helpful description
 * - Consistent icon and styling across app
 * - Opens guide in same tab with returnTo support
 *
 * @example
 * ```tsx
 * // Basic usage - link to indicators section
 * <HelpButton
 *   section="indicators"
 *   tooltip="Learn about technical indicators"
 * />
 *
 * // Deep link to specific indicator
 * <HelpButton
 *   section="indicators"
 *   subsection="rsi"
 *   tooltip="Learn about RSI indicator"
 * />
 *
 * // Custom styling
 * <HelpButton
 *   section="operations"
 *   tooltip="Operations guide"
 *   className="ml-2"
 *   variant="outline"
 *   size="sm"
 * />
 * ```
 */
export function HelpButton({
  section,
  subsection,
  tooltip = 'Get help',
  className,
  size = 'sm',
  variant = 'ghost'
}: HelpButtonProps) {
  const router = useRouter()

  const handleClick = () => {
    // Build guide URL with section and optional subsection
    let guideUrl = `/guide?section=${section}`

    if (subsection) {
      guideUrl += `&subsection=${subsection}`
    }

    // Navigate to guide
    router.push(guideUrl)
  }

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            variant={variant}
            size={size}
            className={cn('gap-2', className)}
            onClick={handleClick}
            type="button"
          >
            <HelpCircle className="h-4 w-4" />
            <span className="sr-only">{tooltip}</span>
          </Button>
        </TooltipTrigger>
        <TooltipContent>
          <p>{tooltip}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}
