import React from "react";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ScrapingTelemetry } from "@/lib/operations/types";
import { SegmentedDayProgress } from "./SegmentedDayProgress";
import { cn } from "@/lib/utils";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { AlertCircle } from "lucide-react";
import type { PipelineContext } from "@/lib/operations/pipeline-context";

const toNumber = (value: any): number | undefined => {
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "string" && value.trim().length) {
    const parsed = Number(value);
    if (Number.isFinite(parsed)) return parsed;
  }
  return undefined;
};

interface ScrapingOperationTicketProps {
  telemetry?: ScrapingTelemetry;
  operation?: any;
  pipelineContext?: PipelineContext | null;
}

export function ScrapingOperationTicket({
  telemetry,
  operation,
  pipelineContext,
}: ScrapingOperationTicketProps) {
  const telem = telemetry ?? {};
  const missingTelemetry = !telemetry || Object.keys(telem).length === 0;

  const fromDate = telem.from_date || telem.fromDate;
  const toDate = telem.to_date || telem.toDate;
  const totalTradingDays = toNumber(telem.trading_days_total);
  const completedTradingDays = toNumber(telem.trading_days_completed);
  const remainingTradingDays = toNumber(telem.trading_days_remaining);
  const holidaysDetected = toNumber(telem.holidays_detected) ?? (Array.isArray(telem.skipped_files) ? telem.skipped_files.length : undefined);

  const progressPercent =
    typeof telem.progress_percent === "number"
      ? Math.max(0, Math.min(100, telem.progress_percent))
      : (typeof totalTradingDays === "number" &&
        typeof completedTradingDays === "number" &&
        totalTradingDays > 0)
        ? Math.round((completedTradingDays / totalTradingDays) * 100)
        : undefined;

  const normalizedPhase =
    typeof telem.phase === "string"
      ? telem.phase.toLowerCase()
      : undefined;
  const status =
    telem.status ||
    (normalizedPhase === "complete" ? "completed" : normalizedPhase) ||
    "running";
  const statusMessage =
    telem.stage_message ||
    (status === "completed"
      ? "Scraping completed successfully"
      : status === "failed"
        ? "Scraping failed"
        : missingTelemetry
          ? "Telemetry missing for scraping stage"
          : "Downloading files...");

  const telemetryErrors: string[] = [];
  if (typeof telem.progress_percent !== "number" &&
    !(typeof totalTradingDays === "number" && typeof completedTradingDays === "number")) {
    telemetryErrors.push("Missing progress telemetry");
  }
  if (typeof totalTradingDays !== "number") telemetryErrors.push("Missing trading_days_total");
  if (typeof completedTradingDays !== "number") telemetryErrors.push("Missing trading_days_completed");
  if (typeof holidaysDetected !== "number") telemetryErrors.push("Missing holidays_detected");

  const title =
    operation?.name ||
    (fromDate && toDate
      ? `Data Collection: ${fromDate} to ${toDate}`
      : "Data Collection");

  const stageLabel = (() => {
    const total = pipelineContext?.totalStages;
    const num = pipelineContext?.stageNumberById?.scraping;
    if (typeof num === "number" && typeof total === "number" && total > 0) {
      return `Stage ${num} of ${total}`;
    }
    return "Stage 1";
  })();

  const metadataForSegments = {
    skipped_files: telem.skipped_files || [],
    stage_id: "scraping",
    phase: status === "completed" ? "complete" : status,
    phase_message: statusMessage,
    current_file: telem.current_file,
    trading_days_total: totalTradingDays,
    trading_days_completed: completedTradingDays,
    holidays_detected: holidaysDetected,
    status,
  };

  return (
    <Card className="overflow-hidden">
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between gap-4">
          <div>
            <p className="text-sm text-muted-foreground">{stageLabel}</p>
            <h3 className="text-lg font-semibold">{title}</h3>
            {fromDate && toDate && (
              <p className="text-xs text-muted-foreground">
                {toDate} - {fromDate}
              </p>
            )}
          </div>
          <div className="flex flex-col items-end gap-1">
            <Badge
              variant={
                status === "completed"
                  ? "secondary"
                  : status === "failed"
                    ? "destructive"
                    : "default"
              }
            >
              {status === "completed"
                ? "Completed"
                : status === "failed"
                  ? "Failed"
                  : "Running"}
            </Badge>
            {telem.watchdog_status === "timeout" && (
              <Badge variant="destructive" className="text-[10px]">
                Watchdog timeout
              </Badge>
            )}
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-3">
        {telemetryErrors.length > 0 && (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription className="text-xs sm:text-sm">
              Telemetry error: {telemetryErrors.join(", ")}
            </AlertDescription>
          </Alert>
        )}

        {fromDate && toDate && (
          <SegmentedDayProgress
            fromDate={fromDate}
            toDate={toDate}
            downloadedFiles={telem.downloaded_files || []}
            currentFile={telem.current_file}
            metadata={metadataForSegments}
          />
        )}

        <div className="grid gap-3 rounded-md border border-border/60 bg-muted/20 p-3 text-xs sm:text-sm">
          <div className="flex flex-wrap gap-4">
            <div>
              <span className="font-semibold text-foreground">
                {completedTradingDays ?? "—"}
              </span>{" "}
              / {totalTradingDays ?? "—"} trading days
            </div>
            <div>{remainingTradingDays ?? "—"} remaining</div>
            <div>{holidaysDetected ?? "—"} holidays detected</div>
            {telem.current_file && status !== "completed" && (
              <div className="truncate">
                Current file:{" "}
                <span className="font-medium text-foreground">
                  {telem.current_file}
                </span>
              </div>
            )}
          </div>

          <div
            className={cn(
              "rounded-md px-3 py-2 text-xs sm:text-sm",
              status === "completed" && "bg-emerald-100 text-emerald-900",
              status === "failed" && "bg-red-100 text-red-900",
              status === "running" && "bg-blue-100 text-blue-900",
            )}
          >
            {statusMessage}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export default ScrapingOperationTicket;
