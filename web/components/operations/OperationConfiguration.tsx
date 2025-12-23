/**
 * OperationConfiguration Component
 * 
 * AGENT REQUIREMENTS FOR MODIFICATIONS:
 * =====================================
 * 
 * PRIMARY AGENTS:
 * - frontend-modernizer: For form handling, validation, UI/UX
 * - security-auditor: For input validation, security checks, data sanitization
 * 
 * SECONDARY AGENTS:
 * - test-architect: For form validation testing, edge cases
 * - documentation-enforcer: For configuration documentation
 * - api-integration-specialist: For config persistence and API calls
 * 
 * TASK-SPECIFIC AGENT ASSIGNMENTS:
 * - Form validation logic → security-auditor → frontend-modernizer
 * - Input sanitization → security-auditor (MANDATORY)
 * - Webhook configuration → operation-orchestrator → security-auditor
 * - Schedule/cron validation → devops-automator → frontend-modernizer
 * - Email validation → security-auditor → frontend-modernizer
 * - Config persistence → api-integration-specialist
 * - Form UX improvements → frontend-modernizer
 * - Error handling → frontend-modernizer → observability-engineer
 * 
 * SECURITY REQUIREMENTS (MANDATORY):
 * 1. ALL input validation MUST be reviewed by security-auditor
 * 2. URL validation for webhooks requires security-auditor
 * 3. Email validation patterns need security-auditor approval
 * 4. JSON parsing (customHeaders) requires security-auditor review
 * 5. Cron expressions must be validated by devops-automator
 * 
 * QUALITY GATES:
 * 1. Form changes require security-auditor review FIRST
 * 2. UI/UX changes need frontend-modernizer approval
 * 3. API integration needs api-integration-specialist validation
 * 4. Test coverage (min 85%) validated by test-architect
 * 
 * CRITICAL SECURITY NOTES:
 * - Never trust user input - always sanitize
 * - Validate URLs against allowlist patterns
 * - Prevent XSS in custom headers JSON
 * - Rate limit form submissions
 * - Log all configuration changes for audit
 * 
 * @see .claude/agents-workflow.md for detailed agent selection guide
 */

'use client'

import React, { useState, useEffect } from 'react'
import { apiClient } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { 
  AlertCircle, 
  Save, 
  X,
  ChevronDown,
  ChevronUp,
  Loader2
} from 'lucide-react'
import type { OperationConfig, OperationType } from '@/types'

interface OperationConfigurationProps {
  operationId: string
  operationType: OperationType
  currentConfig: OperationConfig
  onConfigSave: (config: OperationConfig) => void
  onCancel: () => void
  isLoading?: boolean
}

export function OperationConfiguration({
  operationId,
  operationType,
  currentConfig,
  onConfigSave,
  onCancel,
  isLoading = false
}: OperationConfigurationProps): JSX.Element {
  const [config, setConfig] = useState<OperationConfig>(currentConfig)
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [saving, setSaving] = useState(false)
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false)
  const [showConfirmDialog, setShowConfirmDialog] = useState(false)
  const [emailInput, setEmailInput] = useState('')
  const [webhookEnabled, setWebhookEnabled] = useState(!!currentConfig.webhookUrl)

  // Track changes
  useEffect(() => {
    const hasChanges = JSON.stringify(config) !== JSON.stringify(currentConfig)
    setHasUnsavedChanges(hasChanges)
  }, [config, currentConfig])

  // Sync webhook enabled state with incoming props
  useEffect(() => {
    setWebhookEnabled(!!currentConfig.webhookUrl)
  }, [currentConfig.webhookUrl])

  // Validate configuration
  const validateConfig = (): boolean => {
    const newErrors: Record<string, string> = {}

    // Common validations
    if (!config.schedule) {
      newErrors.schedule = 'Schedule is required'
    }

    if (config.retryAttempts !== undefined) {
      if (config.retryAttempts < 0 || config.retryAttempts > 10) {
        newErrors.retryAttempts = 'Must be between 0 and 10'
      }
    }

    // Email validation
    if (config.emailRecipients?.some((email: string) => !isValidEmail(email))) {
      newErrors.emailRecipients = 'Invalid email format'
    }

    // URL validation
    if (config.webhookUrl && !isValidUrl(config.webhookUrl)) {
      newErrors.webhookUrl = 'Must be a valid URL'
    }

    // Webhook enabled but no URL provided
    if (webhookEnabled && !config.webhookUrl) {
      newErrors.webhookUrl = 'Webhook URL is required'
    }

    // Cron validation for custom schedule
    if (config.schedule === 'custom' && !isValidCron(config.customCron || '')) {
      newErrors.customCron = 'Invalid cron expression'
    }

    // Conditional validations
    if (config.enableNotifications && !config.webhookUrl) {
      newErrors.webhookUrl = 'Webhook URL is required when webhook is enabled'
    }

    setErrors(newErrors)
    return Object.keys(newErrors).length === 0
  }

  // Helper functions
  const isValidEmail = (email: string): boolean => {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
  }

  const isValidUrl = (url: string): boolean => {
    try {
      new URL(url)
      return true
    } catch {
      return false
    }
  }

  const isValidCron = (cron: string): boolean => {
    // Simple cron validation - can be enhanced
    const parts = cron.split(' ')
    return parts.length >= 5
  }

  // Handle form submission
  const handleSave = async () => {
    if (!validateConfig()) {
      return
    }

    // Check for critical changes that need confirmation
    if (currentConfig.schedule !== 'manual' && config.schedule === 'manual') {
      setShowConfirmDialog(true)
      return
    }

    setSaving(true)
    try {
      await apiClient.updateOperationConfig(operationId, config)
      onConfigSave(config)
    } catch (error) {
      setErrors({ submit: error instanceof Error ? error.message : 'Configuration update failed' })
    } finally {
      setSaving(false)
    }
  }

  // Handle cancel with unsaved changes warning
  const handleCancel = () => {
    if (hasUnsavedChanges) {
      if (window.confirm('You have unsaved changes. Are you sure you want to cancel?')) {
        onCancel()
      }
    } else {
      onCancel()
    }
  }

  // Add email recipient
  const addEmailRecipient = () => {
    if (emailInput && isValidEmail(emailInput)) {
      setConfig((prev: OperationConfig) => ({
        ...prev,
        emailRecipients: [...(prev.emailRecipients || []), emailInput]
      }))
      setEmailInput('')
      setErrors((prev: Record<string, string>) => ({ ...prev, emailRecipients: '' }))
    } else {
      setErrors((prev: Record<string, string>) => ({ ...prev, emailRecipients: 'Invalid email format' }))
    }
  }

  // Remove email recipient
  const removeEmailRecipient = (email: string) => {
    setConfig((prev: OperationConfig) => {
      if (!prev.emailRecipients) return prev
      return {
        ...prev,
        emailRecipients: prev.emailRecipients.filter((e: string) => e !== email)
      }
    })
  }

  // Update config field
  const updateConfig = (field: keyof OperationConfig, value: any) => {
    setConfig((prev: OperationConfig) => ({ ...prev, [field]: value }))
    setErrors((prev: Record<string, string>) => ({ ...prev, [field]: '' }))
  }

  // Loading state
  if (isLoading) {
    return (
      <div className="flex items-center justify-center p-8" data-testid="config-loading">
        <Loader2 className="h-8 w-8 animate-spin mr-2" />
        <span>Loading configuration...</span>
      </div>
    )
  }

  // Save button disabled state
  const isSaveDisabled = !hasUnsavedChanges || saving || Object.keys(errors).length > 0

  return (
    <div className="panel-content-scrollable panel-scroll-custom max-h-[70vh]">
      <form
        role="form"
        aria-label="Operation Configuration"
        onSubmit={(e) => { e.preventDefault(); handleSave(); }}
        className="space-y-6"
      >
      {/* Error Alert */}
      {errors.submit && (
        <Alert variant="destructive" data-testid="error-alert">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>{errors.submit}</AlertDescription>
        </Alert>
      )}

      {/* Basic Configuration */}
      <Card className="panel-responsive-radius">
        <CardHeader className="sticky top-0 bg-card/95 backdrop-blur-sm z-10 border-b">
          <CardTitle>Basic Configuration</CardTitle>
          <CardDescription>Configure the basic settings for this operation</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-6 md:grid-cols-2 lg:grid-cols-3 space-y-4 p-6">
          {/* Schedule */}
          <div className="space-y-2">
            <Label htmlFor="schedule">Schedule*</Label>
            <select
              id="schedule"
              className="w-full px-3 py-2 border rounded-md"
              value={config.schedule || 'daily'}
              onChange={(e) => updateConfig('schedule', e.target.value)}
              aria-invalid={!!errors.schedule}
              aria-describedby={errors.schedule ? 'schedule-error' : undefined}
            >
              <option value="manual">Manual</option>
              <option value="hourly">Hourly</option>
              <option value="daily">Daily</option>
              <option value="weekly">Weekly</option>
              <option value="custom">Custom</option>
            </select>
            {errors.schedule && (
              <span id="schedule-error" role="alert" className="text-sm text-destructive">
                {errors.schedule}
              </span>
            )}
          </div>

          {/* Hourly specific */}
          {config.schedule === 'hourly' && (
            <div className="space-y-2">
              <Label htmlFor="hour">Hour of day</Label>
              <Input
                id="hour"
                type="number"
                min="0"
                max="23"
                value={config.hour || 0}
                onChange={(e) => updateConfig('hour', Number(e.target.value) || 0)}
              />
            </div>
          )}

          {/* Custom cron */}
          {config.schedule === 'custom' && (
            <div className="space-y-2">
              <Label htmlFor="cron">Cron Expression</Label>
              <Input
                id="cron"
                value={config.customCron || ''}
                onChange={(e) => updateConfig('customCron', e.target.value)}
                placeholder="0 0 * * *"
                aria-invalid={!!errors.customCron}
              />
              {errors.customCron && (
                <span role="alert" className="text-sm text-destructive">
                  {errors.customCron}
                </span>
              )}
            </div>
          )}

          {/* Timezone */}
          <div className="space-y-2">
            <Label htmlFor="timezone">Timezone</Label>
            <select
              id="timezone"
              className="w-full px-3 py-2 border rounded-md"
              value={config.timezone || 'Asia/Baghdad'}
              onChange={(e) => updateConfig('timezone', e.target.value)}
            >
              <option value="Asia/Baghdad">Asia/Baghdad</option>
              <option value="UTC">UTC</option>
              <option value="Europe/London">Europe/London</option>
              <option value="America/New_York">America/New York</option>
            </select>
          </div>

          {/* Retry Attempts */}
          <div className="space-y-2">
            <Label htmlFor="retryAttempts">Retry Attempts</Label>
            <Input
              id="retryAttempts"
              type="number"
              min="0"
              max="10"
              value={config.retryAttempts || 3}
              onChange={(e) => updateConfig('retryAttempts', Number(e.target.value) || 3)}
              aria-invalid={!!errors.retryAttempts}
            />
            {errors.retryAttempts && (
              <span role="alert" className="text-sm text-destructive">
                {errors.retryAttempts}
              </span>
            )}
          </div>

          {/* Timeout */}
          <div className="space-y-2">
            <Label htmlFor="timeout">Timeout (seconds)</Label>
            <Input
              id="timeout"
              type="number"
              min="30"
              max="3600"
              value={config.timeout || 300}
              onChange={(e) => updateConfig('timeout', Number(e.target.value) || 300)}
            />
          </div>
        </CardContent>
      </Card>

      {/* Operation-specific Configuration */}
      {operationType === 'report_generation' && (
        <Card className="panel-responsive-radius">
          <CardHeader className="sticky top-0 bg-card/95 backdrop-blur-sm z-10 border-b">
            <CardTitle>Report Settings</CardTitle>
          </CardHeader>
          <CardContent className="grid gap-6 md:grid-cols-2 lg:grid-cols-3 space-y-4 p-6">
            <div className="space-y-2">
              <Label htmlFor="reportFormat">Report Format</Label>
              <select
                id="reportFormat"
                className="w-full px-3 py-2 border rounded-md"
                value={config.reportFormat || 'pdf'}
                onChange={(e) => updateConfig('reportFormat', e.target.value)}
              >
                <option value="pdf">PDF</option>
                <option value="excel">Excel</option>
                <option value="csv">CSV</option>
              </select>
            </div>

            <div className="flex items-center space-x-2">
              <input
                type="checkbox"
                id="includeSummary"
                checked={config.includeSummary || false}
                onChange={(e) => updateConfig('includeSummary', e.target.checked)}
              />
              <Label htmlFor="includeSummary">Include Summary</Label>
            </div>

            <div className="space-y-2">
              <Label>Email Recipients</Label>
              <div className="flex space-x-2">
                <Input
                  placeholder="Add email"
                  value={emailInput}
                  onChange={(e) => setEmailInput(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                      e.preventDefault()
                      addEmailRecipient()
                    }
                  }}
                  aria-label="Add Email"
                />
                <Button type="button" onClick={addEmailRecipient}>Add</Button>
              </div>
              {errors.emailRecipients && (
                <span role="alert" className="text-sm text-destructive">
                  {errors.emailRecipients}
                </span>
              )}
              <div className="flex flex-wrap gap-2 mt-2">
                {config.emailRecipients?.map((email: string) => (
                  <Badge key={email} variant="secondary">
                    {email}
                    <button
                      type="button"
                      onClick={() => removeEmailRecipient(email)}
                      className="ml-2"
                      data-testid={`remove-email-${email}`}
                    >
                      <X className="h-3 w-3" />
                    </button>
                  </Badge>
                ))}
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {operationType === 'data_scraping' && (
        <Card className="panel-responsive-radius">
          <CardHeader className="sticky top-0 bg-card/95 backdrop-blur-sm z-10 border-b">
            <CardTitle>Scraping Settings</CardTitle>
          </CardHeader>
          <CardContent className="grid gap-6 md:grid-cols-2 lg:grid-cols-3 space-y-4 p-6">
            <div className="space-y-2">
              <Label htmlFor="dataSource">Data Source</Label>
              <select
                id="dataSource"
                className="w-full px-3 py-2 border rounded-md"
                value={config.dataSource || 'isx_website'}
                onChange={(e) => updateConfig('dataSource', e.target.value)}
              >
                <option value="isx_website">ISX Website</option>
                <option value="api">API</option>
                <option value="ftp">FTP Server</option>
              </select>
            </div>

            <div className="space-y-2">
              <Label htmlFor="scrapeInterval">Scrape Interval (minutes)</Label>
              <Input
                id="scrapeInterval"
                type="number"
                min="5"
                max="1440"
                value={config.scrapeInterval || 60}
                onChange={(e) => updateConfig('scrapeInterval', Number(e.target.value) || 60)}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="maxRetries">Max Retries</Label>
              <Input
                id="maxRetries"
                type="number"
                min="0"
                max="10"
                value={config.maxRetries || 5}
                onChange={(e) => updateConfig('maxRetries', Number(e.target.value) || 5)}
              />
            </div>
          </CardContent>
        </Card>
      )}

      {operationType === 'data_processing' && (
        <Card className="panel-responsive-radius">
          <CardHeader className="sticky top-0 bg-card/95 backdrop-blur-sm z-10 border-b">
            <CardTitle>Processing Settings</CardTitle>
          </CardHeader>
          <CardContent className="grid gap-6 md:grid-cols-2 lg:grid-cols-3 space-y-4 p-6">
            <div className="space-y-2">
              <Label htmlFor="batchSize">Batch Size</Label>
              <Input
                id="batchSize"
                type="number"
                min="1"
                max="1000"
                value={config.batchSize || 100}
                onChange={(e) => updateConfig('batchSize', Number(e.target.value) || 100)}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="parallelWorkers">Parallel Workers</Label>
              <Input
                id="parallelWorkers"
                type="number"
                min="1"
                max="10"
                value={config.parallelWorkers || 4}
                onChange={(e) => updateConfig('parallelWorkers', Number(e.target.value) || 4)}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="processingMode">Processing Mode</Label>
              <select
                id="processingMode"
                className="w-full px-3 py-2 border rounded-md"
                value={config.processingMode || 'batch'}
                onChange={(e) => updateConfig('processingMode', e.target.value)}
              >
                <option value="batch">Batch</option>
                <option value="stream">Stream</option>
                <option value="realtime">Real-time</option>
              </select>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Notifications */}
      <Card className="panel-responsive-radius">
        <CardHeader className="sticky top-0 bg-card/95 backdrop-blur-sm z-10 border-b">
          <CardTitle>Notifications</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-6 md:grid-cols-2 lg:grid-cols-3 space-y-4 p-6">
          <div className="flex items-center space-x-2">
            <input
              type="checkbox"
              id="enableNotifications"
              checked={config.enableNotifications || false}
              onChange={(e) => updateConfig('enableNotifications', e.target.checked)}
            />
            <Label htmlFor="enableNotifications">Enable Notifications</Label>
          </div>

          {config.enableNotifications && (
            <>
              <div className="space-y-2">
                <Label htmlFor="notificationChannel">Notification Channel</Label>
                <select
                  id="notificationChannel"
                  className="w-full px-3 py-2 border rounded-md"
                  value={config.notificationChannel || 'email'}
                  onChange={(e) => updateConfig('notificationChannel', e.target.value)}
                >
                  <option value="email">Email</option>
                  <option value="webhook">Webhook</option>
                  <option value="slack">Slack</option>
                </select>
              </div>

              <div className="space-y-2">
                <Label htmlFor="notificationThreshold">Notification Threshold (%)</Label>
                <Input
                  id="notificationThreshold"
                  type="number"
                  min="0"
                  max="100"
                  value={config.notificationThreshold || 90}
                  onChange={(e) => updateConfig('notificationThreshold', Number(e.target.value) || 90)}
                />
              </div>
            </>
          )}

          <div className="flex items-center space-x-2">
            <input
              type="checkbox"
              id="enableWebhook"
              checked={webhookEnabled}
              onChange={(e) => {
                setWebhookEnabled(e.target.checked)
                if (!e.target.checked) {
                  setErrors((prev: Record<string, string>) => {
                    const { webhookUrl, ...rest } = prev
                    return rest
                  })
                }
              }}
            />
            <Label htmlFor="enableWebhook">Enable Webhook</Label>
          </div>

          {webhookEnabled && (
            <div className="space-y-2">
              <Label htmlFor="webhookUrl">Webhook URL</Label>
              <Input
                id="webhookUrl"
                type="url"
                value={config.webhookUrl || ''}
                onChange={(e) => updateConfig('webhookUrl', e.target.value)}
                placeholder="https://example.com/webhook"
                aria-invalid={!!errors.webhookUrl}
              />
              {errors.webhookUrl && (
                <span role="alert" className="text-sm text-destructive">
                  {errors.webhookUrl}
                </span>
              )}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Advanced Settings */}
      <Card className="panel-responsive-radius">
        <CardHeader
          className="cursor-pointer sticky top-0 bg-card/95 backdrop-blur-sm z-10 border-b"
          onClick={() => setShowAdvanced(!showAdvanced)}
        >
          <div className="flex items-center justify-between">
            <CardTitle>Advanced Settings</CardTitle>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              className="panel-touch-target"
              aria-expanded={showAdvanced}
              aria-controls="advanced-settings-panel"
            >
              {showAdvanced ? <ChevronUp /> : <ChevronDown />}
              {showAdvanced ? 'Hide' : 'Show'} Advanced
            </Button>
          </div>
        </CardHeader>
        {showAdvanced && (
          <CardContent id="advanced-settings-panel" className="grid gap-6 md:grid-cols-2 lg:grid-cols-3 space-y-4 p-6">
            <div className="flex items-center space-x-2">
              <input
                type="checkbox"
                id="enableDebugMode"
                checked={config.enableDebugMode || false}
                onChange={(e) => updateConfig('enableDebugMode', e.target.checked)}
              />
              <Label htmlFor="enableDebugMode">Enable Debug Mode</Label>
            </div>

            <div className="space-y-2">
              <Label htmlFor="customHeaders">Custom Headers (JSON)</Label>
              <textarea
                id="customHeaders"
                className="w-full px-3 py-2 border rounded-md"
                rows={3}
                value={JSON.stringify(config.customHeaders || {}, null, 2)}
                onChange={(e) => {
                  try {
                    const headers = JSON.parse(e.target.value)
                    updateConfig('customHeaders', headers)
                    setErrors((prev: Record<string, string>) => {
                      const next = { ...prev }
                      delete next.customHeaders
                      return next
                    })
                  } catch {
                    setErrors((prev: Record<string, string>) => ({ ...prev, customHeaders: 'Invalid JSON format' }))
                  }
                }}
                aria-invalid={!!errors.customHeaders}
                aria-describedby={errors.customHeaders ? 'customHeaders-error' : undefined}
              />
              {errors.customHeaders && (
                <span id="customHeaders-error" role="alert" className="text-sm text-destructive">
                  {errors.customHeaders}
                </span>
              )}
            </div>
          </CardContent>
        )}
      </Card>

      {/* Actions */}
      <div className="flex flex-col sm:flex-row justify-end gap-4 p-6 bg-muted/30 rounded-lg">
        <Button
          type="button"
          variant="outline"
          onClick={handleCancel}
          className="panel-touch-target w-full sm:w-auto"
        >
          Cancel
        </Button>
        <Button
          type="submit"
          disabled={isSaveDisabled}
          aria-busy={saving}
          title={isSaveDisabled ? 'Fix errors or change something first' : 'Save configuration'}
          data-testid="save-config"
          className="panel-touch-target w-full sm:w-auto"
        >
          {saving ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Saving...
            </>
          ) : (
            <>
              <Save className="mr-2 h-4 w-4" />
              Save
            </>
          )}
        </Button>
      </div>

      {/* Confirmation Dialog */}
      {showConfirmDialog && (
        <div 
          role="dialog"
          aria-modal="true"
          aria-labelledby="confirm-dialog-heading"
          className="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50"
        >
          <Card className="max-w-md">
            <CardHeader>
              <CardTitle id="confirm-dialog-heading">Confirm Schedule Change</CardTitle>
            </CardHeader>
            <CardContent>
              <p>This will disable automatic execution. Are you sure you want to continue?</p>
            </CardContent>
            <div className="p-6 pt-0 flex justify-end space-x-4">
              <Button 
                variant="outline" 
                onClick={() => setShowConfirmDialog(false)}
              >
                Cancel
              </Button>
              <Button 
                onClick={() => {
                  setShowConfirmDialog(false)
                  setSaving(true)
                  apiClient.updateOperationConfig(operationId, config)
                    .then(() => onConfigSave(config))
                    .catch(err => setErrors({ submit: err.message }))
                    .finally(() => setSaving(false))
                }}
              >
                Confirm
              </Button>
            </div>
          </Card>
        </div>
      )}
      </form>
    </div>
  )
}