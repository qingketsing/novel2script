type AppHeaderProps = {
  statusLabel: string;
};

export function AppHeader({ statusLabel }: AppHeaderProps) {
  return (
    <header className="flex flex-col justify-between gap-3 border-b border-zinc-200 pb-4 md:flex-row md:items-end">
      <div>
        <h1 className="mt-2 text-3xl font-semibold tracking-tight text-zinc-950 sm:text-4xl">
Novel2Script
        </h1>
      </div>
      <div className="rounded-lg border border-zinc-200 bg-white px-3 py-2 text-sm text-zinc-600">
        {statusLabel}
      </div>
    </header>
  );
}
