import type { IChartApi, ISeriesApi, Time } from 'lightweight-charts'
import type { DrawingPoint } from './types'

interface BusinessDay {
  year: number
  month: number
  day: number
}

export interface DrawingCoordinate {
  x: number
  y: number
}

const isBusinessDay = (time: Time): time is BusinessDay => {
  return typeof time === 'object' && time !== null && 'year' in time
}

const businessDayToEpochSeconds = (time: BusinessDay): number => {
  const utcMs = Date.UTC(time.year, time.month - 1, time.day)
  return Math.floor(utcMs / 1000)
}

export const timeToCoordinate = (chart: IChartApi, time: number): number | null => {
  return chart.timeScale().timeToCoordinate(time as Time)
}

export const coordinateToTime = (chart: IChartApi, coordinate: number): number | null => {
  const time = chart.timeScale().coordinateToTime(coordinate)
  if (time === null) return null
  if (typeof time === 'number') return time
  if (isBusinessDay(time)) return businessDayToEpochSeconds(time)
  return null
}

export const priceToCoordinate = (
  series: ISeriesApi<'Candlestick'>,
  price: number
): number | null => {
  return series.priceToCoordinate(price)
}

export const coordinateToPrice = (
  series: ISeriesApi<'Candlestick'>,
  coordinate: number
): number | null => {
  const price = series.coordinateToPrice(coordinate)
  return price ?? null
}

export const pointToCoordinate = (
  chart: IChartApi,
  series: ISeriesApi<'Candlestick'>,
  point: DrawingPoint
): DrawingCoordinate | null => {
  const x = timeToCoordinate(chart, point.time)
  const y = priceToCoordinate(series, point.price)
  if (x === null || y === null) return null
  return { x, y }
}

export const coordinateToPoint = (
  chart: IChartApi,
  series: ISeriesApi<'Candlestick'>,
  coordinate: DrawingCoordinate
): DrawingPoint | null => {
  const time = coordinateToTime(chart, coordinate.x)
  const price = coordinateToPrice(series, coordinate.y)
  if (time === null || price === null) return null
  return { time, price }
}
