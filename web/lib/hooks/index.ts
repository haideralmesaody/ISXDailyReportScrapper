/**
 * Custom React hooks for ISX Daily Reports Scrapper
 * Provides reusable state management and API integration
 */

export { useApi } from './use-api'
export { useToast } from './use-toast'
export { 
  useWebSocket,
  usePipelineUpdates,
  useMarketUpdates,
  useSystemStatus,
  useWebSocketEvent,
  useConnectionStatus
} from './use-websocket'
export {
  useHydration,
  useClientValue,
  withHydration
} from './use-hydration'
export {
  useViewportPosition,
  isViewportZoomed,
  getViewportZoomLevel
} from './use-viewport-position'
export {
  useAdaptiveSize,
  useAdaptiveGridCols,
  useAdaptiveSpacing,
  useAdaptiveTextSize,
  ADAPTIVE_PRESETS
} from './use-adaptive-size'