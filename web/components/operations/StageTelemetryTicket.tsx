/**
 * Lightweight stage ticket for stages without dedicated UI.
 * Shows backend telemetry as-is (no generic progress bars).
 */

"use client";

import React from "react";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

interface StageTelemetryTicketProps {
  operation: any;
}

export function StageTelemetryTicket({ operation }: StageTelemetryTicketProps) {
  const step = Array.isArray(operation?.steps) ? operation.steps[0] : operation;
  const metadata = step?.metadata || operation?.metadata || {};
  const stageId = metadata.stage_id || step?.id || operation?.stage_id || operation?.current_step || "stage";
  const status = step?.status || operation?.status || "pending";
  const progressPercent = typeof metadata.progress_percent === "number" ? metadata.progress_percent : undefined;

  const telemetryErrors: string[] = [];
  if (typeof metadata.stage_id !== "string" && typeof step?.id !== "string") {
    telemetryErrors.push("stage_id missing");
  }
  if (typeof progressPercent !== "number") {
    telemetryErrors.push("progress_percent missing");
  }

  // Pick a few numeric fields to surface
  const metrics = Object.entries(metadata as Record<string, unknown>)
    .filter((entry): entry is [string, number] => typeof entry[1] === "number")
    .slice(0, 6);

  return (
    <Card className="overflow-hidden">
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div>
            <p className="text-sm text-muted-foreground">Stage</p>
            <h3 className="text-lg font-semibold capitalize">{stageId}</h3>
          </div>
          <Badge
            variant={
              status === "completed"
                ? "secondary"
                : status === "failed"
                  ? "destructive"
                  : "default"
            }
          >
            {status || "unknown"}
          </Badge>
        </div>
      </CardHeader>
      <CardContent className="space-y-2 text-sm">
        {telemetryErrors.length > 0 && (
          <div className="flex items-center gap-2 text-xs text-red-600 bg-red-50 border border-red-200 rounded-md px-3 py-2">
            Telemetry error: {telemetryErrors.join(", ")}
          </div>
        )}
        <div className="flex items-center gap-2 text-muted-foreground">
          <span className="font-semibold text-foreground">Progress:</span>
          <span>{progressPercent !== undefined ? `${progressPercent}%` : "—"}</span>
        </div>
        {metrics.length > 0 ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-xs">
            {metrics.map(([key, value]) => (
              <div key={key} className="flex items-center justify-between rounded-md bg-muted/30 px-2 py-1">
                <span className="text-muted-foreground">{key}</span>
                <span className="font-medium text-foreground">{value}</span>
              </div>
            ))}
          </div>
        ) : (
          <div className="text-xs text-muted-foreground">No numeric telemetry provided.</div>
        )}
        {metadata.stage_message && (
          <div
            className={cn(
              "rounded-md px-3 py-2 text-xs",
              status === "completed" && "bg-emerald-100 text-emerald-900",
              status === "failed" && "bg-red-100 text-red-900",
              status !== "completed" && status !== "failed" && "bg-blue-100 text-blue-900",
            )}
          >
            {metadata.stage_message}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export default StageTelemetryTicket;
