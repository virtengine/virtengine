/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * OrganizationCard — Card component for displaying an organization summary.
 */

import { createElement } from "react";
import type { Organization, OrganizationRole } from "../../types/organization";
import { ROLE_LABELS } from "../../types/organization";

export interface OrganizationCardProps {
  organization: Organization;
  role?: OrganizationRole | null;
  deploymentCount?: number;
  onClick?: () => void;
}

export function OrganizationCard({
  organization,
  role,
  deploymentCount,
  onClick,
}: OrganizationCardProps): JSX.Element {
  return createElement(
    "div",
    {
      className:
        "cursor-pointer rounded-lg border border-border bg-card p-4 transition-all hover:border-primary hover:shadow-md",
      onClick,
      role: "button",
      tabIndex: 0,
      onKeyDown: (e: KeyboardEvent) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          onClick?.();
        }
      },
    },
    createElement(
      "div",
      { className: "flex items-center gap-3" },
      createElement(
        "div",
        {
          className:
            "flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-primary/10",
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
        className: "mt-4 flex items-center gap-4 text-sm text-muted-foreground",
      },
      createElement("span", null, `${organization.totalWeight} members`),
      deploymentCount !== undefined
        ? createElement(
            "span",
            null,
            `${deploymentCount} deployment${deploymentCount !== 1 ? "s" : ""}`,
          )
        : null,
      role
        ? createElement(
            "span",
            {
              className:
                "ml-auto rounded-full bg-secondary px-2 py-0.5 text-xs capitalize text-secondary-foreground",
            },
            ROLE_LABELS[role],
          )
        : null,
    ),
  );
}
