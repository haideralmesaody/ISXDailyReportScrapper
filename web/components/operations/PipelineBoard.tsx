import React, { useMemo, useCallback } from "react";
import { Card } from "@/components/ui/card";
import { PipelineCollapsible } from "@/components/operations/PipelineCollapsible";
import { PipelineSummaryCard } from "@/components/operations/PipelineSummaryCard";
import { OperationCompleteCard } from "@/components/operations/OperationCompleteCard";
import ProcessingOperationTicket from "@/components/operations/ProcessingOperationTicket";
import ScrapingOperationTicket from "@/components/operations/ScrapingOperationTicket";
import { IndicesStageCard } from "@/components/operations/IndicesStageCard";
import { Card as ShadCard, CardContent as ShadCardContent, CardHeader as ShadCardHeader } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import StageTelemetryTicket from "@/components/operations/StageTelemetryTicket";

interface PipelineBoardProps {
  operations: any[];
  operationTypes: any[];
  resolveOperationType: (op: any) => string | undefined;
  onConfigureOperation: (type: any) => void;
}

export function PipelineBoard({
  operations,
  operationTypes,
  resolveOperationType,
  onConfigureOperation,
}: PipelineBoardProps) {
  const pipelineOperations = useMemo(() => {
    if (!operations) return [];
    try {
      return operations.filter(
        (op) => resolveOperationType(op) === "full_pipeline",
      );
    } catch (error) {
      console.warn('Pipeline operations filtering failed:', error);
      return [];
    }
  }, [operations, resolveOperationType]);

  if (!pipelineOperations.length) {
    return null;
  }

  // Memoize active pipeline finder to prevent TDZ errors
  const activePipeline = useMemo(() => {
    try {
      return pipelineOperations.find(
        (op) => op.status === "running" || op.status === "pending",
      ) || pipelineOperations[0];
    } catch (error) {
      console.warn('Active pipeline finder failed:', error);
      return pipelineOperations[0] || null;
    }
  }, [pipelineOperations]);

  // Memoize operation type finder to prevent TDZ errors
  const findNextOperationType = useCallback((type: string) => {
    try {
      return operationTypes?.find((t) => t.id === type) || null;
    } catch (error) {
      console.warn('Next operation type finder failed:', error);
      return null;
    }
  }, [operationTypes]);

  // Memoize pipeline steps processing to prevent TDZ errors
  const processedPipelineSteps = useMemo(() => {
    return pipelineOperations.map((operation) => {
      if (
        operation.status === "completed" &&
        Array.isArray(operation.steps) &&
        operation.steps.length > 1
      ) {
        return operation.steps.map((step, stepIndex) => {
          const stageOperation = {
            ...operation,
            operation_id: operation.operation_id,
            react_key: `${operation.operation_id}-${step.id}-completed`,
            steps: [step],
            progress: 100,
            status: "completed",
            metadata: {
              pipelineMetadata: operation.metadata,
              ...operation.metadata,
              ...step.metadata,
              stage_id: step.metadata?.stage_id || step.id,
              stage_timeline:
                step.metadata?.stage_timeline ||
                operation.metadata?.pipeline_summary?.stage_timeline,
            },
          };
          return {
            key: stageOperation.react_key,
            operation: stageOperation,
            operationId: operation.operation_id,
            stepIndex,
            totalStages: operation.steps.length,
            showConnector: stepIndex < operation.steps.length - 1
          };
        });
      }
      return null;
    }).filter(Boolean);
  }, [pipelineOperations]);

  const renderIndicesTicket = useCallback((operation: any) => {
    const step = Array.isArray(operation.steps) ? operation.steps[0] : operation;
    const metadata = step?.metadata || operation?.metadata || {};
    return (
      <ShadCard className="overflow-hidden">
        <ShadCardHeader className="pb-3">
          <div className="flex items-start justify-between">
            <div>
              <p className="text-sm text-muted-foreground">Index Extraction</p>
              <h3 className="text-lg font-semibold">Stage 3 of 6</h3>
            </div>
            <Badge variant={operation.status === "completed" ? "secondary" : operation.status === "failed" ? "destructive" : "default"}>
              {operation.status || "unknown"}
            </Badge>
          </div>
        </ShadCardHeader>
        <ShadCardContent>
          <IndicesStageCard metadata={metadata} isComplete={operation.status === "completed"} isFailed={operation.status === "failed"} />
        </ShadCardContent>
      </ShadCard>
    );
  }, []);

  const renderStageTicket = useCallback((operation: any) => {
    const step = Array.isArray(operation.steps) ? operation.steps[0] : operation;
    const stageId =
      step?.metadata?.stage_id ||
      step?.id ||
      operation?.metadata?.stage_id ||
      operation?.current_step ||
      operation?.stage_id ||
      "scraping";

    switch (stageId) {
      case "scraping":
        return (
          <ScrapingOperationTicket
            key={`${operation.operation_id}-scraping-ticket`}
            telemetry={operation.metadata}
            operation={operation}
          />
        );
      case "processing":
        return (
          <ProcessingOperationTicket
            key={`${operation.operation_id}-processing-ticket`}
            operation={operation}
          />
        );
      case "indices":
        return (
          <React.Fragment key={`${operation.operation_id}-indices-ticket`}>
            {renderIndicesTicket(operation)}
          </React.Fragment>
        );
      default:
        return (
          <StageTelemetryTicket
            key={`${operation.operation_id}-generic-${stageId}`}
            operation={operation}
          />
        );
    }
  }, [renderIndicesTicket]);

  return (
    <div className="space-y-4">
      {pipelineOperations.map((operation, index) => {
        if (
          operation.status === "completed" &&
          Array.isArray(operation.steps) &&
          operation.steps.length > 1
        ) {
          return (
            <div key={`${operation.operation_id}-pipeline-complete`}>
              <OperationCompleteCard
                operation={operation}
                onNextOperation={(type) => {
                  const nextType = findNextOperationType(type);
                  if (nextType) {
                    onConfigureOperation(nextType);
                  }
                }}
                showPipelineDetails
                pipelineMetadata={operation.metadata?.pipeline_summary}
              />
              {processedPipelineSteps
                .filter(step => step?.operationId === operation.operation_id)
                .map((stepData) => (
                  <React.Fragment key={stepData.key}>
                    {renderStageTicket(stepData.operation)}
                  </React.Fragment>
                ))}
            </div>
          );
        }

        return (
          <Card key={`${operation.operation_id}-pipeline`}>
            <PipelineSummaryCard
              pipelineOperation={operation}
              operationTypes={operationTypes}
              onNextOperation={(type) => {
                const nextType = findNextOperationType(type);
                if (nextType) {
                  onConfigureOperation(nextType);
                }
              }}
              index={index}
            />
            <PipelineCollapsible
              operation={operation}
              operationTypes={operationTypes}
              onConfigureOperation={onConfigureOperation}
              resolveOperationType={resolveOperationType}
            />
          </Card>
        );
      })}
    </div>
  );
}

export default PipelineBoard;
