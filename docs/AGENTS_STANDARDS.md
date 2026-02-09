# AGENTS.md Standards

This guide defines when and how to create AGENTS.md files across VirtEngine.

## When to create a new AGENTS.md
- New module or subsystem with its own entry points (new top-level folder or `x/<module>`).
- Major refactor that introduces new architecture or replaces core flows.
- Any area that onboards new contributors or requires operational context.

## Required detail level
- Explain purpose and scope in 3-5 bullets.
- Document key entry points with file:line references.
- Provide at least one usage example (CLI, API, or code snippet).
- Call out extension points and anti-patterns.
- Include configuration defaults and testing commands.

## Structuring complex architectures
- Use a high-level architecture section before deep details.
- Prefer diagrams for multi-step flows (Mermaid flowchart or sequence diagram).
- Break down subsystems into bullet lists instead of long paragraphs.
- Keep optional sections (API Reference, Dependencies) after required sections.

## Code reference best practices
- Use repo-relative `path/file.ext:line` format.
- Point to real entry points and public APIs, not test-only files.
- Update references when files move or line numbers change.

## Keeping docs synchronized with code
- Update AGENTS.md in the same PR as code changes.
- Add a validation run to your local checklist:
  - `node scripts/validate-agents-docs.mjs`
- Keep AGENTS_INDEX updated with any new module docs.

## Review checklist for documentation PRs
- [ ] Required sections are present and complete.
- [ ] Code references are valid and up to date.
- [ ] Internal links resolve and anchors work.
- [ ] Examples compile or clearly indicate pseudo-code.
- [ ] Troubleshooting section contains at least one real issue.
- [ ] AGENTS_INDEX includes new or moved AGENTS.md files.

## Tooling
- Template: `docs/templates/AGENTS.template.md`
- Index: `docs/AGENTS_INDEX.md`
- Validation script: `scripts/validate-agents-docs.mjs`

## Optional enhancements
- Add a table of contents if the file exceeds two screens.
- Keep a VS Code snippet library in your editor for repeated sections.
