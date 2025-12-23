/**
 * Modern Scratch Card Component for ISX Pulse License Activation
 * Features scratch-off animation, copy functionality, and mobile responsive design
 */

'use client'

import { useState, useRef, useCallback, useEffect } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Copy, Check, Gift, Sparkles, Lock, Unlock } from 'lucide-react'

import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { useToast } from '@/lib/hooks/use-toast'
import { useHydration } from '@/lib/hooks'
import { copyToClipboard } from '@/lib/utils/license-helpers'
import type { ScratchCardData } from '@/types/index'

interface ScratchCardProps {
  data: ScratchCardData
  onReveal?: (code: string) => void
  onCopy?: (code: string) => void
  className?: string
  size?: 'sm' | 'md' | 'lg'
  theme?: 'default' | 'premium' | 'gold'
}

const CARD_SIZES = {
  sm: { width: '280px', height: '180px' },
  md: { width: '350px', height: '220px' },
  lg: { width: '420px', height: '260px' }
}

const CARD_THEMES = {
  default: {
    background: 'from-blue-600 via-blue-700 to-indigo-800',
    overlay: 'from-gray-400 via-gray-500 to-gray-600',
    accent: 'text-blue-200',
    border: 'border-blue-500/30'
  },
  premium: {
    background: 'from-purple-600 via-purple-700 to-indigo-800',
    overlay: 'from-purple-400 via-purple-500 to-purple-600',
    accent: 'text-purple-200',
    border: 'border-purple-500/30'
  },
  gold: {
    background: 'from-yellow-600 via-amber-600 to-orange-700',
    overlay: 'from-yellow-400 via-amber-500 to-orange-600',
    accent: 'text-yellow-200',
    border: 'border-amber-500/30'
  }
}

export default function ScratchCard({
  data,
  onReveal,
  onCopy,
  className = '',
  size = 'md',
  theme = 'default'
}: ScratchCardProps) {
  const isHydrated = useHydration()
  const { toast } = useToast()
  const [isRevealed, setIsRevealed] = useState(data.revealed)
  const [isScratching, setIsScratching] = useState(false)
  const [scratchProgress, setScratchProgress] = useState(0)
  const [copied, setCopied] = useState(false)
  const [isHovered, setIsHovered] = useState(false)
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  
  const cardSize = CARD_SIZES[size]
  const cardTheme = CARD_THEMES[theme]

  // Initialize scratch canvas
  useEffect(() => {
    if (!isHydrated || isRevealed) return

    const canvas = canvasRef.current
    if (!canvas) return

    const ctx = canvas.getContext('2d')
    if (!ctx) return

    // Set canvas size
    const rect = canvas.getBoundingClientRect()
    canvas.width = rect.width * 2 // High DPI
    canvas.height = rect.height * 2
    ctx.scale(2, 2)

    // Create scratch overlay
    const gradient = ctx.createLinearGradient(0, 0, rect.width, rect.height)
    gradient.addColorStop(0, '#9CA3AF')
    gradient.addColorStop(0.5, '#6B7280')
    gradient.addColorStop(1, '#4B5563')
    
    ctx.fillStyle = gradient
    ctx.fillRect(0, 0, rect.width, rect.height)

    // Add texture pattern
    ctx.globalCompositeOperation = 'overlay'
    for (let i = 0; i < 100; i++) {
      ctx.fillStyle = `rgba(255, 255, 255, ${Math.random() * 0.1})`
      ctx.fillRect(
        Math.random() * rect.width,
        Math.random() * rect.height,
        Math.random() * 3,
        Math.random() * 3
      )
    }

    // Add "Scratch to reveal" text
    ctx.globalCompositeOperation = 'source-over'
    ctx.fillStyle = 'rgba(255, 255, 255, 0.8)'
    ctx.font = 'bold 18px Arial'
    ctx.textAlign = 'center'
    ctx.fillText('Scratch to Reveal', rect.width / 2, rect.height / 2 - 10)
    
    ctx.font = '14px Arial'
    ctx.fillStyle = 'rgba(255, 255, 255, 0.6)'
    ctx.fillText('Your License Code', rect.width / 2, rect.height / 2 + 15)

  }, [isHydrated, isRevealed])

  // Handle scratching
  const handleScratch = useCallback((e: React.MouseEvent | React.TouchEvent) => {
    if (isRevealed) return

    const canvas = canvasRef.current
    if (!canvas) return

    const ctx = canvas.getContext('2d')
    if (!ctx) return

    const rect = canvas.getBoundingClientRect()
    let x, y

    if ('touches' in e) {
      x = e.touches[0].clientX - rect.left
      y = e.touches[0].clientY - rect.top
    } else {
      x = e.clientX - rect.left
      y = e.clientY - rect.top
    }

    // Scale for high DPI
    x *= 2
    y *= 2

    // Create scratch effect
    ctx.globalCompositeOperation = 'destination-out'
    ctx.beginPath()
    ctx.arc(x, y, 30, 0, 2 * Math.PI)
    ctx.fill()

    // Calculate scratch progress
    const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height)
    const pixels = imageData.data
    let transparent = 0

    for (let i = 3; i < pixels.length; i += 4) {
      if (pixels[i] < 128) transparent++
    }

    const progress = transparent / (pixels.length / 4)
    setScratchProgress(progress)

    // Auto-reveal when enough is scratched
    if (progress > 0.5 && !isRevealed) {
      setTimeout(() => {
        setIsRevealed(true)
        onReveal?.(data.code)
      }, 500)
    }
  }, [isRevealed, data.code, onReveal])

  // Handle copy to clipboard
  const handleCopy = useCallback(async () => {
    if (!isRevealed) return

    const success = await copyToClipboard(data.code)
    if (success) {
      setCopied(true)
      onCopy?.(data.code)
      toast({
        title: 'Copied!',
        description: 'License code copied to clipboard',
      })
      setTimeout(() => setCopied(false), 2000)
    } else {
      toast({
        title: 'Copy failed',
        description: 'Unable to copy to clipboard',
        variant: 'destructive'
      })
    }
  }, [isRevealed, data.code, onCopy, toast])

  // Format code for display
  const formatCode = (code: string) => {
    if (data.format === 'scratch') {
      // ISX-XXXX-XXXX-XXXX format
      return code.replace(/(.{3})(.{4})(.{4})(.{4})/, '$1-$2-$3-$4')
    }
    // Standard format: ISX1M02LYE1F9QJHR9D7Z
    return code
  }

  if (!isHydrated) {
    return (
      <Card className={`${className} w-full max-w-sm mx-auto`} style={cardSize}>
        <CardContent className="flex items-center justify-center h-full">
          <div className="text-center">
            <Sparkles className="h-8 w-8 animate-spin mx-auto mb-2 text-muted-foreground" />
            <p className="text-sm text-muted-foreground">Loading scratch card...</p>
          </div>
        </CardContent>
      </Card>
    )
  }

  return (
    <motion.div
      className={`relative ${className}`}
      style={{ width: cardSize.width, height: cardSize.height }}
      initial={{ opacity: 0, scale: 0.9 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ duration: 0.5 }}
      onHoverStart={() => setIsHovered(true)}
      onHoverEnd={() => setIsHovered(false)}
    >
      {/* Background Card */}
      <Card className={`absolute inset-0 overflow-hidden ${cardTheme.border}`}>
        <div className={`absolute inset-0 bg-gradient-to-br ${cardTheme.background}`}>
          {/* Animated background pattern */}
          <div className="absolute inset-0 opacity-20">
            {Array.from({ length: 20 }).map((_, i) => (
              <motion.div
                key={i}
                className="absolute w-2 h-2 bg-white rounded-full"
                initial={{ 
                  x: Math.random() * 400, 
                  y: Math.random() * 300,
                  opacity: 0.1
                }}
                animate={{ 
                  x: Math.random() * 400, 
                  y: Math.random() * 300,
                  opacity: [0.1, 0.3, 0.1]
                }}
                transition={{ 
                  duration: 4 + Math.random() * 4,
                  repeat: Infinity,
                  ease: "linear"
                }}
              />
            ))}
          </div>
        </div>

        <CardContent className="relative z-10 h-full flex flex-col justify-between p-6">
          {/* Header */}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Gift className={`h-5 w-5 ${cardTheme.accent}`} />
              <span className={`text-sm font-medium ${cardTheme.accent}`}>
                ISX Pulse License
              </span>
            </div>
            <Badge variant="secondary" className="bg-white/20 text-white border-0">
              {data.format === 'scratch' ? 'Scratch Card' : 'Standard'}
            </Badge>
          </div>

          {/* Code Display Area */}
          <div className="flex-1 flex items-center justify-center">
            <AnimatePresence mode="wait">
              {isRevealed ? (
                <motion.div
                  key="revealed"
                  initial={{ opacity: 0, scale: 0.8 }}
                  animate={{ opacity: 1, scale: 1 }}
                  exit={{ opacity: 0, scale: 0.8 }}
                  className="text-center"
                >
                  <div className="flex items-center gap-2 mb-2">
                    <Unlock className="h-4 w-4 text-green-300" />
                    <span className="text-xs text-white/80 uppercase tracking-wide">
                      Revealed
                    </span>
                  </div>
                  <div className="font-mono text-lg md:text-xl font-bold text-white bg-black/20 rounded-lg px-4 py-2 mb-4">
                    {formatCode(data.code)}
                  </div>
                  <Button
                    variant="secondary"
                    size="sm"
                    onClick={handleCopy}
                    className="bg-white/20 hover:bg-white/30 text-white border-0"
                  >
                    {copied ? (
                      <>
                        <Check className="h-4 w-4 mr-2" />
                        Copied!
                      </>
                    ) : (
                      <>
                        <Copy className="h-4 w-4 mr-2" />
                        Copy Code
                      </>
                    )}
                  </Button>
                </motion.div>
              ) : (
                <motion.div
                  key="hidden"
                  initial={{ opacity: 1 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  className="text-center"
                >
                  <Lock className="h-8 w-8 text-white/60 mx-auto mb-2" />
                  <p className="text-white/80 text-sm">
                    Scratch to reveal your license code
                  </p>
                </motion.div>
              )}
            </AnimatePresence>
          </div>

          {/* Footer */}
          <div className="flex items-center justify-between text-xs text-white/60">
            <span>ISX Pulse</span>
            {data.activationId && (
              <span className="font-mono">
                ID: {data.activationId.slice(-8)}
              </span>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Scratch Overlay */}
      {!isRevealed && (
        <canvas
          ref={canvasRef}
          className="absolute inset-0 w-full h-full cursor-pointer rounded-lg"
          onMouseDown={(e) => {
            setIsScratching(true)
            handleScratch(e)
          }}
          onMouseMove={(e) => {
            if (isScratching) {
              handleScratch(e)
            }
          }}
          onMouseUp={() => setIsScratching(false)}
          onMouseLeave={() => setIsScratching(false)}
          onTouchStart={(e) => {
            setIsScratching(true)
            handleScratch(e)
          }}
          onTouchMove={handleScratch}
          onTouchEnd={() => setIsScratching(false)}
        />
      )}

      {/* Hover Effect */}
      <motion.div
        className="absolute inset-0 pointer-events-none rounded-lg"
        animate={{
          boxShadow: isHovered
            ? '0 20px 40px rgba(0, 0, 0, 0.3), 0 0 0 2px rgba(255, 255, 255, 0.1)'
            : '0 10px 20px rgba(0, 0, 0, 0.2)'
        }}
        transition={{ duration: 0.2 }}
      />

      {/* Sparkle effect on reveal */}
      <AnimatePresence>
        {isRevealed && (
          <motion.div
            initial={{ opacity: 1 }}
            animate={{ opacity: 0 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 2 }}
            className="absolute inset-0 pointer-events-none"
          >
            {Array.from({ length: 8 }).map((_, i) => (
              <motion.div
                key={i}
                className="absolute w-2 h-2"
                initial={{
                  x: '50%',
                  y: '50%',
                  scale: 0,
                  rotate: 0
                }}
                animate={{
                  x: `${50 + (Math.random() - 0.5) * 200}%`,
                  y: `${50 + (Math.random() - 0.5) * 200}%`,
                  scale: [0, 1, 0],
                  rotate: 360
                }}
                transition={{
                  duration: 1.5,
                  delay: i * 0.1,
                  ease: "easeOut"
                }}
              >
                <Sparkles className="h-4 w-4 text-yellow-400" />
              </motion.div>
            ))}
          </motion.div>
        )}
      </AnimatePresence>
    </motion.div>
  )
}