import type { CalculateInput } from './api'

export const PENDING_JOURNEY_STORAGE_KEY = 'yj_pending_journey'
export const PENDING_JOURNEY_SCHEMA_VERSION = 1
export const PENDING_JOURNEY_TTL_MS = 24 * 60 * 60 * 1000

export type PendingJourneyIntent = 'view_result' | 'generate_report'

export interface PendingBaziJourney {
  version: typeof PENDING_JOURNEY_SCHEMA_VERSION
  type: 'bazi'
  input: CalculateInput
  anonymousChartId?: string
  displayLabel?: string
  intent: PendingJourneyIntent
  returnPath?: string
  createdAt: number
}

export type PendingJourney = PendingBaziJourney

type JourneyStorage = Pick<Storage, 'getItem' | 'setItem' | 'removeItem'>

function getSessionStorage(): JourneyStorage | null {
  if (typeof window === 'undefined') return null
  return window.sessionStorage
}

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function isValidBaziInput(value: unknown): value is CalculateInput {
  if (!isPlainObject(value)) return false
  const gender = value.gender
  const calendarType = value.calendar_type
  return (
    typeof value.year === 'number' &&
    value.year >= 1900 &&
    value.year <= 2100 &&
    typeof value.month === 'number' &&
    value.month >= 1 &&
    value.month <= 12 &&
    typeof value.day === 'number' &&
    value.day >= 1 &&
    value.day <= 31 &&
    typeof value.hour === 'number' &&
    value.hour >= 0 &&
    value.hour <= 23 &&
    (gender === 'male' || gender === 'female') &&
    (calendarType === undefined || calendarType === 'solar' || calendarType === 'lunar')
  )
}

export function createPendingBaziJourney(params: {
  input: CalculateInput
  anonymousChartId?: string
  displayLabel?: string
  intent?: PendingJourneyIntent
  returnPath?: string
  now?: number
}): PendingBaziJourney {
  return {
    version: PENDING_JOURNEY_SCHEMA_VERSION,
    type: 'bazi',
    input: params.input,
    anonymousChartId: params.anonymousChartId || undefined,
    displayLabel: params.displayLabel || undefined,
    intent: params.intent ?? 'view_result',
    returnPath: params.returnPath || undefined,
    createdAt: params.now ?? Date.now(),
  }
}

export function isValidPendingJourney(value: unknown, now = Date.now()): value is PendingJourney {
  if (!isPlainObject(value)) return false
  if (value.version !== PENDING_JOURNEY_SCHEMA_VERSION) return false
  if (value.type !== 'bazi') return false
  if (!isValidBaziInput(value.input)) return false
  if (value.intent !== 'view_result' && value.intent !== 'generate_report') return false
  if (typeof value.createdAt !== 'number') return false
  if (now - value.createdAt > PENDING_JOURNEY_TTL_MS) return false
  return true
}

export function savePendingJourney(journey: PendingJourney, storage: JourneyStorage | null = getSessionStorage()) {
  storage?.setItem(PENDING_JOURNEY_STORAGE_KEY, JSON.stringify(journey))
}

export function clearPendingJourney(storage: JourneyStorage | null = getSessionStorage()) {
  storage?.removeItem(PENDING_JOURNEY_STORAGE_KEY)
}

export function readPendingJourney(storage: JourneyStorage | null = getSessionStorage(), now = Date.now()) {
  const raw = storage?.getItem(PENDING_JOURNEY_STORAGE_KEY)
  if (!raw) return null
  try {
    const parsed = JSON.parse(raw)
    if (isValidPendingJourney(parsed, now)) return parsed
  } catch {
    // Fall through to cleanup below.
  }
  clearPendingJourney(storage)
  return null
}
