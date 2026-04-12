function Skeleton({ className }) {
  return <div className={`animate-pulse rounded-lg bg-gray-100 ${className}`} />
}

export function InsightSkeleton() {
  return (
    <div className="rounded-lg border border-gray-100 p-4 flex items-start gap-3">
      <Skeleton className="w-4 h-4 rounded shrink-0 mt-0.5" />
      <div className="flex-1 flex flex-col gap-2">
        <Skeleton className="h-4 w-3/4" />
        <Skeleton className="h-3 w-full" />
      </div>
    </div>
  )
}

export function GiftCardSkeleton() {
  return (
    <div className="rounded-lg border border-gray-200 p-4 flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <Skeleton className="h-4 w-1/2" />
        <Skeleton className="h-5 w-16 rounded-full" />
      </div>
      <Skeleton className="h-3 w-full" />
      <Skeleton className="h-3 w-4/5" />
      <div className="flex gap-2 mt-1">
        <Skeleton className="h-8 w-20 rounded-md" />
        <Skeleton className="h-8 w-20 rounded-md" />
        <Skeleton className="h-8 w-20 rounded-md" />
      </div>
    </div>
  )
}

export function ResultsSkeleton() {
  return (
    <div className="flex flex-col gap-6">
      <div>
        <Skeleton className="h-5 w-32 mb-3" />
        <div className="flex flex-col gap-2">
          <InsightSkeleton />
          <InsightSkeleton />
        </div>
      </div>
      <div>
        <Skeleton className="h-5 w-24 mb-3" />
        <div className="flex flex-col gap-3">
          <GiftCardSkeleton />
          <GiftCardSkeleton />
          <GiftCardSkeleton />
        </div>
      </div>
    </div>
  )
}
