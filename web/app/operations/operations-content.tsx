"use client";

import React, {
  useState,
  useEffect,
  useCallback,
  useRef,
  useMemo,
} from "react";
import { useRouter } from "next/navigation";
import { apiClient } from "@/lib/api";
import { useAllOperationUpdates } from "@/lib/hooks/use-websocket";
import { useHydration } from "@/lib/hooks/use-hydration";
import { OperationRequestBuilder } from "@/lib/api/operation-request-builder";
import { Card, Alert, AlertDescription, Button } from "@/components/ui/safe-components";
import { Skeleton } from "@/components/ui/skeleton";
import OperationTickets from "@/components/operations/OperationTickets";
import LauncherGrid from "@/components/operations/LauncherGrid";
import { FloatingConfigPanel } from "@/components/operations/FloatingConfigPanel";
import { AlertCircle, Loader2, Info } from "lucide-react";
import { HelpButton } from "@/components/guide/HelpButton";
import { useToast } from "@/lib/hooks/use-toast";
import { useOperationHistory } from "@/app/operations/useOperationHistory";
import { OperationErrorBoundary, useErrorBoundary } from "@/components/operations/OperationErrorBoundary";

function OperationSkeletonCard() {
  return (
    <Card className="min-h-[120px] p-4">
      <div className="space-y-3">
        <div className="flex items-center gap-3">
          <Skeleton className="h-8 w-8 rounded-full" />
          <Skeleton className="h-5 w-24" />
        </div>
        <div className="space-y-2">
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-3 w-3/4" />
        </div>
        <div className="flex justify-end pt-2">
          <Skeleton className="h-11 w-20" />
        </div>
      </div>
    </Card>
  );
}

export default function OperationsContent() {
  const router = useRouter();
  const { toast } = useToast();
  const { error: boundaryError, resetError, captureError } = useErrorBoundary();

  // WebSocket hook for real-time updates - wrapped in error boundary
  const {
    operations,
    connected,
    error: wsError,
    getScrapingTelemetry,
  } = useAllOperationUpdates();

  // State - minimal and simple
  const [operationTypes, setOperationTypes] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [startingOperation, setStartingOperation] = useState<string | null>(
    null,
  );
  const [selectedOperation, setSelectedOperation] = useState<any>(null);
  const [configPanelOpen, setConfigPanelOpen] = useState(false);
  const selectedCardRef = useRef<HTMLDivElement>(null);
  const [launchGridOpen, setLaunchGridOpen] = useState(true);

  // Use hydration hook for client-side operations
  const isHydrated = useHydration();

  const resolveOperationType = useCallback((op: any) => {
    return (
      op?.metadata?.operation_type ||
      op?.metadata?.operation ||
      op?.operation_type ||
      op?.type ||
      (op?.name === "Full Pipeline" ? "full_pipeline" : undefined)
    );
  }, []);

  const busyOperationTypes = useMemo(() => {
    const busy = new Set<string>();
    if (!operations) return busy;
    try {
      operations.forEach((op) => {
        const opType = resolveOperationType(op);
        if (!opType) return;
        if (
          op.status &&
          op.status !== "completed" &&
          op.status !== "failed" &&
          op.status !== "cancelled"
        ) {
          busy.add(opType);
        }
      });
      return busy;
    } catch (error) {
      console.warn('Busy operation types calculation failed:', error);
      return busy;
    }
  }, [operations, resolveOperationType]);

  const displayOperations = useMemo(() => {
    if (!operations) return [];
    try {
      return operations.filter((op) => resolveOperationType(op) !== "full_pipeline");
    } catch (error) {
      console.warn('Operations filtering failed:', error);
      return [];
    }
  }, [operations, resolveOperationType]);

  const filteredOperationTypes = useMemo(() => {
    return (operationTypes || []).filter((t) => t.id !== "full_pipeline");
  }, [operationTypes]);

  // Enhanced completion detection and state tracking
  useOperationHistory(operations, toast);

  // Fetch available operation types with error boundary protection
  useEffect(() => {
    const fetchTypes = async () => {
      try {
        setLoading(true);
        const types = await apiClient.getOperationTypes();
        setOperationTypes(types);
      } catch (err) {
        const errorMessage = err instanceof Error
          ? err.message
          : "Failed to fetch operation types";

        setError(errorMessage);
        captureError(err instanceof Error ? err : new Error(errorMessage));
      } finally {
        setLoading(false);
      }
    };

    fetchTypes();
  }, [captureError]);

  // Handle operation configuration
  const handleConfigureOperation = useCallback(
    (type: any) => {
      setSelectedOperation(type);
      setConfigPanelOpen(true);
    },
    [setSelectedOperation, setConfigPanelOpen],
  );

  // Handle quick start for operations that don't need configuration.
  // For the first four stages, call the dedicated stage endpoints; otherwise fall back to createOperation.
  const handleDirectStart = useCallback(
    async (type: any) => {
      if (!isHydrated) return;

      try {
        setStartingOperation(type.id);
        setError(null);

        if (["scraping", "processing", "indices", "liquidity"].includes(type.id)) {
          let stageParams: Record<string, any> = {};
          if (type.id === "scraping") {
            stageParams = {
              mode: "initial",
              from_date: new Date().toISOString().split("T")[0],
              to_date: new Date().toISOString().split("T")[0],
              headless: true,
            };
          }
          console.log("Quick Start stage execution:", type.id, stageParams);
          await apiClient.executeStageDirectly(type.id, stageParams);
        } else {
          // Fallback to legacy operation creation (e.g., full pipeline)
          const params = OperationRequestBuilder.buildQuickStart(type.id);
          console.log(
            "Quick Start operation with validated request:",
            type.id,
            params,
          );
          await apiClient.createOperation(params);
        }
      } catch (err) {
        console.error("Failed to start operation:", err);
        setError(
          err instanceof Error ? err.message : "Failed to start operation",
        );
      } finally {
        setStartingOperation(null);
      }
    },
    [isHydrated, setStartingOperation, setError],
  );

  // Handle direct execution for immediate stage execution (bypass JobQueue)
  const handleDirectExecution = useCallback(
    async (stageId: string, params?: any) => {
      if (!isHydrated) return;

      try {
        setStartingOperation(stageId);
        setError(null);

        // Prepare parameters for direct execution
        let executionParams: Record<string, any> = params || {};

        // Add default parameters for scraping stage
        if (stageId === "scraping") {
          executionParams = {
            mode: "initial",
            from_date:
              params?.from_date ||
              params?.from ||
              new Date().toISOString().split("T")[0],
            to_date:
              params?.to_date ||
              params?.to ||
              new Date().toISOString().split("T")[0],
            headless: params?.headless ?? true,
            ...params,
          };
        }

        console.log("Direct stage execution:", stageId, executionParams);

        const response = await apiClient.executeStageDirectly(
          stageId,
          executionParams,
        );
        console.log("Direct execution started:", response);

        // WebSocket will automatically update with progress
        toast({
          title: "Stage Execution Started",
          description: `${stageId} execution started successfully`,
        });
      } catch (err) {
        console.error("Failed to execute stage directly:", err);
        setError(
          err instanceof Error ? err.message : "Failed to execute stage",
        );
        toast({
          title: "Execution Failed",
          description:
            err instanceof Error ? err.message : "Failed to execute stage",
          variant: "destructive",
        });
      } finally {
        setStartingOperation(null);
      }
    },
    [isHydrated, setStartingOperation, setError],
  );

  // Start operation from floating panel
  const handleStartOperation = useCallback(
    async (params: any) => {
      if (!isHydrated || !selectedOperation) return;

      try {
        setStartingOperation(selectedOperation.id);
        setError(null);

        const stageId = selectedOperation.id;
        // Use stage execution for single stages; fallback to createOperation for full pipeline
        if (stageId !== "full_pipeline") {
          const stageParams = {
            ...(params?.parameters || params || {}),
          };

          if (stageId === "scraping") {
            stageParams.mode =
              stageParams.mode || params?.mode || "initial";
            stageParams.from_date =
              stageParams.from_date ||
              stageParams.from ||
              params?.from_date ||
              params?.from ||
              new Date().toISOString().split("T")[0];
            stageParams.to_date =
              stageParams.to_date ||
              stageParams.to ||
              params?.to_date ||
              params?.to ||
              new Date().toISOString().split("T")[0];
          }

          await apiClient.executeStageDirectly(stageId, stageParams);
        } else {
          await apiClient.createOperation(params);
        }
      } catch (err) {
        setError(
          err instanceof Error ? err.message : "Failed to start operation",
        );
      } finally {
        setStartingOperation(null);
      }
    },
    [isHydrated, selectedOperation, setStartingOperation, setError],
  );

  return (
    <div
      className="min-h-screen px-6 py-4 md:py-6 overflow-x-hidden"
      aria-busy={loading}
    >
      <div className="max-w-screen-2xl mx-auto p-4 md:p-6 space-y-6">
        {/* Show hydration state only if not hydrated AND no data available */}
        {!isHydrated && !loading && !operationTypes.length && (
          <div className="flex items-center justify-center py-16">
            <div className="text-center">
              <Loader2 className="h-8 w-8 animate-spin mx-auto mb-4" />
              <p className="text-muted-foreground">Initializing...</p>
            </div>
          </div>
        )}
        {/* Small loading banner for screen readers */}
        {loading && (
          <div className="sr-only" role="status" aria-live="polite">
            Loading operations...
          </div>
        )}
        {/* Header */}
        <div>
          <div className="flex items-center gap-2 mb-2">
            <h1 className="text-3xl font-bold">Operations</h1>
            <HelpButton
              section="operations"
              tooltip="Learn about operations and pipeline stages"
            />
          </div>
          <p className="text-muted-foreground">
            Manage and monitor data processing operations
          </p>
        </div>

        {/* WebSocket Status */}
        <div className="flex items-center gap-2">
          <div
            className={`h-2 w-2 rounded-full ${connected ? "bg-green-500" : "bg-red-500"}`}
          />
          <span className="text-sm text-muted-foreground">
            {connected ? "Connected" : "Disconnected"}
          </span>
        </div>

        {/* Error Display with Boundary Error Support */}
        {(error || wsError || boundaryError) && (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription className="space-y-2">
              <div>{error || wsError || boundaryError?.message}</div>
              {boundaryError && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={resetError}
                  className="mt-2"
                >
                  Retry
                </Button>
              )}
            </AlertDescription>
          </Alert>
        )}

        <OperationErrorBoundary
          onError={(error, errorInfo) => {
            console.error("LauncherGrid error:", error, errorInfo);
            captureError(error);
          }}
        >
          <LauncherGrid
            loading={loading}
            operationTypes={filteredOperationTypes}
            launchGridOpen={launchGridOpen}
            busyOperationTypes={busyOperationTypes}
            selectedOperation={selectedOperation}
            selectedCardRef={selectedCardRef}
            startingOperation={startingOperation}
            onToggleLaunchers={() => setLaunchGridOpen((prev) => !prev)}
            onConfigure={handleConfigureOperation}
            onDirectStart={handleDirectStart}
          />
        </OperationErrorBoundary>

        {/* Active and Completed Operations */}
        {displayOperations.length > 0 || loading ? (
          <div>
            <h2 className="text-xl font-semibold mb-3">Operations</h2>
            <div className="space-y-6">
              {loading && displayOperations.length === 0 ? (
                // Show skeleton cards for operations while loading
                Array.from({ length: 3 }).map((_, index) => (
                  <OperationSkeletonCard key={`op-skeleton-${index}`} />
                ))
              ) : (
                <OperationErrorBoundary
                  onError={(error, errorInfo) => {
                    console.error("OperationTickets error:", error, errorInfo);
                    captureError(error);
                  }}
                >
                  <OperationTickets
                    operations={displayOperations}
                    operationTypes={filteredOperationTypes}
                    resolveOperationType={resolveOperationType}
                    getScrapingTelemetry={getScrapingTelemetry}
                    onConfigureOperation={handleConfigureOperation}
                  />
                </OperationErrorBoundary>
              )}
            </div>
          </div>
        ) : (
          <>
            {/* Empty State */}
            <Card className="p-6 md:p-8">
              <div className="text-center">
                <Info className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                <h3 className="text-lg font-medium mb-2">
                  No Active Operations
                </h3>
                <p className="text-muted-foreground">
                  Start a new operation to begin processing data
                </p>
              </div>
            </Card>
          </>
        )}

        {/* Floating Configuration Panel */}
        <FloatingConfigPanel
          isOpen={configPanelOpen}
          onClose={() => {
            setConfigPanelOpen(false);
            setSelectedOperation(null);
          }}
          onStart={handleStartOperation}
          operationType={selectedOperation}
          isStarting={startingOperation === selectedOperation?.id}
          anchorRef={selectedCardRef}
        />
      </div>
    </div>
  );
}
