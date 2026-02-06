/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * OrganizationList — Lists all organizations the user belongs to.
 */

import { createElement } from "react";
import type { Organization, OrganizationRole } from "../../types/organization";

export interface OrganizationListProps {
  organizations: Organization[];
  isLoading: boolean;
  error: string | null;
  currentUserAddress?: string;
  onSelect: (orgId: string) => void;
  onCreateNew: () => void;
  getUserRole?: (org: Organization) => OrganizationRole | null;
}

function OrganizationListSkeleton(): JSX.Element {
  return createElement(
    "div",
    { className: "space-y-4" },
    ...Array.from({ length: 3 }, (_, i) =>
      createElement("div", {
        key: i,
        className: "h-24 animate-pulse rounded-lg border bg-muted",
      }),
    ),
  );
}

function EmptyState({ onCreateNew }: { onCreateNew: () => void }): JSX.Element {
  return createElement(
    "div",
    {
      className: "flex flex-col items-center justify-center py-16 text-center",
    },
    createElement(
      "div",
      { className: "rounded-full bg-muted p-4" },
      createElement("span", { className: "text-4xl" }, "🏢"),
    ),
    createElement(
      "h2",
      { className: "mt-4 text-lg font-medium" },
      "No organizations",
    ),
    createElement(
      "p",
      { className: "mt-2 text-sm text-muted-foreground" },
      "Create an organization to manage team deployments",
    ),
    createElement(
      "button",
      {
        type: "button",
        className:
          "mt-4 rounded-lg bg-primary px-4 py-2 text-sm text-primary-foreground hover:bg-primary/90",
        onClick: onCreateNew,
      },
      "Create Organization",
    ),
  );
}

export function OrganizationList({
  organizations,
  isLoading,
  error,
  onSelect,
  onCreateNew,
  getUserRole,
}: OrganizationListProps): JSX.Element {
  if (isLoading) {
    return createElement(OrganizationListSkeleton);
  }

  if (error) {
    return createElement(
      "div",
      {
        className:
          "rounded-lg border border-destructive/50 bg-destructive/10 p-4",
        role: "alert",
      },
      createElement(
        "p",
        { className: "text-sm text-destructive" },
        `Failed to load organizations: ${error}`,
      ),
    );
  }

  return createElement(
    "div",
    { className: "space-y-4" },
    createElement(
      "div",
      { className: "flex items-center justify-between" },
      createElement(
        "h2",
        { className: "text-xl font-semibold" },
        "Organizations",
      ),
      createElement(
        "button",
        {
          type: "button",
          className:
            "rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90",
          onClick: onCreateNew,
        },
        "+ New Organization",
      ),
    ),
    organizations.length === 0
      ? createElement(EmptyState, { onCreateNew })
      : createElement(
          "div",
          { className: "grid gap-4 md:grid-cols-2 lg:grid-cols-3" },
          ...organizations.map((org) =>
            createElement(OrganizationListCard, {
              key: org.id,
              organization: org,
              role: getUserRole?.(org) ?? null,
              onSelect: () => onSelect(org.id),
            }),
          ),
        ),
  );
}

interface OrganizationListCardProps {
  organization: Organization;
  role: OrganizationRole | null;
  onSelect: () => void;
}

function OrganizationListCard({
  organization,
  role,
  onSelect,
}: OrganizationListCardProps): JSX.Element {
  return createElement(
    "button",
    {
      type: "button",
      className:
        "w-full rounded-lg border border-border bg-card p-4 text-left transition-all hover:border-primary hover:shadow-md",
      onClick: onSelect,
    },
    createElement(
      "div",
      { className: "flex items-center gap-3" },
      createElement(
        "div",
        {
          className:
            "flex h-10 w-10 items-center justify-center rounded-full bg-primary/10",
        },
        createElement(
          "span",
          { className: "text-lg font-semibold text-primary" },
          organization.name.charAt(0).toUpperCase(),
        ),
      ),
      createElement(
        "div",
        { className: "min-w-0 flex-1" },
        createElement(
          "h3",
          { className: "truncate font-medium" },
          organization.name,
        ),
        createElement(
          "p",
          { className: "truncate text-sm text-muted-foreground" },
          organization.description || "No description",
        ),
      ),
    ),
    createElement(
      "div",
      {
        className: "mt-3 flex items-center gap-4 text-xs text-muted-foreground",
      },
      createElement("span", null, `${organization.totalWeight} members`),
      role
        ? createElement(
            "span",
            {
              className:
                "rounded-full bg-secondary px-2 py-0.5 text-xs capitalize text-secondary-foreground",
            },
            role,
          )
        : null,
      createElement(
        "span",
        { className: "ml-auto" },
        `Created ${organization.createdAt.toLocaleDateString()}`,
      ),
    ),
  );
}
