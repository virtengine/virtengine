# Version Control and Branching

This document covers the version-control practices that match the current VirtEngine repository state. For release automation and publication rules, see [RELEASE.md](../RELEASE.md).

## Current Branch Reality

The repository currently uses `main` as the active integration branch for code, docs, and release preparation.

Important notes:

- public documentation should treat `main` as the current source branch;
- some workflows and helper scripts still contain legacy references to `mainnet/main`;
- those legacy references should not be presented as proof that there is an active stable-branch release model in use today.

## Tags and Release Utilities

VirtEngine uses semantic-versioned tags and includes helper scripts for working with them:

```bash
./script/semver.sh validate v0.10.0
./script/mainnet-from-tag.sh v0.10.0
./script/is_prerelease.sh v0.10.0-rc.1
./script/semver.sh bump patch v0.10.0
./script/semver.sh bump minor v0.10.0
./script/semver.sh bump major v0.10.0
```

These utilities help classify and manipulate tags. They do not replace the actual release and launch approval process.

## Working Model

Contributors should assume this workflow unless a release manager documents an exception:

1. branch from `main`;
2. merge back through normal review and CI on `main`;
3. cut release tags from the approved target commit;
4. run the manual release workflow;
5. describe production status only according to the checked-in verification and go/no-go evidence.

## Legacy Stable-Branch References

If you encounter references to `mainnet/main` in automation or historical documentation, treat them as compatibility or migration remnants unless a current release document explicitly reactivates that branch model.

Do not:

- instruct contributors to merge into `mainnet/main` as a standard step;
- describe `mainnet/main` as the current stable release branch;
- assume even or odd minor numbering alone defines the real support state of a release.

## Documentation Sync Rule

When branch or release policy changes, update these files together:

- [README.md](../README.md)
- [RELEASE.md](../RELEASE.md)
- [docs/COMPATIBILITY.md](../docs/COMPATIBILITY.md)
- [VERIFICATION.md](../VERIFICATION.md)
- this document

That keeps the public branch, release, compatibility, and verification story aligned.

## Related Documentation

- [RELEASE.md](../RELEASE.md)
- [README.md](../README.md)
- [docs/COMPATIBILITY.md](../docs/COMPATIBILITY.md)
- [VERIFICATION.md](../VERIFICATION.md)
