/**
 * IndicatorSettingsPopover - Generic settings popup for indicators
 *
 * Displays a form for editing indicator parameters
 * Anchored to gear icon on indicator pane
 */

'use client'

import { useState } from 'react'
import { Settings } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogFooter,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'

export interface SettingsField {
  name: string
  label: string
  type: 'number' | 'color' | 'select'
  min?: number
  max?: number
  step?: number
  options?: { value: number | null; label: string }[]
  hasCustomInput?: boolean  // Enable custom number input for select fields
}

interface IndicatorSettingsPopoverProps {
  indicatorName: string
  settings: Record<string, any>
  fields: SettingsField[]
  onUpdate: (newSettings: Record<string, any>) => void
  triggerButton?: React.ReactNode
}

export function IndicatorSettingsPopover({
  indicatorName,
  settings,
  fields,
  onUpdate,
  triggerButton,
}: IndicatorSettingsPopoverProps) {
  const [open, setOpen] = useState(false)
  const [localSettings, setLocalSettings] = useState(settings)

  const handleSave = () => {
    onUpdate(localSettings)
    setOpen(false)
  }

  const handleCancel = () => {
    setLocalSettings(settings) // Reset to original
    setOpen(false)
  }

  // Default trigger for chart panes (positioned absolutely)
  const defaultTrigger = (
    <Button
      variant="ghost"
      size="icon"
      className="absolute top-2 right-2 h-6 w-6 rounded-sm bg-background/50 hover:bg-background/80 backdrop-blur-sm"
    >
      <Settings className="h-3.5 w-3.5" />
    </Button>
  )

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        {triggerButton || defaultTrigger}
      </DialogTrigger>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>{indicatorName} Settings</DialogTitle>
        </DialogHeader>
        <div className="space-y-4 py-4">
          {fields.map((field) => {
            // For select fields with custom input, check if current value is in presets
            const currentValue = localSettings[field.name]
            const isCustomValue = field.hasCustomInput && field.options
              ? !field.options.some(opt => opt.value === currentValue)
              : false

            // Determine what to show in the select dropdown
            let selectValue: string
             if (isCustomValue) {
               selectValue = 'custom'  // Show "Custom" option
             } else {
              selectValue = currentValue === null ? 'null' : currentValue?.toString() ?? ''
             }

            return (
              <div key={field.name} className="grid gap-2">
                <Label htmlFor={field.name} className="text-sm">
                  {field.label}
                </Label>
                {field.type === 'select' && field.options ? (
                  <>
                    <Select
                      value={selectValue}
                      onValueChange={(value) => {
                        if (value === 'custom') {
                          // Custom selected - initialize with min value if needed
                          if (!isCustomValue) {
                            setLocalSettings({
                              ...localSettings,
                              [field.name]: field.min || 1,
                            })
                          }
                        } else {
                          // Preset selected - parse and save
                          const parsedValue = value === 'null' ? null : parseFloat(value)
                          setLocalSettings({
                            ...localSettings,
                            [field.name]: parsedValue,
                          })
                        }
                      }}
                    >
                      <SelectTrigger className="h-8">
                        <SelectValue placeholder="Select an option" />
                      </SelectTrigger>
                      <SelectContent>
                        {field.options.map((option) => (
                          <SelectItem
                            key={option.value?.toString() ?? 'null'}
                            value={option.value?.toString() ?? 'null'}
                          >
                            {option.label}
                          </SelectItem>
                        ))}
                        {field.hasCustomInput && (
                          <SelectItem value="custom">
                            Custom{isCustomValue ? ` (${currentValue})` : ''}
                          </SelectItem>
                        )}
                      </SelectContent>
                    </Select>
                    {/* Show custom input when in custom mode */}
                    {field.hasCustomInput && isCustomValue && (
                      <Input
                        type="number"
                        value={currentValue || ''}
                        onChange={(e) => {
                          const val = parseFloat(e.target.value)
                          setLocalSettings({
                            ...localSettings,
                            [field.name]: isNaN(val) ? (field.min || 1) : val,
                          })
                        }}
                        min={field.min}
                        max={field.max}
                        step={field.step || 1}
                        placeholder={`Enter value (${field.min}-${field.max})`}
                        className="h-8"
                      />
                    )}
                  </>
                ) : (
                  <Input
                    id={field.name}
                    type={field.type}
                    value={localSettings[field.name] || ''}
                    onChange={(e) =>
                      setLocalSettings({
                        ...localSettings,
                        [field.name]:
                          field.type === 'number'
                            ? parseFloat(e.target.value)
                            : e.target.value,
                      })
                    }
                    min={field.min}
                    max={field.max}
                    step={field.step}
                    className="h-8"
                  />
                )}
              </div>
            )
          })}
        </div>
        <DialogFooter>
          <Button
            variant="outline"
            onClick={handleCancel}
          >
            Cancel
          </Button>
          <Button onClick={handleSave}>
            Apply
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
