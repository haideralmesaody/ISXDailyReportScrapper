/**
 * Pipeline Connector Component Tests
 *
 * Tests for the visual connector between pipeline stages with accessibility support.
 * Ensures proper state transitions, ARIA compliance, and responsive behavior.
 */

import React from 'react'
import { render, screen } from '@testing-library/react'
import { describe, it, expect, test } from 'vitest'
import PipelineConnector from '../PipelineConnector'

describe('PipelineConnector', () => {
  it('renders correctly with default props', () => {
    render(<PipelineConnector isActive={false} isComplete={false} />)

    const connector = screen.getByRole('separator')
    expect(connector).toBeInTheDocument()
    expect(connector).toHaveAttribute('aria-label', 'Stage connector: pending')
    expect(connector).toHaveAttribute('aria-valuenow', '0')
  })

  it('displays active state correctly', () => {
    render(<PipelineConnector isActive={true} isComplete={false} />)

    const connector = screen.getByRole('separator')
    expect(connector).toHaveAttribute('aria-label', 'Stage connector: active')
    expect(connector).toHaveAttribute('aria-valuenow', '50')
    expect(connector).toHaveClass('bg-gradient-to-b', 'from-blue-300', 'to-blue-400')
  })

  it('displays completed state correctly', () => {
    render(<PipelineConnector isActive={false} isComplete={true} />)

    const connector = screen.getByRole('separator')
    expect(connector).toHaveAttribute('aria-label', 'Stage connector: completed')
    expect(connector).toHaveAttribute('aria-valuenow', '100')
    expect(connector).toHaveClass('bg-gradient-to-b', 'from-green-400', 'to-blue-400')
  })

  it('applies custom className prop', () => {
    render(<PipelineConnector isActive={false} isComplete={false} className="custom-class" />)

    const connector = screen.getByRole('separator')
    expect(connector).toHaveClass('custom-class')
  })

  it('prioritizes completed state over active state', () => {
    render(<PipelineConnector isActive={true} isComplete={true} />)

    const connector = screen.getByRole('separator')
    expect(connector).toHaveAttribute('aria-label', 'Stage connector: completed')
    expect(connector).toHaveAttribute('aria-valuenow', '100')
    expect(connector).toHaveClass('from-green-400', 'to-blue-400')
  })

  it('has proper accessibility attributes', () => {
    render(<PipelineConnector isActive={true} isComplete={false} />)

    const connector = screen.getByRole('separator')
    expect(connector).toHaveAttribute('role', 'separator')
    expect(connector).toHaveAttribute('aria-label')
    expect(connector).toHaveAttribute('aria-valuemin', '0')
    expect(connector).toHaveAttribute('aria-valuemax', '100')
    expect(connector).toHaveAttribute('aria-valuenow')
    expect(connector).toHaveAttribute('aria-valuetext')
  })

  test('accessibility attributes update correctly', () => {
    const { rerender } = render(<PipelineConnector isActive={false} isComplete={false} />)

    let connector = screen.getByRole('separator')
    expect(connector).toHaveAttribute('aria-valuetext', 'Stage connector: pending')

    // Update to active state
    rerender(<PipelineConnector isActive={true} isComplete={false} />)
    connector = screen.getByRole('separator')
    expect(connector).toHaveAttribute('aria-valuetext', 'Stage connector: active')

    // Update to completed state
    rerender(<PipelineConnector isActive={false} isComplete={true} />)
    connector = screen.getByRole('separator')
    expect(connector).toHaveAttribute('aria-valuetext', 'Stage connector: completed')
  })

  it('has consistent visual structure', () => {
    const { container } = render(<PipelineConnector isActive={true} isComplete={false} />)

    // Check for proper container structure
    const containerDiv = container.firstChild
    expect(containerDiv).toHaveClass('flex', 'justify-center', 'py-2')

    // Check for connector element
    const connector = containerDiv?.firstChild
    expect(connector).toHaveClass('w-0.5', 'h-8', 'rounded-full', 'transition-all', 'duration-500')
  })

  test('handles rapid state changes gracefully', () => {
    const { rerender } = render(<PipelineConnector isActive={false} isComplete={false} />)

    // Rapidly change states to test transition handling
    rerender(<PipelineConnector isActive={true} isComplete={false} />)
    rerender(<PipelineConnector isActive={true} isComplete={true} />)
    rerender(<PipelineConnector isActive={false} isComplete={true} />)

    const connector = screen.getByRole('separator')
    expect(connector).toHaveAttribute('aria-valuenow', '100')
    expect(connector).toHaveClass('from-green-400', 'to-blue-400')
  })
})