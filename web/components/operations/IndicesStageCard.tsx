/**
 * IndicesStageCard
 * Aligns with Scraping/Processing visuals but focused on index extraction metrics.
 */

"use client";

import React from "react";
import { cn } from "@/lib/utils";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { AlertCircle } from "lucide-react";
import { SegmentedFileProgress } from "./SegmentedFileProgress";

interface IndicesStageCardProps {
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

export function IndicesStageCard({
  metadata,
  statusMessage,
  isComplete,
  isFailed,
  variant = "progress",
  className,
}: IndicesStageCardProps) {
  // Handle both backend field names for compatibility
  const recordsProcessed = toNumber(metadata?.records_processed) ?? toNumber(metadata?.files_processed);
  const indicesExtracted = toNumber(metadata?.indices_extracted);
  const isx15Extracted = toNumber(metadata?.isx15_extracted);
  const totalFiles = toNumber(metadata?.total_files);
  const currentFile = metadata?.current_file;
  const failedFiles = toNumber(metadata?.failed_files) ?? 0;

  // Get file statuses - backend uses snake_case JSON field names
  const fileStatuses = metadata?.file_statuses as Array<{
    file_name: string;      // snake_case from backend
    status: "pending" | "processing" | "completed" | "failed";
    size_mb?: number;
    progress?: number;
    error_message?: string;
    processing_ms?: number;
  }> | undefined;

  const telemetryErrors: string[] = [];
  if (recordsProcessed === undefined && (totalFiles ?? 0) !== 0) telemetryErrors.push("Missing files_processed");
  if (indicesExtracted === undefined && (totalFiles ?? 0) > 0) telemetryErrors.push("Missing indices_extracted");
  if (totalFiles === undefined && recordsProcessed !== undefined) telemetryErrors.push("Missing total_files");

  return (
    <div className={cn("space-y-3", className)}>
      <SegmentedFileProgress
        totalFiles={totalFiles}
        processedFiles={recordsProcessed}
        failedFiles={failedFiles}
        currentFile={currentFile}
        fileStatuses={fileStatuses ? fileStatuses.map(f => ({
          filename: f.file_name,
          status: f.status,
          ...(f.size_mb !== undefined ? { size_mb: f.size_mb } : {}),
          ...(f.progress !== undefined ? { progress: f.progress } : {}),
          ...(f.error_message !== undefined ? { error_message: f.error_message } : {}),
          ...(f.processing_ms !== undefined ? { processing_time_ms: f.processing_ms } : {})
        })) : undefined}
        showDetails={false}
        showIndices={false}
      />

      {/* Status Message - consistent with other stages */}
      {statusMessage && (
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
        <div className="text-xs text-muted-foreground">
          Extracted{" "}
          <span className="font-semibold text-foreground">
            {indicesExtracted?.toLocaleString() ?? 0}
          </span>{" "}
          ISX60 indices from{" "}
          <span className="font-semibold text-foreground">
            {recordsProcessed?.toLocaleString() ?? 0}
          </span>{" "}
          files
          {isx15Extracted !== undefined && isx15Extracted > 0 && (
            <>
              {" "}and{" "}
              <span className="font-semibold text-foreground">
                {isx15Extracted.toLocaleString()}
              </span>{" "}
              ISX15 indices
            </>
          )}.
          {failedFiles > 0 && (
            <span className="text-red-600 ml-2">
              ({failedFiles} failed)
            </span>
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

export default IndicesStageCard;
