import React from "react";
import { Card } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { CompactOperationCard } from "@/components/operations/CompactOperationCard";
import {
  Download,
  FileSpreadsheet,
  BarChart3,
  Zap,
  Workflow,
  Database,
} from "lucide-react";

type OperationType = {
  id: string;
  name: string;
  description: string;
  [key: string]: any;
};

interface LauncherGridProps {
  loading: boolean;
  operationTypes: OperationType[];
  launchGridOpen: boolean;
  busyOperationTypes: Set<string>;
  selectedOperation: OperationType | null;
  selectedCardRef: React.RefObject<HTMLDivElement>;
  startingOperation: string | null;
  onToggleLaunchers: () => void;
  onConfigure: (type: OperationType) => void;
  onDirectStart: (type: OperationType) => void;
}

const operationIcons = {
  scraping: Download,
  processing: FileSpreadsheet,
  indices: BarChart3,
  liquidity: Zap,
  full_pipeline: Workflow,
  data_processing: Database,
} as const;

function SkeletonCard() {
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

export function LauncherGrid({
  loading,
  operationTypes,
  launchGridOpen,
  busyOperationTypes,
  selectedOperation,
  selectedCardRef,
  startingOperation,
  onToggleLaunchers,
  onConfigure,
  onDirectStart,
}: LauncherGridProps) {
  return (
    <div className="mb-8">
      <div className="flex flex-wrap items-start justify-between gap-3 mb-3">
        <div>
          <h2 className="text-xl font-semibold">Start New Operation</h2>
          <p className="text-sm text-muted-foreground">
            Launch individual stages. Use Quick Start for defaults or configure before running.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="ghost" size="sm" onClick={onToggleLaunchers}>
            {launchGridOpen ? "Hide Launchers" : "Show Launchers"}
          </Button>
        </div>
      </div>
      {launchGridOpen ? (
        <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
          {loading
            ? Array.from({ length: 6 }).map((_, index) => (
                <SkeletonCard key={`launcher-skeleton-${index}`} />
              ))
            : operationTypes.map((type) => {
                const Icon =
                  operationIcons[type.id as keyof typeof operationIcons] ||
                  Database;
                const needsDates =
                  type.id === "scraping" || type.id === "full_pipeline";
                const isBusy = busyOperationTypes.has(type.id);
                const statusBadge = isBusy ? "In Progress" : undefined;

                return (
                  <div
                    key={type.id}
                    ref={
                      selectedOperation?.id === type.id ? selectedCardRef : null
                    }
                  >
                    <CompactOperationCard
                      compact
                      type={{
                        ...type,
                        requiresDates: needsDates,
                      }}
                      icon={Icon}
                      onConfigure={() => onConfigure(type)}
                      onDirectStart={
                        isBusy ? undefined : () => onDirectStart(type)
                      }
                      isStarting={startingOperation === type.id}
                      statusBadge={statusBadge}
                      footerMessage={
                        isBusy
                          ? "This operation already has a run in progress."
                          : undefined
                      }
                    />
                  </div>
                );
              })}
        </div>
      ) : (
        <Card className="p-4 text-sm text-muted-foreground">
          <p>
            Launch controls are hidden to keep the focus on monitoring. Use the
            toggle above whenever you need to start a new operation.
          </p>
        </Card>
      )}
    </div>
  );
}

export default LauncherGrid;
