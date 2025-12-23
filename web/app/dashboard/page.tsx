/**
 * Dashboard Route - Redirects to Operations
 * Operations is the main working area, replacing the traditional dashboard
 */

import { redirect } from 'next/navigation'

export default function DashboardPage() {
  // Redirect to operations (main working area)
  redirect('/operations')
}