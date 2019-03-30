# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
- Add ReceiveMessageEvent.Data field to allow using the underlying message type of the adapters
- Add ReceiveMessageEvent.AuthorID field to identify the author of the message
- Add Message.Data field which contains a copy of the ReceiveMessageEvent.Data value
- Add Message.AuthorID field which contains a copy of the ReceiveMessageEvent.AuthorID value 

## [v0.6.0] - 2019-03-30
### Added
- implement `NewConfig` function to allow create configuration for unit tests of modules

## [v0.5.0] - 2019-03-18
### Fixed
- Fixed nil pointer panic in slack adapter when context is nil

## [v0.4.0] - 2019-03-18
### Changed
- Change type of `Module` from function to interface to allow more flexibility
- Introduce new `ModuleFunc` type to migrate old modules to new interface type

## [v0.3.0] - 2019-03-17
### Added
- Event handler functions can now accept interfaces instead of structs
- Add new `github.com/go-joe/joe/joetest` package for unit tests
- Add new `joetest.Brain` type
- Add new `WithLogger(…)` option

### Changed
- Switch license from MIT to BSD-3-Clause
- Move `TestingT` type into new `joetest` package
- Move `TestBot` type into new `joetest` package and rename to `joetest.Bot`

### Fixed
- Fixed flaky unit test of `CLIAdapter`

## [v0.2.0] - 2019-03-10
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

[Unreleased]: https://github.com/go-joe/joe/compare/v0.6.0...HEAD
[v0.6.0]: https://github.com/go-joe/joe/compare/v0.5.0...v0.6.0
[v0.5.0]: https://github.com/go-joe/joe/compare/v0.4.0...v0.5.0
[v0.4.0]: https://github.com/go-joe/joe/compare/v0.3.0...v0.4.0
[v0.3.0]: https://github.com/go-joe/joe/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/go-joe/joe/compare/v0.1.0...v0.2.0
[v0.1.0]: https://github.com/go-joe/joe/releases/tag/v0.1.0
