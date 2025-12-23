/**
 * LiquidityStageCard
 * Displays liquidity calculation progress with tickers analyzed and liquidity bucket telemetry.
 * Mirrors the visual style of other stage cards (Scraping, Processing, Indices).
 */

"use client";

import React from "react";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { Progress } from "@/components/ui/progress";
import { Droplets, TrendingUp, TrendingDown, Minus } from "lucide-react";

interface LiquidityStageCardProps {
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

export function LiquidityStageCard({
    metadata,
    statusMessage,
    isComplete,
    isFailed,
    variant = "progress",
    className,
}: LiquidityStageCardProps) {
    // Parse telemetry from backend
    const tickersAnalyzed = toNumber(metadata?.tickers_analyzed) ?? 0;
    const totalFiles = toNumber(metadata?.total_files);
    const filesProcessed = toNumber(metadata?.files_processed);
    const currentFile = metadata?.current_file;
    const stageProgress =
        Math.max(0, Math.min(100, toNumber(metadata?.progress_percent) ?? toNumber(metadata?.progress) ?? 0));

    // Liquidity buckets
    const buckets = metadata?.liquidity_buckets as Record<string, number> | undefined;
    const highLiquidity = toNumber(buckets?.high) ?? 0;
    const mediumLiquidity = toNumber(buckets?.medium) ?? 0;
    const lowLiquidity = toNumber(buckets?.low) ?? 0;

    const hasProgress = typeof filesProcessed === 'number' && typeof totalFiles === 'number' && totalFiles > 0;
    const hasBuckets = highLiquidity > 0 || mediumLiquidity > 0 || lowLiquidity > 0;

    return (
        <div className={cn("space-y-3", className)}>
            {/* Overall Stage Progress */}
            <div className="space-y-1">
                <div className="flex justify-between text-xs text-muted-foreground">
                    <span>Stage Progress</span>
                    <span className="font-medium text-foreground">{Math.round(stageProgress)}%</span>
                </div>
                <Progress value={stageProgress} className="h-2" />
                {hasProgress && (
                    <div className="flex justify-between text-[11px] text-muted-foreground">
                        <span>Input files</span>
                        <span>
                            {filesProcessed}/{totalFiles}
                        </span>
                    </div>
                )}
                {currentFile && (
                    <div className="text-[11px] text-muted-foreground truncate">
                        Currently: {String(currentFile).split('/').pop()?.split('\\').pop() || String(currentFile)}
                    </div>
                )}
            </div>

            {/* Liquidity Stats - Only show when we have data */}
            {(tickersAnalyzed > 0 || hasBuckets) && (
                <div className="grid grid-cols-4 gap-2">
                    <div className={cn(
                        "flex flex-col items-center p-2 rounded-lg",
                        isComplete ? "bg-emerald-50 dark:bg-emerald-900/20" : "bg-blue-50 dark:bg-blue-900/20"
                    )}>
                        <Droplets className="h-4 w-4 text-blue-500 mb-1" />
                        <div className="text-lg font-bold text-foreground">{tickersAnalyzed}</div>
                        <div className="text-[10px] text-muted-foreground">Tickers</div>
                    </div>
                    <div className="flex flex-col items-center p-2 rounded-lg bg-green-50 dark:bg-green-900/20">
                        <TrendingUp className="h-4 w-4 text-green-500 mb-1" />
                        <div className="text-lg font-bold text-green-600">{highLiquidity}</div>
                        <div className="text-[10px] text-muted-foreground">High</div>
                    </div>
                    <div className="flex flex-col items-center p-2 rounded-lg bg-yellow-50 dark:bg-yellow-900/20">
                        <Minus className="h-4 w-4 text-yellow-500 mb-1" />
                        <div className="text-lg font-bold text-yellow-600">{mediumLiquidity}</div>
                        <div className="text-[10px] text-muted-foreground">Medium</div>
                    </div>
                    <div className="flex flex-col items-center p-2 rounded-lg bg-red-50 dark:bg-red-900/20">
                        <TrendingDown className="h-4 w-4 text-red-500 mb-1" />
                        <div className="text-lg font-bold text-red-600">{lowLiquidity}</div>
                        <div className="text-[10px] text-muted-foreground">Low</div>
                    </div>
                </div>
            )}

            {/* Status Message */}
            {statusMessage && (
                <div
                    className={cn(
                        "rounded-md p-3 text-sm",
                        isComplete && "bg-emerald-50 text-emerald-900 dark:bg-emerald-900/20 dark:text-emerald-100",
                        isFailed && "bg-red-50 text-red-900 dark:bg-red-900/20 dark:text-red-100",
                        !isComplete && !isFailed && "bg-blue-50 text-blue-900 dark:bg-blue-900/20 dark:text-blue-100",
                    )}
                >
                    {statusMessage}
                </div>
            )}

            {/* Completion Summary */}
            {variant === "complete" && (
                <div className="flex flex-wrap gap-2">
                    <Badge variant="outline" className="bg-blue-50 text-blue-700 border-blue-200">
                        <Droplets className="h-3 w-3 mr-1" />
                        {tickersAnalyzed} tickers analyzed
                    </Badge>
                    {highLiquidity > 0 && (
                        <Badge variant="outline" className="bg-green-50 text-green-700 border-green-200">
                            {highLiquidity} high liquidity
                        </Badge>
                    )}
                    {lowLiquidity > 0 && (
                        <Badge variant="outline" className="bg-red-50 text-red-700 border-red-200">
                            {lowLiquidity} low liquidity
                        </Badge>
                    )}
                </div>
            )}
        </div>
    );
}

export default LiquidityStageCard;
