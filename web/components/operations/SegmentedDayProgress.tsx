/**
 * Segmented Day Progress Component
 * Professional progress bar where each segment represents one day
 * with intelligent holiday detection and color coding
 */

"use client";

import React, { useMemo } from "react";
import { cn } from "@/lib/utils";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

interface DaySegment {
  date: Date;
  dateString: string;
  dayOfWeek: number;
  dayName: string;
  dayNumber: number;
  monthName: string;
  status: "downloaded" | "pending" | "weekend" | "holiday" | "downloading";
}

interface SegmentedDayProgressProps {
  fromDate?: string;
  toDate?: string;
  downloadedFiles?: string[];
  currentFile?: string;
  metadata?: Record<string, any>;
  className?: string;
}

const toNumber = (value: any): number | undefined => {
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "string" && value.trim().length) {
    const parsed = Number(value);
    if (Number.isFinite(parsed)) return parsed;
  }
  return undefined;
};

export function SegmentedDayProgress({
  fromDate,
  toDate,
  downloadedFiles = [],
  currentFile,
  metadata,
  className,
}: SegmentedDayProgressProps) {
  // Generate day segments
  const segments = useMemo(() => {
    if (!fromDate || !toDate) return [];

    const result: DaySegment[] = [];
    const start = new Date(fromDate);
    const end = new Date(toDate);

    // Create set of downloaded dates for fast lookup
    // Handle both space-separated (2025 08 07) and hyphenated (2025-08-07) formats
    const downloadedDates = new Set(
      downloadedFiles
        .map((f) => {
          // Try space-separated format first (backend sends this)
          const spaceMatch = f.match(/(\d{4})\s+(\d{2})\s+(\d{2})/);
          if (spaceMatch) {
            return `${spaceMatch[1]}-${spaceMatch[2]}-${spaceMatch[3]}`;
          }
          // Fallback to hyphenated format
          const hyphenMatch = f.match(/\d{4}-\d{2}-\d{2}/);
          return hyphenMatch ? hyphenMatch[0] : null;
        })
        .filter((value): value is string => typeof value === "string"),
    );

    const normalizedStatus =
      typeof metadata?.status === "string"
        ? metadata.status.toLowerCase()
        : undefined;
    const isReallyComplete =
      normalizedStatus === "completed" ||
      metadata?.phase === "completed" ||
      metadata?.phase === "complete" ||
      metadata?.completed === true;

    // Extract currently downloading date (handle both formats)
    // CRITICAL: Never set downloading date if operation is complete
    let downloadingDate: string | undefined;
    if (!isReallyComplete && currentFile) {
      const spaceMatch = currentFile.match(/(\d{4})\s+(\d{2})\s+(\d{2})/);
      if (spaceMatch) {
        downloadingDate = `${spaceMatch[1]}-${spaceMatch[2]}-${spaceMatch[3]}`;
      } else {
        downloadingDate = currentFile.match(/\d{4}-\d{2}-\d{2}/)?.[0];
      }
    }

    // Build list of all trading days that should have files (in REVERSE order)
    const tradingDays: string[] = [];
    // Start from END date and go backwards to START date
    for (let d = new Date(end); d >= start; d.setDate(d.getDate() - 1)) {
      const current = new Date(d);
      const dayOfWeek = current.getDay();
      // Iraq weekend is Friday (5) and Saturday (6)
      if (dayOfWeek !== 5 && dayOfWeek !== 6) {
        tradingDays.push(current.toISOString().slice(0, 10));
      }
    }

    // Get skipped files from metadata (authoritative holiday detection)
    const skippedFilesArray = Array.isArray(metadata?.skipped_files)
      ? metadata.skipped_files.filter((item: unknown): item is string => typeof item === "string")
      : [];

    // Create set of skipped dates for fast lookup
    const skippedDates = new Set(
      skippedFilesArray
        .map((f) => {
          // Try space-separated format first (backend sends this)
          const spaceMatch = f.match(/(\d{4})\s+(\d{2})\s+(\d{2})/);
          if (spaceMatch) {
            return `${spaceMatch[1]}-${spaceMatch[2]}-${spaceMatch[3]}`;
          }
          // Fallback to hyphenated format
          const hyphenMatch = f.match(/\d{4}-\d{2}-\d{2}/);
          return hyphenMatch ? hyphenMatch[0] : null;
        })
        .filter((value): value is string => typeof value === "string"),
    );

    // Generate segments for each day (REVERSE ORDER - most recent first)
    // Start from END date and go backwards to START date
    for (let d = new Date(end); d >= start; d.setDate(d.getDate() - 1)) {
      const current = new Date(d);
      const dateStr = current.toISOString().slice(0, 10);
      const dayOfWeek = current.getDay();

      const isWeekend = dayOfWeek === 5 || dayOfWeek === 6;
      const isHoliday = skippedDates.has(dateStr);
      const isDownloaded = downloadedDates.has(dateStr);
      const isDownloading = dateStr === downloadingDate;

      // Determine status
      let status: DaySegment["status"];
      if (isWeekend) {
        status = "weekend";
      } else if (isHoliday) {
        status = "holiday";
      } else if (isDownloaded) {
        status = "downloaded";
      } else if (isDownloading) {
        status = "downloading";
      } else {
        status = "pending";
      }

      result.push({
        date: new Date(current),
        dateString: dateStr,
        dayOfWeek,
        dayName: current.toLocaleDateString("en-US", { weekday: "short" }),
        dayNumber: current.getDate(),
        monthName: current.toLocaleDateString("en-US", { month: "short" }),
        status,
      });
    }

    return result;
  }, [
    fromDate,
    toDate,
    downloadedFiles,
    currentFile,
    metadata?.phase,
    metadata?.status,
    metadata?.skipped_files,
  ]);

  // Calculate statistics using simple file-based progress
  const stats = useMemo(() => {
    const total = segments.length;
    const weekends = segments.filter((s) => s.status === "weekend").length;
    const holidays = segments.filter((s) => s.status === "holiday").length;
    const downloadedSegments = segments.filter((s) => s.status === "downloaded").length;
    const downloading = segments.filter((s) => s.status === "downloading").length;

    const tradingDaysTotalMeta = toNumber(metadata?.trading_days_total);
    const tradingDaysCompletedMeta = toNumber(metadata?.trading_days_completed);
    const tradingDaysRemainingMeta = toNumber(metadata?.trading_days_remaining);
    const holidayCount = toNumber(metadata?.holidays_detected);

    // Prefer backend totals; fall back to segment counts only if missing
    const resolvedTradingDaysTotal = typeof tradingDaysTotalMeta === "number" ? tradingDaysTotalMeta : segments.length - weekends;
    const resolvedDownloaded = typeof tradingDaysCompletedMeta === "number" ? tradingDaysCompletedMeta : downloadedSegments;
    const resolvedPending =
      typeof tradingDaysRemainingMeta === "number"
        ? tradingDaysRemainingMeta
        : Math.max(resolvedTradingDaysTotal - resolvedDownloaded, 0);

    // Prefer backend progress_percent; otherwise clamp resolved values
    let progress: number | undefined;
    if (typeof metadata?.progress_percent === "number") {
      progress = Math.max(0, Math.min(100, metadata.progress_percent));
    } else if (typeof resolvedTradingDaysTotal === "number" && resolvedTradingDaysTotal > 0) {
      progress = Math.max(
        0,
        Math.min(100, Math.round((resolvedDownloaded / resolvedTradingDaysTotal) * 100))
      );
    }
    if (metadata?.phase === "completed" || metadata?.phase === "complete" || metadata?.completed === true) {
      progress = 100;
    }

    // Determine calculation method for display (favor backend hint)
    const calculationMethod = metadata?.progress_calculation || "backend";

    return {
      total,
      weekends,
      holidays, // Keep for UI display only
      downloaded: resolvedDownloaded,
      downloading,
      tradingDaysTotal: resolvedTradingDaysTotal,
      pending: resolvedPending,
      progress,
      calculationMethod,
      missingTelemetry: {
        tradingDaysTotalMissing: typeof tradingDaysTotalMeta !== "number",
        tradingDaysCompletedMissing: typeof tradingDaysCompletedMeta !== "number",
        holidaysMissing: typeof holidayCount !== "number",
      },
    };
  }, [segments, metadata]);

  // Don't render if no date range
  if (!fromDate || !toDate || segments.length === 0) {
    return null;
  }

  const showTelemetryError =
    stats.missingTelemetry.tradingDaysTotalMissing ||
    stats.missingTelemetry.tradingDaysCompletedMissing ||
    stats.missingTelemetry.holidaysMissing;

  // Calculate segment width
  const segmentWidth = 100 / segments.length;

  return (
    <div className={cn("space-y-2", className)}>
      {showTelemetryError && (
        <div className="text-xs text-red-600 bg-red-50 border border-red-200 rounded-md px-3 py-2">
          Telemetry missing:
          {stats.missingTelemetry.tradingDaysTotalMissing && " trading_days_total"}
          {stats.missingTelemetry.tradingDaysCompletedMissing && " trading_days_completed"}
          {stats.missingTelemetry.holidaysMissing && " holidays_detected"}
        </div>
      )}
      {/* Progress header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <span className="text-sm font-medium">Progress</span>
          <span className="text-lg font-bold">
            {stats.progress !== undefined ? `${stats.progress}%` : "—"}
          </span>
          <span className="text-xs text-muted-foreground">
            ({stats.downloaded}/{stats.tradingDaysTotal} files)
            {stats.calculationMethod === 'file_based' && (
              <span className="text-xs text-blue-600 ml-1">• File-based</span>
            )}
            {stats.calculationMethod === 'trading_days' && (
              <span className="text-xs text-blue-600 ml-1">• Trading days</span>
            )}
          </span>
        </div>

        {/* Compact legend */}
        <div className="flex items-center gap-2 text-xs">
          <div className="flex items-center gap-1">
            <div className="h-2 w-2 rounded-sm bg-green-500" />
            <span className="text-muted-foreground">Done</span>
          </div>
          <div className="flex items-center gap-1">
            <div className="h-2 w-2 rounded-sm bg-gray-600" />
            <span className="text-muted-foreground">Weekend</span>
          </div>
          <div className="flex items-center gap-1">
            <div className="h-2 w-2 rounded-sm bg-orange-500" />
            <span className="text-muted-foreground">Holiday</span>
          </div>
          {(stats.calculationMethod === 'file_based' || stats.calculationMethod === 'trading_days') && (
            <div className="flex items-center gap-1">
              <div className="h-2 w-2 rounded-sm bg-blue-200 border border-blue-400" />
              <span className="text-muted-foreground">
                {stats.calculationMethod === 'trading_days' ? 'Trading-day Progress' : 'File-based Progress'}
              </span>
            </div>
          )}
        </div>
      </div>

      {/* Segmented progress bar */}
      <div className="w-full">
        <TooltipProvider delayDuration={0}>
          <div className="flex h-3 w-full overflow-hidden rounded-full bg-gray-200 dark:bg-gray-800">
            {segments.map((segment, index) => (
              <Tooltip key={segment.dateString}>
                <TooltipTrigger asChild>
                  <div
                    className={cn(
                      "h-full transition-all duration-300",
                      "hover:opacity-80 cursor-default",
                      // Colors based on status
                      segment.status === "downloaded" && "bg-green-500",
                      segment.status === "downloading" &&
                        "bg-blue-500 animate-pulse",
                      segment.status === "weekend" && "bg-gray-600",
                      segment.status === "holiday" && "bg-orange-500",
                      segment.status === "pending" &&
                        "bg-white dark:bg-gray-700",
                      // Add subtle borders between segments
                      index > 0 && "border-l border-gray-300/30",
                    )}
                    style={{ width: `${segmentWidth}%` }}
                  />
                </TooltipTrigger>
                <TooltipContent className="text-xs">
                  <div className="space-y-0.5">
                    <div className="font-medium">
                      {segment.dayName}, {segment.monthName} {segment.dayNumber}
                    </div>
                    <div className="text-muted-foreground capitalize">
                      {segment.status === "downloading"
                        ? "⏳ Downloading..."
                        : segment.status === "downloaded"
                          ? "✓ Downloaded"
                          : segment.status === "weekend"
                            ? "Weekend"
                            : segment.status === "holiday"
                              ? "🎉 Holiday"
                              : "Pending"}
                    </div>
                    {stats.calculationMethod === 'file_based' && (
                      <div className="text-xs text-blue-600 pt-1 border-t border-gray-200 mt-1">
                        Simple file-based progress calculation
                      </div>
                    )}
                  </div>
                </TooltipContent>
              </Tooltip>
            ))}
          </div>
        </TooltipProvider>
      </div>

      {/* Statistics */}
      {(stats.holidays > 0 || stats.pending > 0) && (
        <div className="flex items-center justify-between text-xs text-muted-foreground">
          <div>
            {new Date(fromDate).toLocaleDateString("en-US", {
              month: "short",
              day: "numeric",
            })}
            {" - "}
            {new Date(toDate).toLocaleDateString("en-US", {
              month: "short",
              day: "numeric",
              year: "numeric",
            })}
          </div>
          <div className="flex items-center gap-3">
            {stats.pending > 0 && <span>{stats.pending} remaining</span>}
            {stats.holidays > 0 && (
              <span>{stats.holidays} holidays detected</span>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

export default SegmentedDayProgress;
