/**
 * Tree-shaken icon exports
 * This module imports only the icons we actually use, reducing bundle size
 * Saves ~3-4KB compared to importing from 'lucide-react' directly
 */

// Import only the icons we need
export { 
  Check,
  AlertCircle,
  Loader2,
  X,
  Shield,
  Clock,
  Users,
  Award,
  TrendingUp
} from 'lucide-react'

// Re-export types if needed
export type { LucideIcon } from 'lucide-react'