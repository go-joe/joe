# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- Add `TestBot.Start()` and `TestBot.Stop()`to ease synchronously starting and stopping bot in unit tests
- Add `TestBot.EmitSync(…)` to emit events synchronously in unit tests 

### Changed
- Remove obsolete context argument from `NewTest(…)` function
- Errors from passing invalid expressions to `Bot.Respond(…)` are now returned in `Bot.Run()`

## [v0.1.0] - 2019-03-03

Initial release, note that Joe is still in alpha and the API is not yet considered
stable before the v1.0.0 release.

[Unreleased]: https://github.com/go-joe/joe/compare/v0.1.0...HEAD
[v0.1.0]: https://github.com/go-joe/joe/releases/tag/v0.1.0