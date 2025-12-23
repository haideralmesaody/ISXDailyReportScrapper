"use client"

import React from "react"
import DatePicker from "react-datepicker"
import { Calendar } from "lucide-react"
import { cn } from "@/lib/utils"

// Import the default styles
import "react-datepicker/dist/react-datepicker.css"

interface DatePickerProps {
  selected?: Date
  onChange?: (date: Date | null) => void
  placeholderText?: string
  className?: string
  disabled?: boolean
  minDate?: Date
  maxDate?: Date
}

export function CustomDatePicker({
  selected,
  onChange,
  placeholderText = "Select date",
  className,
  disabled = false,
  minDate,
  maxDate
}: DatePickerProps) {
  return (
    <div className="relative">
      <DatePicker
        selected={selected}
        onChange={onChange}
        placeholderText={placeholderText}
        disabled={disabled}
        minDate={minDate}
        maxDate={maxDate}
        dateFormat="MMMM d, yyyy"
        className={cn(
          "flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background",
          "placeholder:text-muted-foreground",
          "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2",
          "disabled:cursor-not-allowed disabled:opacity-50",
          className
        )}
        wrapperClassName="w-full"
        showPopperArrow={false}
        popperClassName="react-datepicker-popper"
        calendarClassName="react-datepicker-calendar"
        showMonthDropdown
        showYearDropdown
        dropdownMode="select"
        isClearable
        autoComplete="off"
      />
      <Calendar className="absolute right-3 top-2.5 h-4 w-4 text-muted-foreground pointer-events-none" />
    </div>
  )
}

// Export with simpler names
export { CustomDatePicker as DatePickerField }
export { CustomDatePicker as DatePicker }