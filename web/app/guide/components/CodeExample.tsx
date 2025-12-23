'use client'

import { useState } from 'react'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { vscDarkPlus, vs } from 'react-syntax-highlighter/dist/esm/styles/prism'
import { useTheme } from 'next-themes'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Check, Copy, FileCode } from 'lucide-react'
import { cn } from '@/lib/utils'

interface CodeExampleProps {
  language: 'typescript' | 'javascript' | 'bash' | 'json' | 'go'
  code: string
  title?: string
  filename?: string
  highlightLines?: number[]
  showLineNumbers?: boolean
  className?: string
}

/**
 * Syntax-highlighted code example component
 * Supports multiple languages and copy-to-clipboard
 */
export function CodeExample({
  language,
  code,
  title,
  filename,
  highlightLines = [],
  showLineNumbers = true,
  className
}: CodeExampleProps) {
  const { theme } = useTheme()
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    await navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const languageLabels: Record<typeof language, string> = {
    typescript: 'TypeScript',
    javascript: 'JavaScript',
    bash: 'Bash',
    json: 'JSON',
    go: 'Go'
  }

  return (
    <div className={cn('rounded-lg border overflow-hidden', className)}>
      {/* Header */}
      {(title || filename) && (
        <div className="flex items-center justify-between px-4 py-2 bg-muted/50 border-b">
          <div className="flex items-center gap-2">
            <FileCode className="h-4 w-4 text-muted-foreground" />
            {filename && (
              <span className="text-sm font-mono text-muted-foreground">
                {filename}
              </span>
            )}
            {title && !filename && (
              <span className="text-sm font-medium">
                {title}
              </span>
            )}
          </div>
          <div className="flex items-center gap-2">
            <Badge variant="secondary" className="text-xs">
              {languageLabels[language]}
            </Badge>
            <Button
              size="sm"
              variant="ghost"
              onClick={handleCopy}
              className="h-7 gap-1.5"
            >
              {copied ? (
                <>
                  <Check className="h-3 w-3" />
                  Copied
                </>
              ) : (
                <>
                  <Copy className="h-3 w-3" />
                  Copy
                </>
              )}
            </Button>
          </div>
        </div>
      )}

      {/* Code Block */}
      <div className="relative">
        <SyntaxHighlighter
          language={language}
          style={theme === 'dark' ? vscDarkPlus : vs}
          showLineNumbers={showLineNumbers}
          wrapLines={highlightLines.length > 0}
          lineProps={(lineNumber) => {
            const style: React.CSSProperties = { display: 'block' }
            if (highlightLines.includes(lineNumber)) {
              style.backgroundColor = theme === 'dark' ? 'rgba(255, 255, 0, 0.1)' : 'rgba(255, 255, 0, 0.2)'
            }
            return { style }
          }}
          customStyle={{
            margin: 0,
            borderRadius: 0,
            fontSize: '0.875rem',
            lineHeight: '1.5'
          }}
        >
          {code.trim()}
        </SyntaxHighlighter>
      </div>
    </div>
  )
}
