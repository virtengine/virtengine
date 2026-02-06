/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 *
 * CreateOrganizationDialog — Dialog for creating a new organization.
 */

import { createElement, useState } from "react";
import type { CreateOrganizationRequest } from "../../types/organization";

export interface CreateOrganizationDialogProps {
  open: boolean;
  isPending: boolean;
  onClose: () => void;
  onCreate: (request: CreateOrganizationRequest) => void;
}

export function CreateOrganizationDialog({
  open,
  isPending,
  onClose,
  onCreate,
}: CreateOrganizationDialogProps): JSX.Element | null {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");

  if (!open) return null;

  const handleSubmit = (e: Event) => {
    e.preventDefault();
    if (!name.trim()) return;
    onCreate({
      name: name.trim(),
      description: description.trim() || undefined,
    });
    setName("");
    setDescription("");
  };

  return createElement(
    "div",
    {
      className:
        "fixed inset-0 z-50 flex items-center justify-center bg-black/50",
      onClick: (e: MouseEvent) => {
        if ((e.target as HTMLElement) === e.currentTarget) onClose();
      },
    },
    createElement(
      "div",
      {
        className:
          "w-full max-w-md rounded-lg border bg-background p-6 shadow-lg",
        role: "dialog",
        "aria-modal": true,
        "aria-labelledby": "create-org-dialog-title",
      },
      createElement(
        "h2",
        { id: "create-org-dialog-title", className: "text-lg font-semibold" },
        "Create Organization",
      ),
      createElement(
        "p",
        { className: "mt-1 text-sm text-muted-foreground" },
        "Create an organization to manage team deployments and billing.",
      ),
      createElement(
        "form",
        { className: "mt-4 space-y-4", onSubmit: handleSubmit },
        // Name input
        createElement(
          "div",
          null,
          createElement(
            "label",
            { className: "text-sm font-medium", htmlFor: "org-name" },
            "Organization Name",
          ),
          createElement("input", {
            id: "org-name",
            type: "text",
            className:
              "mt-1 flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2",
            placeholder: "My Organization",
            value: name,
            onChange: (e: Event) =>
              setName((e.target as HTMLInputElement).value),
            required: true,
            maxLength: 100,
          }),
        ),
        // Description input
        createElement(
          "div",
          null,
          createElement(
            "label",
            { className: "text-sm font-medium", htmlFor: "org-description" },
            "Description (optional)",
          ),
          createElement("textarea", {
            id: "org-description",
            className:
              "mt-1 flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2",
            placeholder: "Describe what this organization is for...",
            value: description,
            onChange: (e: Event) =>
              setDescription((e.target as HTMLTextAreaElement).value),
            maxLength: 500,
            rows: 3,
          }),
        ),
        // Info box
        createElement(
          "div",
          { className: "rounded-lg bg-muted/50 p-3" },
          createElement(
            "p",
            { className: "text-xs text-muted-foreground" },
            "You will be the admin of this organization. This will submit a transaction to the x/group module on-chain.",
          ),
        ),
        // Actions
        createElement(
          "div",
          { className: "flex justify-end gap-2 pt-2" },
          createElement(
            "button",
            {
              type: "button",
              className: "rounded-lg border px-4 py-2 text-sm hover:bg-muted",
              onClick: onClose,
            },
            "Cancel",
          ),
          createElement(
            "button",
            {
              type: "submit",
              className:
                "rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50",
              disabled: isPending || !name.trim(),
            },
            isPending ? "Creating..." : "Create Organization",
          ),
        ),
      ),
    ),
  );
}
