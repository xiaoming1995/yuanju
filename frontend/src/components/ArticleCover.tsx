import { useState } from 'react'

interface ArticleCoverProps {
  className: string
  coverUrl?: string
  sourceName?: string
  fallback?: string
  loading?: 'lazy' | 'eager'
}

export default function ArticleCover({
  className,
  coverUrl,
  sourceName,
  fallback = '资',
  loading = 'lazy',
}: ArticleCoverProps) {
  const [failedUrl, setFailedUrl] = useState<string | null>(null)

  const fallbackText = sourceName?.trim().slice(0, 1) || fallback
  const shouldShowImage = Boolean(coverUrl && failedUrl !== coverUrl)

  return (
    <div className={className} aria-hidden="true">
      {shouldShowImage ? (
        <img
          src={coverUrl}
          alt=""
          loading={loading}
          referrerPolicy="no-referrer"
          onError={() => setFailedUrl(coverUrl || null)}
        />
      ) : (
        <span>{fallbackText}</span>
      )}
    </div>
  )
}
