# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog][],
and this project adheres to [Semantic Versioning][].

<!--
## Unreleased

### Added
### Changed
### Removed
[0.1.1]: https://github.com/WoozyMasta/discord-a2s-bot/compare/v0.1.0...v0.1.1
-->

## [0.1.3][] - 2025-08-07

### Added

* Template helpers: `RoundUp`, `RoundDown`, `Clamp` to reduce frequent
  channel updates when showing values like player counts
* Added extended logs to `editChannel` to warn when rate limit is reached

## Changed

* Replaced FNV-1a (`hash/fnv`) with `xxh3` (`github.com/zeebo/xxh3`) for
  hashing rendered channel and category templates. Reduces chance of hash
  collisions and improves performance of hash comparison.

[0.1.3]: https://github.com/WoozyMasta/discord-a2s-bot/compare/v0.1.2...v0.1.3

## [0.1.2][] - 2025-07-27

### Added

* Certificate authority certificates to container image (Close #1)

### Changed

* Update direct go dependencies

[0.1.2]: https://github.com/WoozyMasta/discord-a2s-bot/compare/v0.1.1...v0.1.2

## [0.1.1][] - 2025-04-17

### Changed

* Updated go version and dependencies
* Fixed a bug in the yaml configuration example that blocked updating
  of the channel description in `category_name`

[0.1.1]: https://github.com/WoozyMasta/discord-a2s-bot/compare/v0.1.0...v0.1.1

## [0.1.0][] - 2025-01-29

### Added

* First public release

[0.1.0]: https://github.com/WoozyMasta/discord-a2s-bot/tree/v0.1.0

<!--links-->
[Keep a Changelog]: https://keepachangelog.com/en/1.1.0/
[Semantic Versioning]: https://semver.org/spec/v2.0.0.html
