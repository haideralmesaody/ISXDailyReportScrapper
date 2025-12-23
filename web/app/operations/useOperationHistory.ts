import { useEffect, useState } from "react";

interface OperationHistoryEntry {
  timestamp: number;
  status: string;
  progress: number;
  stepsCompleted: number;
  totalSteps: number;
}

type OperationHistoryMap = Map<string, OperationHistoryEntry[]>;

interface CompletionEvent {
  operationId: string;
  timestamp: number;
}

type ToastFunction = (props: {
  title?: string;
  description?: string;
  variant?: "default" | "destructive";
  duration?: number;
}) => void;

export function useOperationHistory(
  operations: any[] | undefined,
  toast: ToastFunction,
) {
  const [lastCompletionEvent, setLastCompletionEvent] =
    useState<CompletionEvent | null>(null);
  const [operationStateHistory, setOperationStateHistory] =
    useState<OperationHistoryMap>(new Map());

  useEffect(() => {
    if (!operations || operations.length === 0) return;

    const now = Date.now();

    setOperationStateHistory((prev) => {
      const newHistoryMap = new Map(prev);

      operations.forEach((operation) => {
        const operationId = operation.operation_id;
        const currentStatus = operation.status;
        const currentProgress = operation.progress;
        const stepsCompleted =
          operation.steps?.filter((step: any) => step.status === "completed")
            .length || 0;
        const totalSteps = operation.steps?.length || 0;

        const history = newHistoryMap.get(operationId) || [];
        const newHistory = [
          ...history,
          {
            timestamp: now,
            status: currentStatus,
            progress: currentProgress,
            stepsCompleted,
            totalSteps,
          },
        ].slice(-10);

        if (currentStatus === "completed" && stepsCompleted === totalSteps) {
          const previousState = history[history.length - 2];
          if (!previousState || previousState.status !== "completed") {
            setLastCompletionEvent({
              operationId,
              timestamp: now,
            });
          }
        }

        if (currentStatus === "running") {
          if (history.length >= 2) {
            const oneStateAgo = history[history.length - 2];
            if (oneStateAgo) {
              const timeSinceLastProgress = now - oneStateAgo.timestamp;
              const progressStuck =
                oneStateAgo.progress === currentProgress &&
                oneStateAgo.stepsCompleted === stepsCompleted;

              if (progressStuck && timeSinceLastProgress > 30000) {
                toast({
                  title: "Operation Stuck",
                  description: `Operation ${operationId} appears stalled. No progress for 30+ seconds.`,
                  variant: "destructive",
                  duration: 10000,
                });
              }
            }
          }
        }

        newHistoryMap.set(operationId, newHistory);
      });

      return newHistoryMap;
    });
  }, [operations, toast]);

  useEffect(() => {
    if (!lastCompletionEvent) return;

    const now = Date.now();
    if (now - lastCompletionEvent.timestamp > 300000) {
      setLastCompletionEvent(null);
    }
  }, [lastCompletionEvent]);

  return { lastCompletionEvent, operationStateHistory };
}
