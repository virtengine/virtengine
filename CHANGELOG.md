
<!--
Guiding Principles:

Changelogs are for humans, not machines.
There should be an entry for every single version.
The same types of changes should be grouped.
Versions and sections should be linkable.
The latest version comes first.
The release date of each version is displayed.
Mention whether you follow Semantic Versioning.

Usage:

Change log entries are to be added to the Unreleased section under the
appropriate stanza (see below). Each entry should ideally include a tag and
the Github issue reference in the following format:

* (<tag>) \#<issue-number> message

The issue numbers will later be link-ified during the release process so you do
not have to worry about including a link manually, but you can if you wish.

Types of changes (Stanzas):

"Features" for new features.
"Improvements" for changes in existing functionality.
"Deprecated" for soon-to-be removed features.
"Bug Fixes" for any bug fixes.
"Client Breaking" for breaking CLI commands and REST routes used by end-users.
"API Breaking" for breaking exported APIs used by developers building on SDK.
"State Machine Breaking" for any changes that result in a different AppState given same genesisState and txList.

Ref: https://keepachangelog.com/en/1.0.0/
-->

# Changelog

## [Unreleased]

### Client Breaking

* (sdk) Replace ACT/vACT conversion commands with `mint-vcc` and `burn-vcc`; wallet credit metadata uses VCC / `uvcc` and native fee/staking metadata uses VE / `uve`.

### API Breaking

* (sdk) Rename BME mint/burn messages and RPCs to VCC, the vault response to `vault_native`, and the oracle parameter to `native_price_feed_id`. Funding-registry and inventory digests change. Old clients, signed requests and digest-bound authorizations require a coordinated upgrade.

### State Machine Breaking

* (market) Use `uve` for native liquidity rewards and require explicit VE oracle-feed configuration. Existing ACT/vACT balances are not migrated to `uvcc`; existing networks need an audited state/configuration migration before adopting these identifiers. Conversion handlers remain pending and the rename creates no peg or redemption guarantee.

* (ibc-go) Use ibc v4.4.0 

### Improvements

* (sdk) Bump Cosmos SDK version to [v0.38.3](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.38.3)
* (sdk) Bump Cosmos SDK version to [v0.53.5](https://github.com/cosmos/cosmos-sdk/releases/tag/v0.53.5)

### Bug Fixes

* Fix bug in ditribution and querying rewards
