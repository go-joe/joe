# Changelog
All notable changes to this project will be documented in this file.

The format is loosely based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

**THIS SOFTWARE IS STILL IN ALPHA AND THERE ARE NO GUARANTEES REGARDING API STABILITY YET.**

Once we reach the v1.0 release, this project will adhere to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
Nothing so far

## [v0.10.0] - 2019-10-26
- Allow event handlers to also use scalar event types (fixes #14)
- Add new `FinishEventContent(…)` function to finish event processing with multiple handlers early
- **Breaking change:** Message handlers registered via `Bot.Respond(…)` and `Bot.RespondRegex(…)` now abort early if the pattern matches
  - This allows users to specify a default response when nothing else matches (see #25)

## [v0.9.0] - 2019-10-22
- Add `Auth.Users()` and `Auth.UserPermissions(…)` functions to allow retrieving all users as well as users permissions.
- Allow adapters to implement the optional `ReactionAwareAdapter` interface if they support emoji reactions
- Add new `reactions` package which contains a compiled list of all officially supported reactions
- Components may now return the new `ErrNotImplemented` if they do not support a feature
- Add new `reactions.Event` that may be emitted by an Adapter so users can listen for it 

## [v0.8.0] - 2019-04-21
- Make `Auth.Grant(…)` idempotent and do not unnecessarily add smaller scopes
- Support extending permissions via `Auth.Grant(…)`
- Add boolean return value to `Auth.Grant(…)` to indicate if a new permission was granted
- Add `Auth.Revoke(…)` to remove permissions
- Fix flaky unit test TestBrain_Memory
- Fix flaky TestCLIAdapter_Register test
- Add new `Storage` type which manages encoding/decoding, concurrent access and logging for a `Memory`
- Factor out `Memory` related logic from Brain into new `Storage` type
    - Removed `Brain.SetMemory(…)`, `Brain.Set(…)`, `Brain.Get(…)`, `Brain.Delete(…)`, `Brain.Memories(…)`, `Brain.Close(…)`
    - All functions above except `Brain.Memories(…)` are now available as functions on the `Bot.Store` field
- The `Auth` type no longer uses the `Memory` interface but instead requires an instance of the new `Storage` type
- Removed the `BrainMemoryEvent` without replacement
- Add `joetest.Storage` type to streamline making assertions on a bots storage/memory
- Change the `Memory` interface to treat values as `[]byte` and not `string`
- Remove `Memories()` function from `Memory` interface and instead add a `Keys()` function  
- `NewConfig(…)` now requires an instance of a `Storage`

## [v0.7.0] - 2019-04-18
- Add ReceiveMessageEvent.Data field to allow using the underlying message type of the adapters
- Add ReceiveMessageEvent.AuthorID field to identify the author of the message
- Add Message.Data field which contains a copy of the ReceiveMessageEvent.Data value
- Add Message.AuthorID field which contains a copy of the ReceiveMessageEvent.AuthorID value 
- Add Auth.Grant(…) and Auth.CheckPermission(…) functions to allow implementing user permissions
- Add Brain.Close() function to let the brain implement the Memory interface
- Add Brain.SetMemory(…) function to give more control over a joe.Brain
- Fix joetest.Bot.Start(…) function to return only when actually _all_ initialization is done

## [v0.6.0] - 2019-03-30
- implement `NewConfig` function to allow create configuration for unit tests of modules

## [v0.5.0] - 2019-03-18
- Fixed nil pointer panic in slack adapter when context is nil

## [v0.4.0] - 2019-03-18
- Change type of `Module` from function to interface to allow more flexibility
- Introduce new `ModuleFunc` type to migrate old modules to new interface type

## [v0.3.0] - 2019-03-17
- Event handler functions can now accept interfaces instead of structs
- Add new `github.com/go-joe/joe/joetest` package for unit tests
- Add new `joetest.Brain` type
- Add new `WithLogger(…)` option
- Switch license from MIT to BSD-3-Clause
- Move `TestingT` type into new `joetest` package
- Move `TestBot` type into new `joetest` package and rename to `joetest.Bot`
- Fixed flaky unit test of `CLIAdapter`

## [v0.2.0] - 2019-03-10
- Add a lot more unit tests
- Add `TestBot.Start()` and `TestBot.Stop()`to ease synchronously starting and stopping bot in unit tests
- Add `TestBot.EmitSync(…)` to emit events synchronously in unit tests 
- Remove obsolete context argument from `NewTest(…)` function
- Errors from passing invalid expressions to `Bot.Respond(…)` are now returned in `Bot.Run()`
- Events are now processed in the exact same order in which they are emitted
- All pending events are now processed before the brain event loop returns
- Replace context argument from `Brain.HandleEvents()` with new `Brain.Shutdown()` function
- `Adapter` interface was simplified again to directly use the `Brain`
- Remove unnecessary `t` argument from `TestBot.EmitSync(…)` function
- Deleted `Brain.Close()` because it was not actually meant to be used to close the brain and is thus confusing

## [v0.1.0] - 2019-03-03

Initial release, note that Joe is still in alpha and the API is not yet considered
stable before the v1.0.0 release.

[Unreleased]: https://github.com/go-joe/joe/compare/v0.10.0...HEAD
[v0.9.0]: https://github.com/go-joe/joe/compare/v0.9.0...v0.10.0
[v0.9.0]: https://github.com/go-joe/joe/compare/v0.8.0...v0.9.0
[v0.8.0]: https://github.com/go-joe/joe/compare/v0.7.0...v0.8.0
[v0.7.0]: https://github.com/go-joe/joe/compare/v0.6.0...v0.7.0
[v0.6.0]: https://github.com/go-joe/joe/compare/v0.5.0...v0.6.0
[v0.5.0]: https://github.com/go-joe/joe/compare/v0.4.0...v0.5.0
[v0.4.0]: https://github.com/go-joe/joe/compare/v0.3.0...v0.4.0
[v0.3.0]: https://github.com/go-joe/joe/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/go-joe/joe/compare/v0.1.0...v0.2.0
[v0.1.0]: https://github.com/go-joe/joe/releases/tag/v0.1.0
