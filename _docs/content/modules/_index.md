+++
title = "Available Modules"
slug = "modules"
weight = 3
pre = "<b>3. </b>"
+++

Joe ships with no third-party modules such as Redis integration to avoid pulling
in more dependencies than you actually require. There are however already some
modules that you can use directly to extend the functionality of your bot without
writing too much code yourself.

### Chat Adapters
<p style="font-size: 80%;margin-top: 0px;">
Adapters let you interact with the outside world by receiving and sending messages.
</p>

- <i class='fab fa-slack fa-fw'></i> Slack Adapter: https://github.com/go-joe/slack-adapter
- <i class='fab fa-rocketchat fa-fw'></i> Rocket.Chat Adapter: https://github.com/dwmunster/rocket-adapter
- <i class='fab fa-telegram fa-fw'></i> Telegram Adapter: https://github.com/robertgzr/joe-telegram-adapter
- <i class='fas fa-hashtag fa-fw'></i> IRC Adapter: https://github.com/akrennmair/joe-irc-adapter
- <i class="fas fa-circle-notch"></i> Mattermost Adapter: https://github.com/dwmunster/joe-mattermost-adapter
- <i class='fab fa-vk'></i> VK Adapter: https://github.com/tdakkota/joe-vk-adapter

### Memory Modules
<p style="font-size: 80%;margin-top: 0px;">
Memory modules let you persist key value data so it can be accessed again later.
</p>

- Redis Memory: https://github.com/go-joe/redis-memory
- File Memory: https://github.com/go-joe/file-memory
- Bolt Memory: https://github.com/robertgzr/joe-bolt-memory
- SQLite Memory: https://github.com/warmans/sqlite-memory

### Other Modules
<p style="font-size: 80%;margin-top: 0px;">
General purpose Modules may register handlers or emit events.
</p>

- HTTP Server: https://github.com/go-joe/http-server
- Cron Jobs: https://github.com/go-joe/cron
