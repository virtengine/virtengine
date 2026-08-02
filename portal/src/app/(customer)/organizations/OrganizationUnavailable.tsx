export const organizationUnavailable = {
  code: 'feature_unavailable',
  message: 'Organization support is unavailable for this deployment.',
} as const;

export function OrganizationUnavailable() {
  return (
    <section
      className="rounded-lg border border-border bg-card p-6"
      role="status"
      aria-labelledby="organization-unavailable-title"
      aria-describedby="organization-unavailable-description"
      data-error-code={organizationUnavailable.code}
    >
      <p className="text-sm font-medium text-muted-foreground">{organizationUnavailable.code}</p>
      <h2 id="organization-unavailable-title" className="mt-2 text-lg font-semibold">
        Organizations unavailable
      </h2>
      <p id="organization-unavailable-description" className="mt-2 text-sm text-muted-foreground">
        {organizationUnavailable.message}
      </p>
    </section>
  );
}
