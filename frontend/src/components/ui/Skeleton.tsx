export interface SkeletonProps {
  /** Number of label + input blocks to imitate. */
  fields?: number;
  label?: string;
}

/**
 * Layout loading state (DESIGN states.md): blocks shaped like the real fields,
 * `animate-pulse` only, no shimmer. Render it only when loading exceeds 1 s.
 */
export default function Skeleton({ fields = 4, label = 'Cargando sección' }: SkeletonProps) {
  return (
    <div aria-busy="true" className="space-y-6">
      <span className="sr-only" role="status">
        {label}
      </span>
      {Array.from({ length: fields }, (_, index) => (
        <div key={index} aria-hidden="true" className="space-y-2 motion-safe:animate-pulse">
          <div className="h-5 w-40 rounded-md bg-line" />
          <div className="h-11 w-full rounded-md bg-line" />
        </div>
      ))}
    </div>
  );
}
