/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * OrganizationSwitcher — Dropdown for switching between organizations.
 */

import { createElement, useState } from "react";
import type { Organization } from "../../types/organization";

export interface OrganizationSwitcherProps {
  organizations: Organization[];
  currentOrgId: string | null;
  onSelect: (orgId: string | null) => void;
  onCreateNew: () => void;
}

export function OrganizationSwitcher({
  organizations,
  currentOrgId,
  onSelect,
  onCreateNew,
}: OrganizationSwitcherProps): JSX.Element {
  const [isOpen, setIsOpen] = useState(false);
  const currentOrg = organizations.find((o) => o.id === currentOrgId);

  return createElement(
    "div",
    { className: "relative" },
    // Trigger button
    createElement(
      "button",
      {
        type: "button",
        className:
          "flex w-[200px] items-center justify-between rounded-lg border px-3 py-2 text-sm hover:bg-muted",
        onClick: () => setIsOpen(!isOpen),
        "aria-expanded": isOpen,
        "aria-haspopup": "listbox",
      },
      createElement(
        "div",
        { className: "flex items-center gap-2" },
        createElement("span", { className: "text-muted-foreground" }, "🏢"),
        createElement(
          "span",
          { className: "truncate" },
          currentOrg?.name || "Personal",
        ),
      ),
      createElement("span", { className: "text-muted-foreground" }, "▾"),
    ),
    // Dropdown menu
    isOpen
      ? createElement(
          "div",
          {
            className:
              "absolute left-0 top-full z-50 mt-1 w-[200px] overflow-hidden rounded-lg border bg-popover shadow-md",
            role: "listbox",
          },
          // Personal option
          createElement(
            "button",
            {
              type: "button",
              className: `flex w-full items-center gap-2 px-3 py-2 text-sm hover:bg-accent ${
                currentOrgId === null ? "bg-accent" : ""
              }`,
              role: "option",
              "aria-selected": currentOrgId === null,
              onClick: () => {
                onSelect(null);
                setIsOpen(false);
              },
            },
            createElement("span", { className: "text-muted-foreground" }, "🏢"),
            "Personal",
          ),
          // Separator
          createElement("div", { className: "-mx-1 my-1 h-px bg-muted" }),
          // Organization options
          ...organizations.map((org) =>
            createElement(
              "button",
              {
                key: org.id,
                type: "button",
                className: `flex w-full items-center gap-2 px-3 py-2 text-sm hover:bg-accent ${
                  currentOrgId === org.id ? "bg-accent" : ""
                }`,
                role: "option",
                "aria-selected": currentOrgId === org.id,
                onClick: () => {
                  onSelect(org.id);
                  setIsOpen(false);
                },
              },
              createElement(
                "span",
                { className: "text-muted-foreground" },
                "🏢",
              ),
              createElement("span", { className: "truncate" }, org.name),
            ),
          ),
          // Separator
          createElement("div", { className: "-mx-1 my-1 h-px bg-muted" }),
          // Create new
          createElement(
            "button",
            {
              type: "button",
              className:
                "flex w-full items-center gap-2 px-3 py-2 text-sm hover:bg-accent",
              onClick: () => {
                onCreateNew();
                setIsOpen(false);
              },
            },
            createElement("span", { className: "text-muted-foreground" }, "+"),
            "Create Organization",
          ),
        )
      : null,
    // Click-away overlay
    isOpen
      ? createElement("div", {
          className: "fixed inset-0 z-40",
          onClick: () => setIsOpen(false),
        })
      : null,
  );
}
