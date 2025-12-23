import type { Metadata } from 'next'
import { Inter } from 'next/font/google'
import { ThemeProvider } from 'next-themes'

import './globals.css'
import { ToastProvider, ToastViewport } from '@/components/ui/toast'
import { AppContent } from '@/components/app-content'

const inter = Inter({ subsets: ['latin'] })

// Metadata configuration for ISX Pulse
export const metadata: Metadata = {
  title: {
    default: 'ISX Pulse - The Heartbeat of Iraqi Markets',
    template: '%s | ISX Pulse'
  },
  description: 'ISX Pulse provides real-time market intelligence for the Iraq Stock Exchange, delivering professional-grade data analytics and comprehensive financial reporting for investors and institutions.',
  keywords: [
    'ISX Pulse', 
    'ISX', 
    'Iraqi Stock Exchange', 
    'Market Intelligence', 
    'Professional Trading', 
    'Financial Analysis', 
    'Investment Research',
    'Middle East Markets',
    'Iraq Finance',
    'Bull Market Analytics'
  ],
  authors: [{ name: 'ISX Pulse Team' }],
  icons: {
    icon: [
      { url: '/favicon-16x16.png', sizes: '16x16', type: 'image/png' },
      { url: '/favicon-32x32.png', sizes: '32x32', type: 'image/png' },
      { url: '/favicon.ico', sizes: 'any' }
    ],
    apple: '/apple-touch-icon.png',
    other: [
      { rel: 'android-chrome-192x192', url: '/android-chrome-192x192.png' },
      { rel: 'android-chrome-512x512', url: '/android-chrome-512x512.png' }
    ]
  },
  robots: 'noindex,nofollow',
  referrer: 'strict-origin-when-cross-origin'
}

// Viewport configuration (separate export for Next.js 14+)
export const viewport = {
  width: 'device-width',
  initialScale: 1,
  themeColor: '#2d5a3d',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}): JSX.Element {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <link rel="icon" href="/favicon.ico" sizes="any" />
        <link rel="icon" href="/favicon-16x16.png" sizes="16x16" type="image/png" />
        <link rel="icon" href="/favicon-32x32.png" sizes="32x32" type="image/png" />
        <link rel="apple-touch-icon" href="/apple-touch-icon.png" />
        <link rel="manifest" href="/site.webmanifest" />
        <meta name="robots" content="noindex,nofollow" />
        {/* Note: X-Frame-Options, X-Content-Type-Options, X-XSS-Protection, and CSP should be set as HTTP headers in the Go backend, not meta tags */}
      </head>
      <body className={inter.className} suppressHydrationWarning>
        <ThemeProvider 
          attribute="class" 
          defaultTheme="dark"
          enableSystem={false}
          disableTransitionOnChange
        >
          <ToastProvider>
            <AppContent>
              {children}
            </AppContent>
            <ToastViewport />
          </ToastProvider>
        </ThemeProvider>
      </body>
    </html>
  )
}