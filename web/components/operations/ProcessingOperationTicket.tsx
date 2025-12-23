import React, { useMemo } from "react";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { SegmentedFileProgress } from "@/components/operations/SegmentedFileProgress";
import { cn } from "@/lib/utils";
import {
  CheckCircle2,
  AlertCircle,
  Clock,
  Loader2,
  FileText,
} from "lucide-react";

interface ProcessingOperationTicketProps {
  operation: any;
}

const normalizeProgress = (value: any): number => {
  const num = typeof value === "number" ? value : Number(value);
  if (!Number.isFinite(num)) return 0;
  return Math.min(100, Math.max(0, Math.round(num)));
};

const toNumber = (value: any, fallback = 0): number => {
  const num = typeof value === "number" ? value : Number(value);
  return Number.isFinite(num) ? num : fallback;
};

const fmtDuration = (start?: Date | null, end?: Date | null) => {
  if (!start || !end) return null;
  const ms = end.getTime() - start.getTime();
  if (ms <= 0) return null;
  const minutes = Math.floor(ms / 60000);
  const seconds = Math.floor((ms % 60000) / 1000);
  if (minutes >= 1) {
    return `${minutes}m ${seconds}s`;
  }
  return `${seconds}s`;
};

const parseDate = (value: any): Date | null => {
  if (value instanceof Date) return value;
  if (typeof value === "string" && value) {
    const parsed = new Date(value);
    return Number.isNaN(parsed.getTime()) ? null : parsed;
  }
  return null;
};

export function ProcessingOperationTicket({
  operation,
}: ProcessingOperationTicketProps) {
  if (!operation) return null;

  const processingStep = Array.isArray(operation.steps)
    ? operation.steps.find(
      (step: any) =>
        step?.metadata?.stage_id === "processing" ||
        step?.stage_id === "processing" ||
        step?.id === "processing",
    ) || operation.steps[0]
    : operation;

  const metadata = processingStep?.metadata || operation.metadata || {};

  const status = (metadata.status ||
    processingStep?.status ||
    operation.status ||
    "running") as string;

  const filesProcessed = toNumber(
    metadata.files_processed ??
    metadata.processed_files ??
    processingStep?.files_processed,
    0,
  );
  const totalFiles = toNumber(
    metadata.total_files ??
    processingStep?.total_files ??
    (Array.isArray(metadata.file_list) ? metadata.file_list.length : 0),
    0,
  );
  const failedFiles = toNumber(metadata.failed_files ?? 0, 0);

  const computedProgress =
    totalFiles > 0
      ? Math.min(
        100,
        Math.max(
          0,
          Math.round(((filesProcessed + failedFiles) / totalFiles) * 100),
        ),
      )
      : null;

  const progress = normalizeProgress(
    computedProgress ??
    metadata.progress_percent ??
    processingStep?.progress ??
    operation.progress ??
    0,
  );

  const filesGenerated = toNumber(
    metadata.files_generated ?? metadata.outputs_generated ?? 0,
    0,
  );
  const totalOutputs = toNumber(
    metadata.total_outputs ??
    metadata.expected_outputs ??
    (Array.isArray(metadata.generation_file_statuses)
      ? metadata.generation_file_statuses.length
      : 0),
    0,
  );
  const generationProgress = normalizeProgress(
    (totalOutputs > 0
      ? Math.round((filesGenerated / totalOutputs) * 100)
      : null) ??
    metadata.generation_progress_percent ??
    0,
  );
  const generationStatuses = useMemo(() => {
    const raw = metadata.generation_file_statuses || metadata.file_statuses;
    if (!Array.isArray(raw)) {
      if (totalOutputs > 0) {
        return Array.from({ length: totalOutputs }).map((_, index) => ({
          filename: `Output ${index + 1}`,
          status:
            index < filesGenerated
              ? "completed"
              : index === filesGenerated
                ? "processing"
                : "pending",
          progress: index < filesGenerated ? 100 : 0,
          index,
        }));
      }
      return undefined;
    }
    return raw.map((status: any, index: number) => ({
      ...status,
      filename:
        status.filename ||
        status.file_name ||
        status.FileName ||
        status.name ||
        `Output ${index + 1}`,
      index,
    }));
  }, [metadata.generation_file_statuses, metadata.file_statuses, totalOutputs, filesGenerated]);
  const currentFile =
    metadata.current_file ?? processingStep?.current_file ?? "";

  const message =
    metadata.stage_message ||
    metadata.message ||
    operation.message ||
    (status === "completed"
      ? "Processing completed"
      : "Processing files...");

  const remaining =
    totalFiles > 0 ? Math.max(totalFiles - filesProcessed, 0) : null;

  const startedAt =
    parseDate(metadata.stage_started_at) ||
    parseDate(processingStep?.start_time) ||
    parseDate(operation?.start_time);
  const updatedAt =
    parseDate(metadata.stage_updated_at) ||
    parseDate(processingStep?.updated_at) ||
    parseDate(operation?.updated_at);
  const duration = useMemo(() => fmtDuration(startedAt, updatedAt), [startedAt, updatedAt]);

  const isComplete = status === "completed";
  const isFailed = status === "failed";

  const statusBadgeVariant =
    isComplete ? "secondary" : isFailed ? "destructive" : "default";
  const statusIcon = isComplete
    ? CheckCircle2
    : isFailed
      ? AlertCircle
      : Loader2;

  const showSegments = totalFiles > 0;
  const headerSubtitle = "Stage 2 of 6 • Depends on: Scraping";
  const normalizedFileStatuses = useMemo(() => {
    const raw = metadata.file_statuses;
    if (!Array.isArray(raw)) {
      if (totalFiles > 0) {
        return Array.from({ length: totalFiles }).map((_, index) => ({
          filename:
            (Array.isArray(metadata.file_list) && metadata.file_list[index]) ||
            `File ${index + 1}`,
          status:
            index < filesProcessed
              ? "completed"
              : index < filesProcessed + failedFiles
                ? "failed"
                : index === filesProcessed
                  ? "processing"
                  : "pending",
          progress: index < filesProcessed ? 100 : 0,
          index,
        }));
      }
      return undefined;
    }
    return raw.map((status: any, index: number) => ({
      ...status,
      filename:
        status.filename ||
        status.file_name ||
        status.FileName ||
        status.name ||
        `File ${index + 1}`,
      index,
    }));
  }, [metadata.file_statuses, metadata.file_list, totalFiles, filesProcessed, failedFiles]);

  return (
    <Card className="overflow-hidden border-muted/60">
      <CardHeader className="pb-2 pt-3 px-4">
        <div className="flex items-start justify-between gap-3">
          <div className="space-y-1">
            <p className="text-xs text-muted-foreground">{headerSubtitle}</p>
            <h3 className="text-base font-semibold">Excel to CSV Processing</h3>
          </div>
          <Badge
            variant={statusBadgeVariant as any}
            className="inline-flex items-center gap-1"
          >
            {React.createElement(statusIcon, { className: "h-3.5 w-3.5" })}
            {isComplete ? "Completed" : isFailed ? "Failed" : "Running"}
          </Badge>
        </div>
      </CardHeader>

      <CardContent className="space-y-2 px-4 pb-4">
        <div className="space-y-1.5">
          <div>
            <div className="flex items-center justify-between text-xs text-muted-foreground">
              <span>Input files</span>
              <span className="font-semibold text-foreground">{progress}%</span>
            </div>
            {showSegments ? (
              <SegmentedFileProgress
                totalFiles={totalFiles}
                processedFiles={filesProcessed}
                failedFiles={failedFiles}
                currentFile={currentFile}
                fileStatuses={normalizedFileStatuses}
                showDetails={false}
                showIndices={false}
              />
            ) : (
              <Progress value={progress} className="mt-1 h-2" />
            )}
          </div>

          {totalOutputs > 0 && (
            <div>
              <div className="flex items-center justify-between text-xs text-muted-foreground">
                <span>Outputs generated</span>
                <span className="font-semibold text-foreground">
                  {generationProgress}%
                </span>
              </div>
              <SegmentedFileProgress
                totalFiles={totalOutputs}
                processedFiles={filesGenerated}
                failedFiles={0}
                currentFile=""
                fileStatuses={generationStatuses}
                showDetails={false}
                showIndices={false}
              />
            </div>
          )}
        </div>

        <div
          className={cn(
            "rounded-md border p-2 text-xs sm:text-sm space-y-1.5",
            isComplete && "border-emerald-500/60 bg-emerald-950/10",
            isFailed && "border-red-500/60 bg-red-950/10",
            !isComplete && !isFailed && "border-slate-700/40 bg-slate-900/40",
          )}
        >
          <div className="font-semibold flex items-center gap-2">
            {React.createElement(statusIcon, { className: "h-4 w-4" })}
            <span>{message}</span>
          </div>
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-2 text-xs">
            <div className="flex items-center gap-2">
              <FileText className="h-3.5 w-3.5 text-muted-foreground" />
              <span className="truncate">
                {filesProcessed}/{totalFiles || "?"} processed
              </span>
            </div>
            {totalOutputs > 0 && (
              <div className="flex items-center gap-2">
                <FileText className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="truncate">
                  {filesGenerated}/{totalOutputs} generated
                </span>
              </div>
            )}
            <div className="flex items-center gap-2">
              <AlertCircle className="h-3.5 w-3.5 text-muted-foreground" />
              <span>Failures {failedFiles}</span>
            </div>
            {duration && (
              <div className="flex items-center gap-2">
                <Clock className="h-3.5 w-3.5 text-muted-foreground" />
                <span>{duration}</span>
              </div>
            )}
            {remaining !== null && (
              <div className="flex items-center gap-2">
                <Clock className="h-3.5 w-3.5 text-muted-foreground" />
                <span>{remaining} remaining</span>
              </div>
            )}
            {currentFile && (
              <div className="flex items-center gap-2 sm:col-span-2">
                <FileText className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="truncate" title={currentFile}>
                  {currentFile}
                </span>
              </div>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export default ProcessingOperationTicket;
