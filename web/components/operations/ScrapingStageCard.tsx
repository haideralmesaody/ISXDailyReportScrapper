/**
 * ScrapingStageCard
 * Canonical renderer for scraping telemetry in progress & completion tickets.
 */

"use client";

import React from "react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { AlertCircle } from "lucide-react";
import { cn } from "@/lib/utils";
import { SegmentedDayProgress } from "./SegmentedDayProgress";

interface ScrapingStageCardProps {
  metadata?: Record<string, any>;
  statusMessage?: string;
  isComplete?: boolean;
  isFailed?: boolean;
  variant?: "progress" | "complete";
  className?: string;
}

const toStringArray = (value: any): string[] => {
  if (Array.isArray(value)) {
    return value.filter((item): item is string => typeof item === "string");
  }
  return [];
};

const toNumber = (value: any): number | undefined => {
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "string" && value.trim().length) {
    const parsed = Number(value);
    if (Number.isFinite(parsed)) return parsed;
  }
  return undefined;
};

export function ScrapingStageCard({
  metadata,
  statusMessage,
  isComplete,
  isFailed,
  variant = "progress",
  className,
}: ScrapingStageCardProps) {
  const estimateTradingDays = (start?: string, end?: string): number | undefined => {
    if (!start || !end) return undefined;
    try {
      const from = new Date(start);
      const to = new Date(end);
      if (Number.isNaN(from.getTime()) || Number.isNaN(to.getTime())) return undefined;
      let count = 0;
      for (let d = new Date(from); d <= to; d.setDate(d.getDate() + 1)) {
        const dow = d.getDay();
        // Iraq weekend: Friday (5) and Saturday (6)
        if (dow === 5 || dow === 6) continue;
        count++;
      }
      return count || undefined;
    } catch {
      return undefined;
    }
  };

  const fromDate = metadata?.from_date || metadata?.start_date;
  const toDate = metadata?.to_date || metadata?.end_date;
  const downloadedFiles = toStringArray(metadata?.downloaded_files);
  const skippedFiles = toStringArray(metadata?.skipped_files);
  const tradingDaysTotal =
    toNumber(metadata?.trading_days_total) ??
    toNumber(metadata?.total_files) ??
    toNumber(metadata?.expected_files) ??
    estimateTradingDays(fromDate, toDate);
  const holidaysDetected = toNumber(metadata?.holidays_detected) ?? skippedFiles.length;
  const tradingDaysCompleted =
    toNumber(metadata?.trading_days_completed) ??
    toNumber(metadata?.files_downloaded);
  const filesDownloaded = downloadedFiles.length;

  const progressBase = tradingDaysCompleted ?? (filesDownloaded + (holidaysDetected ?? 0));
  const progress =
    tradingDaysTotal !== undefined && progressBase !== undefined
      ? Math.round((progressBase / tradingDaysTotal) * 100)
      : undefined;

  const telemetryErrors: string[] = [];
  if (tradingDaysTotal === undefined) telemetryErrors.push("Missing trading_days_total");
  if (tradingDaysCompleted === undefined && filesDownloaded === 0) telemetryErrors.push("Missing trading_days_completed");
  if (holidaysDetected === undefined) telemetryErrors.push("Missing holidays_detected");

  const remainingTradingDays =
    tradingDaysTotal !== undefined && progressBase !== undefined
      ? Math.max(tradingDaysTotal - progressBase, 0)
      : undefined;

  const currentFile = metadata?.current_file;
  const watchdogStatus = metadata?.watchdog_status;

  if (!fromDate || !toDate) {
    return null;
  }

  return (
    <div className={cn("space-y-3", className)}>
      <SegmentedDayProgress
        fromDate={fromDate}
        toDate={toDate}
        downloadedFiles={downloadedFiles}
        currentFile={currentFile}
        metadata={metadata}
      />

      <div
        className={cn(
          "grid gap-2 rounded-lg border border-border/60 bg-muted/20 p-3",
          isComplete && "bg-emerald-50/50 dark:bg-emerald-950/30",
          isFailed && "bg-red-50/50 dark:bg-red-950/30",
        )}
      >
        <div className="flex flex-wrap items-center gap-4 text-xs sm:text-sm">
          <div>
            <span className="font-semibold text-foreground">
              {tradingDaysCompleted ?? "—"}
            </span>{" "}
            /<span className="text-muted-foreground"> {tradingDaysTotal ?? "—"}</span>{" "}
            trading days
          </div>
          {holidaysDetected !== undefined && holidaysDetected > 0 && (
            <div className="text-muted-foreground">
              Non-trading days detected: {holidaysDetected}
            </div>
          )}
          {remainingTradingDays !== undefined && (
            <div className="text-muted-foreground">
              {remainingTradingDays} remaining
            </div>
          )}
          {holidaysDetected !== undefined && (
            <div className="text-muted-foreground">
              Holidays detected: {holidaysDetected}
            </div>
          )}
          {progress !== undefined && (
            <div className="text-muted-foreground">
              Progress: {progress}%
            </div>
          )}
          {currentFile && !isComplete && (
            <div className="truncate text-muted-foreground">
              Current file:{" "}
              <span className="font-medium text-foreground">{currentFile}</span>
            </div>
          )}
        </div>

        {statusMessage && variant === "progress" && (
          <div
            className={cn(
              "rounded-md p-2 text-xs sm:text-sm",
              isComplete && "bg-emerald-100 text-emerald-900",
              isFailed && "bg-red-100 text-red-900",
              !isComplete && !isFailed && "bg-blue-100 text-blue-900",
            )}
          >
            {statusMessage}
          </div>
        )}

        {watchdogStatus === "timeout" && (
          <Alert variant="destructive" className="py-2">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription className="text-xs sm:text-sm">
              Scraper watchdog triggered – no files downloaded for several
              minutes. Operation cancelled.
            </AlertDescription>
          </Alert>
        )}
      </div>

      {variant === "complete" && (
        <div className="text-xs text-muted-foreground">
          Completed{" "}
          <span className="font-semibold text-foreground">
            {tradingDaysCompleted ?? "—"}/{tradingDaysTotal ?? "—"}
          </span>{" "}
          trading days
          {holidaysDetected !== undefined && holidaysDetected > 0 && (
            <> with <span className="font-semibold text-foreground">{holidaysDetected}</span> holidays detected.</>
          )}
        </div>
      )}

      {telemetryErrors.length > 0 && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription className="text-xs sm:text-sm">
            Telemetry error: {telemetryErrors.join(", ")}
          </AlertDescription>
        </Alert>
      )}
    </div>
  );
}
