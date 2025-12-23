'use client'

import { useState, useRef, useEffect, useCallback } from 'react'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Progress } from '@/components/ui/progress'
import { Copy, RotateCcw, CheckCircle2 } from 'lucide-react'
import { useToast } from '@/lib/hooks/use-toast'

/**
 * Interactive Scratch Card Demo
 * Canvas-based scratch-off effect to reveal license key
 * Supports both mouse and touch interactions
 */
export function ScratchCardDemo() {
  const [revealed, setRevealed] = useState(0)
  const [copied, setCopied] = useState(false)
  const [licenseKey] = useState('ISX-DEMO-2025-ABCD-1234')
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const isDrawing = useRef(false)
  const { toast } = useToast()

  // Initialize canvas with gray overlay
  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return

    const ctx = canvas.getContext('2d')
    if (!ctx) return

    // Set canvas size
    const rect = canvas.getBoundingClientRect()
    canvas.width = rect.width
    canvas.height = rect.height

    // Draw gray background covering
    ctx.fillStyle = '#CCCCCC'
    ctx.fillRect(0, 0, canvas.width, canvas.height)

    // Add texture pattern for realism
    ctx.fillStyle = '#AAAAAA'
    for (let i = 0; i < 200; i++) {
      const x = Math.random() * canvas.width
      const y = Math.random() * canvas.height
      const size = Math.random() * 3
      ctx.fillRect(x, y, size, size)
    }

    // Add hint text
    ctx.fillStyle = '#666666'
    ctx.font = 'bold 20px Arial'
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'
    ctx.fillText('Scratch to Reveal License Key', canvas.width / 2, canvas.height / 2)

    // Add subtle border effect
    ctx.strokeStyle = '#999999'
    ctx.lineWidth = 2
    ctx.strokeRect(1, 1, canvas.width - 2, canvas.height - 2)
  }, [])

  // Calculate revealed percentage
  const calculateRevealed = useCallback(() => {
    const canvas = canvasRef.current
    if (!canvas) return 0

    const ctx = canvas.getContext('2d')
    if (!ctx) return 0

    const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height)
    const pixels = imageData.data
    let transparent = 0

    // Count transparent pixels (alpha = 0)
    for (let i = 3; i < pixels.length; i += 4) {
      if (pixels[i] === 0) transparent++
    }

    const percentage = (transparent / (pixels.length / 4)) * 100
    return Math.min(percentage, 100)
  }, [])

  // Scratch effect function
  const scratch = useCallback((x: number, y: number) => {
    const canvas = canvasRef.current
    if (!canvas) return

    const ctx = canvas.getContext('2d')
    if (!ctx) return

    const rect = canvas.getBoundingClientRect()
    const scaleX = canvas.width / rect.width
    const scaleY = canvas.height / rect.height

    const canvasX = (x - rect.left) * scaleX
    const canvasY = (y - rect.top) * scaleY

    // Use destination-out to erase (create transparency)
    ctx.globalCompositeOperation = 'destination-out'
    ctx.beginPath()
    ctx.arc(canvasX, canvasY, 25, 0, Math.PI * 2)
    ctx.fill()

    // Restore normal composition
    ctx.globalCompositeOperation = 'source-over'

    // Update revealed percentage
    const percentage = calculateRevealed()
    setRevealed(percentage)
  }, [calculateRevealed])

  // Mouse event handlers
  const handleMouseDown = () => {
    isDrawing.current = true
  }

  const handleMouseUp = () => {
    isDrawing.current = false
  }

  const handleMouseMove = (e: React.MouseEvent<HTMLCanvasElement>) => {
    if (!isDrawing.current) return
    scratch(e.clientX, e.clientY)
  }

  // Touch event handlers
  const handleTouchStart = () => {
    isDrawing.current = true
  }

  const handleTouchEnd = () => {
    isDrawing.current = false
  }

  const handleTouchMove = (e: React.TouchEvent<HTMLCanvasElement>) => {
    e.preventDefault() // Prevent scrolling
    if (!isDrawing.current || e.touches.length === 0) return

    const touch = e.touches[0]
    scratch(touch.clientX, touch.clientY)
  }

  // Reset canvas
  const handleReset = () => {
    setRevealed(0)
    setCopied(false)

    const canvas = canvasRef.current
    if (!canvas) return

    const ctx = canvas.getContext('2d')
    if (!ctx) return

    // Clear and redraw
    ctx.clearRect(0, 0, canvas.width, canvas.height)

    // Redraw gray covering
    ctx.fillStyle = '#CCCCCC'
    ctx.fillRect(0, 0, canvas.width, canvas.height)

    // Redraw texture
    ctx.fillStyle = '#AAAAAA'
    for (let i = 0; i < 200; i++) {
      const x = Math.random() * canvas.width
      const y = Math.random() * canvas.height
      const size = Math.random() * 3
      ctx.fillRect(x, y, size, size)
    }

    // Redraw hint text
    ctx.fillStyle = '#666666'
    ctx.font = 'bold 20px Arial'
    ctx.textAlign = 'center'
    ctx.textBaseline = 'middle'
    ctx.fillText('Scratch to Reveal License Key', canvas.width / 2, canvas.height / 2)

    // Redraw border
    ctx.strokeStyle = '#999999'
    ctx.lineWidth = 2
    ctx.strokeRect(1, 1, canvas.width - 2, canvas.height - 2)
  }

  // Copy license key to clipboard
  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(licenseKey)
      setCopied(true)
      toast({
        title: 'License Key Copied!',
        description: 'The license key has been copied to your clipboard.',
      })
      setTimeout(() => setCopied(false), 3000)
    } catch (error) {
      toast({
        title: 'Copy Failed',
        description: 'Failed to copy license key. Please copy manually.',
        variant: 'destructive'
      })
    }
  }

  return (
    <div className="space-y-6">
      {/* Canvas Container */}
      <div className="relative inline-block w-full max-w-2xl">
        {/* Hidden license key (background layer) */}
        <div className="absolute inset-0 flex items-center justify-center bg-gradient-to-br from-green-50 to-green-100 dark:from-green-950 dark:to-green-900 border-2 border-green-300 dark:border-green-700 rounded-lg">
          <div className="text-center space-y-2">
            <div className="text-xs font-semibold text-green-600 dark:text-green-400 uppercase tracking-wider">
              License Key
            </div>
            <div className="text-2xl md:text-3xl font-mono font-bold text-green-700 dark:text-green-300">
              {licenseKey}
            </div>
            <div className="text-xs text-green-600 dark:text-green-400">
              Valid for 1 Year
            </div>
          </div>
        </div>

        {/* Canvas overlay (scratch layer) */}
        <canvas
          ref={canvasRef}
          className="relative w-full h-40 md:h-48 rounded-lg cursor-crosshair touch-none border-2 border-gray-300 dark:border-gray-700"
          onMouseDown={handleMouseDown}
          onMouseUp={handleMouseUp}
          onMouseLeave={handleMouseUp}
          onMouseMove={handleMouseMove}
          onTouchStart={handleTouchStart}
          onTouchEnd={handleTouchEnd}
          onTouchMove={handleTouchMove}
        />
      </div>

      {/* Progress Indicator */}
      <div className="space-y-2">
        <div className="flex items-center justify-between text-sm">
          <span className="font-medium">Revealed</span>
          <span className="text-muted-foreground">{revealed.toFixed(0)}%</span>
        </div>
        <Progress value={revealed} className="h-2" />
        {revealed > 0 && revealed < 50 && (
          <p className="text-xs text-muted-foreground">
            Keep scratching to reveal the full license key...
          </p>
        )}
      </div>

      {/* Action Buttons */}
      <div className="flex flex-wrap items-center gap-3">
        {revealed >= 50 && (
          <Button
            onClick={handleCopy}
            className="gap-2"
            variant={copied ? "outline" : "default"}
          >
            {copied ? (
              <>
                <CheckCircle2 className="h-4 w-4" />
                Copied!
              </>
            ) : (
              <>
                <Copy className="h-4 w-4" />
                Copy License Key
              </>
            )}
          </Button>
        )}

        <Button
          onClick={handleReset}
          variant="outline"
          className="gap-2"
        >
          <RotateCcw className="h-4 w-4" />
          Reset Card
        </Button>
      </div>

      {/* Instructions */}
      <Card className="bg-muted/50">
        <div className="p-4">
          <h4 className="font-semibold mb-2 text-sm">How It Works:</h4>
          <ul className="space-y-1 text-xs text-muted-foreground">
            <li>• <strong>Desktop:</strong> Click and drag your mouse to scratch</li>
            <li>• <strong>Mobile:</strong> Use your finger to scratch the surface</li>
            <li>• <strong>Copy:</strong> Once 50% is revealed, you can copy the license key</li>
            <li>• <strong>Reset:</strong> Click "Reset Card" to try again</li>
          </ul>
        </div>
      </Card>
    </div>
  )
}
