import { Loader2 } from 'lucide-react'

export default function Loading() {
  return (
    <div className="min-h-screen py-10 bg-background flex items-center justify-center">
      <div
        className="text-center space-y-4 max-w-md"
        role="status"
      >
        {/* Screen-reader only label */}
        <span className="sr-only">Loading license page…</span>

        <h2 className="text-2xl font-medium">
          Loading License Page…
        </h2>

        <Loader2
          className="mx-auto h-8 w-8 text-primary motion-safe:animate-spin motion-reduce:hidden"
          aria-hidden="true"
        />

        <p className="text-muted-foreground/70 md:text-muted-foreground/80 leading-relaxed motion-safe:animate-pulse">
          Please wait while we prepare the license activation page.
        </p>
      </div>
    </div>
  )
}