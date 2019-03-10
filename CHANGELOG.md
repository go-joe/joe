# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- Add a lot more unit tests
- Add `TestBot.Start()` and `TestBot.Stop()`to ease synchronously starting and stopping bot in unit tests
- Add `TestBot.EmitSync(…)` to emit events synchronously in unit tests 

### Changed
- Remove obsolete context argument from `NewTest(…)` function
- Errors from passing invalid expressions to `Bot.Respond(…)` are now returned in `Bot.Run()`
- Events are now processed in the exact same order in which they are emitted
- All pending events are now processed before the brain event loop returns
- Replace context argument from `Brain.HandleEvents()` with new `Brain.Shutdown()` function
- `Adapter` interface was simplified again to directly use the `Brain`
- Remove unnecessary `t` argument from `TestBot.EmitSync(…)` function

### Removed
- Deleted `Brain.Close()` because it was not actually meant to be used to close the brain and is thus confusing

## [v0.1.0] - 2019-03-03

Initial release, note that Joe is still in alpha and the API is not yet considered
stable before the v1.0.0 release.

[Unreleased]: https://github.com/go-joe/joe/compare/v0.1.0...HEAD
[v0.1.0]: https://github.com/go-joe/joe/releases/tag/v0.1.0
