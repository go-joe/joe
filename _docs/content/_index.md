# - Joe Bot :robot: -

Joe is a library used to write chat bots in [the Go programming language][go].

<a id="fork-me-on-github" href="https://github.com/go-joe/joe">Fork me on GitHub</a>

## Features

- **Chat adapters** for <i class='fab fa-slack fa-fw'></i> Slack, <i class='fab fa-rocketchat fa-fw'></i> Rocket.Chat, <i class='fab fa-telegram fa-fw'></i> Telegram and <i class='fas fa-hashtag fa-fw'></i> IRC. Adding your own is easy as well.  
- **Event processing system** to consume HTTP callbacks (e.g. from GitHub) or to trigger events on a schedule using Cron expressions
- **Persistence** of key-value data (e.g. using Redis or SQL)
- **User permissions** to restrict some actions to privileged users
- **Unit tests** are first class citizens, Joe has high code coverage and ships with a dedicated package to facilitate your own tests

## Design

- **Minimal**: Joe ships with no third-party dependencies except for logging and error handling.
- **Modular**: choose your own chat adapter (e.g. Slack), memory implementation (e.g. Redis) and more.
- **Batteries included**: you can start developing your bot on the CLI without extra configuration.
- **Simple**: your own message & event handlers are simple and easy to understand functions without too much cruft or boilerplate setup.  

## Getting Started

To get started writing your own bot with Joe, head over to the
[**Quickstart**](/quick) section or directly have a look at the
[**Basic Tutorials**](/basic) to learn the core concepts.
If you want to dive right in and want to know what modules are currently provided
by the community, then have a look at the [**Available Modules**](/modules) section.
Last but not least, you can find more instructions and best practices in the [**Recipes**](/recipes) section. 

## Contact & Contributing

To contribute to Joe, you can either write your own Module (e.g. to integrate
another chat adapter) or work on Joe's code directly. You can of course also
extend or improve this documentation or help with reviewing issues and pull
requests at https://github.com/go-joe/joe. Further details about how to
contribute can be found in the [CONTRIBUTING.md][contributing] file.

## License

The Joe library is licensed under the [BSD-3-Clause License][license].

[go]: https://golang.org
[hubot]: https://hubot.github.com/
[license]: https://github.com/go-joe/joe/blob/master/LICENSE
[contributing]: https://github.com/go-joe/joe/blob/master/CONTRIBUTING.md
