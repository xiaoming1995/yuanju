export function buildBaziResultRoute(chartId: string | undefined, isGuest: boolean): string {
  if (!isGuest && chartId) {
    return `/history/${chartId}`
  }
  return '/result'
}
