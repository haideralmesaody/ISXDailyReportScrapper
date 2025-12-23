/**
 * ProcessingStageCard
 * Mirrors the scraping card layout but shows processing-specific telemetry.
 */

"use client";

import React from "react";
import { cn } from "@/lib/utils";
import { SegmentedFileProgress } from "./SegmentedFileProgress";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { AlertCircle } from "lucide-react";

interface ProcessingStageCardProps {
  metadata?: Record<string, any>;
  statusMessage?: string;
  isComplete?: boolean;
  isFailed?: boolean;
  variant?: "progress" | "complete";
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

export function ProcessingStageCard({
  metadata,
  statusMessage,
  isComplete,
  isFailed,
  variant = "progress",
  className,
}: ProcessingStageCardProps) {
  const totalFiles = toNumber(metadata?.total_files);
  const processedFiles = toNumber(metadata?.files_processed);
  const failedFiles = toNumber(metadata?.failed_files) ?? 0;
  const currentFile = metadata?.current_file;
  const fileStatuses = metadata?.file_statuses as Array<{
    filename: string;
    status: "pending" | "processing" | "completed" | "failed";
    size_mb?: number;
    progress?: number;
    error_message?: string;
    processing_time_ms?: number;
  }> | undefined;

  const telemetryErrors: string[] = [];
  if (totalFiles === undefined) telemetryErrors.push("Missing total_files");
  if (processedFiles === undefined) telemetryErrors.push("Missing files_processed");
  if (!Array.isArray(fileStatuses) || fileStatuses.length === 0) telemetryErrors.push("Missing file_statuses");

  return (
    <div className={cn("space-y-3", className)}>
      <SegmentedFileProgress
        totalFiles={totalFiles}
        processedFiles={processedFiles}
        failedFiles={failedFiles}
        currentFile={currentFile}
        fileStatuses={fileStatuses}
        processingSpeedMBps={metadata?.processing_speed_mbps}
        totalSizeMB={metadata?.total_size_mb}
        processedSizeMB={metadata?.processed_size_mb}
        estimatedRemainingMs={metadata?.estimated_remaining_ms}
        showDetails={true}
      />

      {statusMessage && variant === "progress" && (
        <div
          className={cn(
            "rounded-md p-3 text-sm",
            isComplete && "bg-emerald-50 text-emerald-900",
            isFailed && "bg-red-50 text-red-900",
            !isComplete && !isFailed && "bg-blue-50 text-blue-900",
          )}
        >
          {statusMessage}
        </div>
      )}

      {variant === "complete" && (
        <div className="text-xs text-muted-foreground space-y-1">
          <div>
            Processed{" "}
            <span className="font-semibold text-foreground">
              {processedFiles}/{totalFiles}
            </span>{" "}
            files{failedFiles > 0 && (
              <> with <span className="font-semibold text-foreground">{failedFiles}</span> failures.</>
            )}
          </div>
          {totalOutputs !== undefined && totalOutputs > 0 && (
            <div>
              Generated{" "}
              <span className="font-semibold text-foreground">
                {generatedOutputs}/{totalOutputs}
              </span>{" "}
              CSV files
            </div>
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

export default ProcessingStageCard;
