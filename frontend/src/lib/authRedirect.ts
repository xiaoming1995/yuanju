export const DEFAULT_POST_AUTH_TARGET = '/profile'

const BLOCKED_USER_ROUTE_PREFIXES = ['/admin']
const FALLBACK_ALLOWED_ROUTES = [
  '/',
  '/result',
  '/history',
  '/profile',
  '/compatibility',
  '/settings/brand',
  '/bazi/',
]

function isAllowedUserRoute(pathname: string) {
  return FALLBACK_ALLOWED_ROUTES.some(prefix => (
    prefix === '/' ? pathname === '/' : pathname === prefix || pathname.startsWith(prefix)
  ))
}

function stripHashOnlyTarget(target: string) {
  return target.startsWith('/#') ? '/' : target
}

export function getSafePostAuthTarget(
  rawTarget: string | null | undefined,
  fallback = DEFAULT_POST_AUTH_TARGET,
) {
  if (!rawTarget) return fallback
  const target = stripHashOnlyTarget(rawTarget.trim())
  if (!target.startsWith('/') || target.startsWith('//') || target.includes('\\')) return fallback

  let url: URL
  try {
    url = new URL(target, 'https://yuanju.local')
  } catch {
    return fallback
  }

  if (url.origin !== 'https://yuanju.local') return fallback
  if (BLOCKED_USER_ROUTE_PREFIXES.some(prefix => url.pathname === prefix || url.pathname.startsWith(`${prefix}/`))) {
    return fallback
  }
  if (!isAllowedUserRoute(url.pathname)) {
    return fallback
  }

  return `${url.pathname}${url.search}${url.hash}`
}

export function getNextTargetFromSearch(search: string, fallback = DEFAULT_POST_AUTH_TARGET) {
  const params = new URLSearchParams(search)
  return getSafePostAuthTarget(params.get('next'), fallback)
}

export function buildAuthPath(authPath: '/login' | '/register', nextTarget?: string | null) {
  const safeNext = nextTarget ? getSafePostAuthTarget(nextTarget, '') : ''
  if (!safeNext) return authPath
  return `${authPath}?next=${encodeURIComponent(safeNext)}`
}

export function resolvePostAuthTarget(options: {
  search?: string
  stateNext?: unknown
  pendingReturnPath?: string
  fallback?: string
}) {
  const fallback = options.fallback ?? DEFAULT_POST_AUTH_TARGET
  if (typeof options.stateNext === 'string') {
    const stateTarget = getSafePostAuthTarget(options.stateNext, '')
    if (stateTarget) return stateTarget
  }
  if (options.search) {
    const searchTarget = getNextTargetFromSearch(options.search, '')
    if (searchTarget) return searchTarget
  }
  if (options.pendingReturnPath) {
    const pendingTarget = getSafePostAuthTarget(options.pendingReturnPath, '')
    if (pendingTarget) return pendingTarget
  }
  return fallback
}
