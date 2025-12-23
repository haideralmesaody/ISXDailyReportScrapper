import dynamic from 'next/dynamic'
import { cn } from '@/lib/utils'

// Dynamic import of client components with proper loading states
const InvestorLogoClient = dynamic(
  () => import('./investor-logo-client').then(mod => ({ default: mod.InvestorLogoClient })),
  {
    ssr: false,
    loading: () => <InvestorLogoSkeleton />
  }
)

const InvestorLogoCompactClient = dynamic(
  () => import('./investor-logo-client').then(mod => ({ default: mod.InvestorLogoCompactClient })),
  {
    ssr: false,
    loading: () => <InvestorLogoSkeleton variant="compact" />
  }
)

const InvestorIconClient = dynamic(
  () => import('./investor-logo-client').then(mod => ({ default: mod.InvestorIconClient })),
  {
    ssr: false,
    loading: () => <InvestorLogoSkeleton variant="icon-only" />
  }
)

const InvestorHeaderLogoClient = dynamic(
  () => import('./investor-logo-client').then(mod => ({ default: mod.InvestorHeaderLogoClient })),
  {
    ssr: false,
    loading: () => <InvestorLogoSkeleton size="xl" />
  }
)

const InvestorFaviconClient = dynamic(
  () => import('./investor-logo-client').then(mod => ({ default: mod.InvestorFaviconClient })),
  {
    ssr: false,
    loading: () => <InvestorFaviconSkeleton />
  }
)

interface InvestorLogoProps {
  className?: string
  size?: 'sm' | 'md' | 'lg' | 'xl' | '2xl'
  showText?: boolean
  variant?: 'full' | 'compact' | 'icon-only'
}

// Loading skeleton component
function InvestorLogoSkeleton({ 
  className,
  size = 'md',
  variant = 'full'
}: {
  className?: string
  size?: 'sm' | 'md' | 'lg' | 'xl' | '2xl'
  variant?: 'full' | 'compact' | 'icon-only'
}) {
  const sizeClasses = {
    sm: 'h-6 w-6',
    md: 'h-10 w-10',
    lg: 'h-14 w-14',
    xl: 'h-20 w-20',
    '2xl': 'h-28 w-28'
  }

  const showText = variant !== 'icon-only'

  return (
    <div className={cn(
      "flex items-center",
      variant === 'compact' ? 'gap-2' : 'gap-3',
      className
    )}>
      <div className={cn(
        "bg-muted animate-pulse rounded-lg",
        sizeClasses[size]
      )} />
      
      {showText && (
        <div className="flex flex-col gap-1">
          <div className={cn(
            "bg-muted animate-pulse rounded h-4",
            size === 'sm' ? 'w-16' : size === 'md' ? 'w-20' : 'w-24'
          )} />
          {size !== 'sm' && variant !== 'compact' && (
            <div className="bg-muted animate-pulse rounded h-3 w-32" />
          )}
        </div>
      )}
    </div>
  )
}

// Favicon skeleton
function InvestorFaviconSkeleton({ 
  className,
  size = 'sm'
}: {
  className?: string
  size?: 'sm' | 'md'
}) {
  const sizeClasses = {
    sm: 'h-4 w-4',
    md: 'h-6 w-6'
  }

  return (
    <div className={cn('relative', className)}>
      <div className={cn(
        "bg-muted animate-pulse rounded",
        sizeClasses[size]
      )} />
    </div>
  )
}

export function InvestorLogo(props: InvestorLogoProps) {
  return <InvestorLogoClient {...props} />
}

// Compact version for smaller spaces
export function InvestorLogoCompact({ 
  className,
  size = 'sm'
}: { 
  className?: string
  size?: 'sm' | 'md'
}) {
  return (
    <InvestorLogoCompactClient 
      {...(className && { className })}
      size={size}
    />
  )
}

// Icon only version
export function InvestorIcon({ 
  className, 
  size = 'md' 
}: { 
  className?: string
  size?: 'sm' | 'md' | 'lg' | 'xl' | '2xl'
}) {
  return (
    <InvestorIconClient 
      {...(className && { className })}
      size={size}
    />
  )
}

// Professional header logo for prominent display
export function InvestorHeaderLogo({ 
  className,
  size = 'xl'
}: { 
  className?: string
  size?: 'lg' | 'xl' | '2xl'
}) {
  return (
    <InvestorHeaderLogoClient 
      {...(className && { className })}
      size={size}
    />
  )
}

// Favicon-style logo for browser tabs and small displays
export function InvestorFavicon({ 
  className,
  size = 'sm'
}: { 
  className?: string
  size?: 'sm' | 'md'
}) {
  return (
    <InvestorFaviconClient 
      {...(className && { className })}
      size={size}
    />
  )
}