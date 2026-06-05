import type { HealthStatus } from "../types";

type HealthBadgeProps = {
  status: HealthStatus;
};

export function HealthBadge({ status }: HealthBadgeProps) {
  return <span className={healthClassName(status)}>{healthText(status)}</span>;
}

function healthText(status: HealthStatus): string {
  switch (status) {
    case "checking":
      return "检查中";
    case "ok":
      return "可用";
    case "failed":
      return "不可用";
    default:
      return "未检查";
  }
}

function healthClassName(status: HealthStatus): string {
  const base = "rounded-full px-2 py-1 text-xs font-semibold";

  switch (status) {
    case "ok":
      return `${base} bg-green-50 text-green-700`;
    case "failed":
      return `${base} bg-amber-50 text-amber-700`;
    case "checking":
      return `${base} bg-blue-50 text-blue-700`;
    default:
      return `${base} bg-zinc-100 text-zinc-600`;
  }
}
