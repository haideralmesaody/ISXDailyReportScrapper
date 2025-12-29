import React, { useMemo, useCallback } from "react";
import ScrapingOperationTicket from "@/components/operations/ScrapingOperationTicket";
import ProcessingOperationTicket from "@/components/operations/ProcessingOperationTicket";
import { OperationCompleteCard } from "@/components/operations/OperationCompleteCard";
import { IndicesStageCard } from "@/components/operations/IndicesStageCard";
import { LiquidityStageCard } from "@/components/operations/LiquidityStageCard";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import StageTelemetryTicket from "@/components/operations/StageTelemetryTicket";
import type { PipelineContext } from "@/lib/operations/pipeline-context";

interface OperationTicketsProps {
  operations: any[];
  operationTypes: any[];
  resolveOperationType: (op: any) => string | undefined;
  getScrapingTelemetry?: (operationId: string) => any;
  onConfigureOperation: (type: any) => void;
  pipelineContext?: PipelineContext | null;
}

export function OperationTickets({
  operations,
  operationTypes,
  resolveOperationType,
  getScrapingTelemetry,
  onConfigureOperation,
  pipelineContext,
}: OperationTicketsProps) {
  // Memoize operations processing to prevent TDZ errors
  const processedOperations = useMemo(() => {
    if (!operations || operations.length === 0) {
      return [];
    }

    return operations.map((operation) => {
      const opType = resolveOperationType(operation);
      const scrapingData = getScrapingTelemetry
        ? getScrapingTelemetry(operation.operation_id)
        : undefined;

      return { operation, opType, scrapingData };
    });
  }, [operations, resolveOperationType, getScrapingTelemetry]);

  const resolveStageLabel = useCallback((stageId: string, fallbackNumber?: number) => {
    const total = pipelineContext?.totalStages || (Array.isArray(operationTypes) ? operationTypes.length : 0) || undefined;
    const number = pipelineContext?.stageNumberById?.[stageId] ?? fallbackNumber;
    if (!number || !total) return "Stage";
    return `Stage ${number} of ${total}`;
  }, [pipelineContext, operationTypes]);

  const renderIndicesTicket = useCallback((operation: any) => {
    const step = Array.isArray(operation.steps) ? operation.steps[0] : operation;
    const metadata = step?.metadata || operation?.metadata || {};

    return (
      <Card className="overflow-hidden">
        <CardHeader className="pb-3">
          <div className="flex items-start justify-between">
            <div>
              <p className="text-sm text-muted-foreground">Index Extraction</p>
              <h3 className="text-lg font-semibold">{resolveStageLabel("indices", 3)}</h3>
            </div>
            <Badge variant={operation.status === "completed" ? "secondary" : operation.status === "failed" ? "destructive" : "default"}>
              {operation.status || "unknown"}
            </Badge>
          </div>
        </CardHeader>
        <CardContent>
          <IndicesStageCard metadata={metadata} isComplete={operation.status === "completed"} isFailed={operation.status === "failed"} />
        </CardContent>
      </Card>
    );
  }, [resolveStageLabel]);

  const renderStageTicket = useCallback((operation: any, opType: string | undefined, scrapingData?: any) => {
    const stageId =
      opType ||
      operation?.metadata?.stage_id ||
      operation?.current_step ||
      operation?.stage_id ||
      (Array.isArray(operation?.steps) ? operation.steps[0]?.metadata?.stage_id || operation.steps[0]?.id : undefined) ||
      "scraping";

    switch (stageId) {
      case "scraping":
        return (
          <ScrapingOperationTicket
            key={`${operation.operation_id}-scraping-ticket`}
            telemetry={scrapingData ?? operation.metadata ?? {}}
            operation={operation}
            pipelineContext={pipelineContext}
          />
        );
      case "processing":
        return (
          <ProcessingOperationTicket
            key={`${operation.operation_id}-processing-ticket`}
            operation={operation}
            pipelineContext={pipelineContext}
          />
        );
      case "indices":
        return (
          <React.Fragment key={`${operation.operation_id}-indices-ticket`}>
            {renderIndicesTicket(operation)}
          </React.Fragment>
        );
      case "liquidity":
        return (
          <Card key={`${operation.operation_id}-liquidity-ticket`} className="overflow-hidden">
            <CardHeader className="pb-3">
              <div className="flex items-start justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">Liquidity Analysis</p>
                  <h3 className="text-lg font-semibold">{resolveStageLabel("liquidity", 4)}</h3>
                </div>
                <Badge variant={operation.status === "completed" ? "secondary" : operation.status === "failed" ? "destructive" : "default"}>
                  {operation.status || "unknown"}
                </Badge>
              </div>
            </CardHeader>
            <CardContent>
              <LiquidityStageCard
                metadata={(() => {
                  const step = Array.isArray(operation.steps) ? operation.steps[0] : operation;
                  return step?.metadata || operation?.metadata || {};
                })()}
                statusMessage={operation.message}
                isComplete={operation.status === "completed"}
                isFailed={operation.status === "failed"}
              />
            </CardContent>
          </Card>
        );
      default:
        return (
          <StageTelemetryTicket
            key={`${operation.operation_id}-generic-${stageId}`}
            operation={operation}
          />
        );
    }
  }, [renderIndicesTicket, pipelineContext, resolveStageLabel]);

  // Memoize next operation finder to prevent TDZ errors
  const findNextOperation = useCallback((type: string) => {
    try {
      return operationTypes?.find((t) => t.id === type) || null;
    } catch (error) {
      console.warn('Next operation type lookup failed:', error);
      return null;
    }
  }, [operationTypes]);

  // Memoize pipeline steps processing to prevent TDZ errors
  const processedPipelineSteps = useMemo(() => {
    return processedOperations.map(({ operation, opType }) => {
      if (
        operation.status === "completed" &&
        Array.isArray(operation.steps) &&
        operation.steps.length > 1
      ) {
        return operation.steps.map((step, index) => {
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
            stageNumber: index + 1,
            totalStages: operation.steps.length,
            showConnector: index < operation.steps.length - 1
          };
        });
      }
      return null;
    }).filter(Boolean);
  }, [processedOperations]);

  if (!processedOperations.length) {
    return null;
  }

  return (
    <>
      {processedOperations.map(({ operation, opType, scrapingData }) => {
        const key = `${operation.operation_id}-${opType}`;

        // Handle completed pipelines with multiple steps
        if (
          operation.status === "completed" &&
          Array.isArray(operation.steps) &&
          operation.steps.length > 1
        ) {
          return (
            <React.Fragment key={`${operation.operation_id}-pipeline-complete`}>
              <OperationCompleteCard
                operation={operation}
                onNextOperation={(type) => {
                  const nextType = findNextOperation(type);
                  if (nextType) {
                    onConfigureOperation(nextType);
                  }
                }}
                showPipelineDetails
                pipelineMetadata={operation.metadata?.pipeline_summary}
              />
              {processedPipelineSteps
                .filter(step => step?.operation?.operation_id === operation.operation_id)
                .map((stepData) => {
                  const stageOpType = resolveOperationType(stepData.operation);
                  const ticket = renderStageTicket(stepData.operation, stageOpType);
                  return (
                    <React.Fragment key={stepData.key}>
                      {ticket}
                    </React.Fragment>
                  );
                })}
            </React.Fragment>
          );
        }

        const ticket = renderStageTicket(operation, opType, scrapingData);

        if (ticket) {
          if (operation.status === "completed" && opType === "scraping" && scrapingData) {
            return (
              <React.Fragment
                key={`${key}-complete`}
              >
                <OperationCompleteCard
                  operation={operation}
                  onNextOperation={(type) => {
                    const nextType = findNextOperation(type);
                    if (nextType) {
                      onConfigureOperation(nextType);
                    }
                  }}
                />
                {ticket}
              </React.Fragment>
            );
          }
          return ticket;
        }

        return null;
      })}
    </>
  );
}

export default OperationTickets;
