/**
 * The Xem assistant mark: an "x" of four tapered petals with a dot in each gap,
 * so it reads as both the Xem initial and a spark. Inherits size and colour.
 */
export function XemMark({
  size = 18,
  className,
}: {
  size?: number;
  className?: string;
  strokeWidth?: number;
}) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="-12 -12 24 24"
      fill="currentColor"
      aria-hidden="true"
      className={className}
    >
      {[45, 135, 225, 315].map((a) => (
        <path key={a} transform={`rotate(${a})`} d="M1 -1.9 Q7 -2.4 11.2 0 Q7 2.4 1 1.9 Q-0.4 0 1 -1.9Z" />
      ))}
      {[0, 90, 180, 270].map((a) => (
        <circle key={a} transform={`rotate(${a})`} cx="7.6" cy="0" r="1.45" />
      ))}
    </svg>
  );
}

/** The mark on its ink tile, for headers and the launcher. */
export function XemMarkTile({ size = 28 }: { size?: number }) {
  return (
    <span
      className="inline-grid shrink-0 place-items-center rounded-[28%] bg-[#18181b] text-[#d4ed55] dark:bg-[#d4ed55] dark:text-[#18181b]"
      style={{ width: size, height: size }}
    >
      <XemMark size={Math.round(size * 0.62)} />
    </span>
  );
}
