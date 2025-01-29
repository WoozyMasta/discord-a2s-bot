---
render_with_liquid: false
---
<!-- omit in toc -->
# Discord A2S Game Server Monitor Bot

![logo]

A Discord bot that monitors game servers using Steam [A2S] server queries
`A2S_INFO` and updates Discord channels and Rich Presence based
on server statuses.

* **Concurrent Monitoring**:
  Monitors multiple game servers simultaneously.
* **Dynamic Channel Updates**:
  Automatically updates Discord channel names and descriptions with real-time
  server information;
* **Rich Presence Integration**:
  Maintains a Rich Presence status reflecting the overall status of
  all monitored servers;
* **Customizable Templates**:
  Use templates to define how server information is displayed in
  channels and Rich Presence;
* **Recycling-friendly API**:
  Updates data in Discord only if something actually changed;
* **Support any Steam Query games**:
  Will work with any game that uses the Steam Query protocol and
  responds to an A2S_INFO request;
* **Extended Arma3 and DayZ support**:
  Supports parsing of special keywords that DayZ and Arma 3 games
  respond to, such as players queue length, server time, mission name or
  current mission status, etc.

![example]

<!-- omit in toc -->
## Table of content

* [Installation](#installation)
* [Usage](#usage)
* [Basic Configuration](#basic-configuration)
* [Templating](#templating)
  * [Explain template](#explain-template)
  * [Templating data](#templating-data)
  * [Templating functions](#templating-functions)
  * [Example template for learning](#example-template-for-learning)
* [Setup](#setup)
  * [Obtaining the Discord Bot Token and Setting Permissions](#obtaining-the-discord-bot-token-and-setting-permissions)
  * [Configure the application launch](#configure-the-application-launch)
    * [Container Image](#container-image)
    * [Systemd service](#systemd-service)
    * [Windows service](#windows-service)
* [Other](#other)
  * [YAML anchors merging](#yaml-anchors-merging)
* [Support me ☕](#support-me-)

## Installation

You can download the latest version of the program by following the links:

| i386             | amd64             | arm           | arm64             |
| ---------------- | ----------------- | ------------- | ----------------- |
|                  | [MacOS amd64][]   |               | [MacOS arm64][]   |
| [Linux i386][]   | [Linux amd64][]   | [Linux arm][] | [Linux arm64][]   |
| [Windows i386][] | [Windows amd64][] |               | [Windows arm64][] |

For Linux you can also use the command

```bash
curl -#SfLo /usr/bin/discord-a2s-bot \
  https://github.com/WoozyMasta/discord-a2s-bot/releases/latest/download/discord-a2s-bot-linux-amd64
chmod +x /usr/bin/discord-a2s-bot
discord-a2s-bot -h && discord-a2s-bot -v
```

## Usage

For show this help pass `-h` or `--help` flag

```txt
Usage:
  discord-a2s-bot [option] [config.(yaml|json)]

Available options:
  -e, --example    Prints an example YAML configuration file.
  -v, --version    Show version, commit, and build time.
  -h, --help       Prints this help message.
```

By default it try open configuration file with name `config.yaml` in current
directory, or you can pass path to config as argument to the program.

## Basic Configuration

Create a `config.yaml` file in the project root directory with the
following structure:

```yaml
# Bot configuration settings
bot:
  token: # Discord bot token (required)
  update_interval: 30s # Interval for query servers for presence status and channels updates (default 30s)
  concurrency: 10 # Number of concurrent servers updates (default 10)

# Defines settings for servers, 
servers:
  - id: my supper server
    host: 127.0.0.1 # Server host address (default 127.0.0.1)
    port: 27016 # Server query port (default 27016)
    timeout: 3 # Timeout for server queries in seconds (default 3)
    buffer_size: 1024 # Buffer size for server responses (default 1024)

    # Discord channel ID to update, not set to disable
    channel_id: CHANNEL_ID_FOR_SERVER1 
    # Template for update name of channel, not change if blank or not set
    channel_name: "{{ .Info.Players }}∶{{ .Info.MaxPlayers }} {{ .ID }}"
    # Template for update description of channel, not change if blank or not set
    channel_description: "{{ .Info.Name }} {{ .Info.Map }}"

    # Discord category ID to update, not set to disable
    category_id: CATEGORY_ID_FOR_SERVER1
    # Template for update name of category, not change if blank or not set
    category_name: "Super server {{ .Info.Name }}"

# Logging configuration settings
logging:
  level: info # Log level (debug, info, warn, error, etc.)
  format: text # Log format (text or json)
  output: stdout # Log output destination (stdout, stderr, or file path)
```

For a more detailed example configuration, you can look at the
[example.config.yaml] file or run the command to create an example
configuration without referring to this document:

```bash
./discord-a2s-bot --example > config.yaml
```

Also, if you have not worked with YAML anchors before, you may find
the [yq] utility useful, which will help validate the config and
merge all anchors for clarity.
Or, for example, save the config in json format, if that is more
convenient for you, json will also be accepted by the program.

```bash
# merge anchors and remove technical block
./discord-a2s-bot -e | yq -er 'explode(.) | del(.base-template)' > config.yaml
# save as json
./discord-a2s-bot -e | yq -er 'del(.base-template)' -o json > config.json
```

## Templating

In the detailed example you can see something like this template for
the channel name:

```go
{{ if .Info -}}
{{ if eq .Info.ID 221100 }}{{ if .Extra.Time }}{{ DurationEmoji .Extra.Time }}{{ end }}{{ else }}🟢{{ end -}}
-{{ .Info.Players }}∶{{ .Info.MaxPlayers }}-{{- .ID -}}
{{ else -}}
🔴-{{- .ID -}}
{{ end -}}
```

This is a standard go mechanism that implements data-driven templates for
generating text output.
You can find more detailed information about the syntax in the official
documentation [text/template].

### Explain template

Let's try to take it apart:

* The above template, in case the block with information `.Info` was
  received from the server
  * If the `.Info` response has an application ID of `221100` (DayZ)
    * If extended information `.Exta.Time` was received
      (contains information about the current time on the server in DayZ)
      * We call a special function `DurationEmoji` that will output the
        duration from the server as an emoji of day 🌞 or night 🌙
    * Otherwise, if it is not DayZ, we will output a green status 🟢
  * We will output the number of players on the server and the number of
    available slots, as well as our short `.ID` of the server as the name.
* Otherwise, if no `.Info` response was received (the server is offline)
  * We will output a red status 🔴 as well as our short `.ID` of the
    server as the name.

And we will get some such result the channel name will be set in Discord:

```bash
# This is a DayZ server and it is daytime, 5 out of 30 players are playing there
🌞-5:30-myserver
# This server is turned off, and we have no information about it except what is in the config
🔴-myserver
# This is some kind of game server with an online of 270 out of 600
🟢-270:600-myserver
# This is a DayZ server and it is nighttime, there are no players on 64 available slots
🌙-0:64-myserver
```

If everything is clear with the logic of the templates,
now we need to understand what variables are available to us.

### Templating data

We have access to the following data:

* `.Info.*` - Server information structure from A2S. A full description of
  the entire structure and fields can be found in the parser file [.Info]
* `.Extra.*` - Additional arbitrary data structure. The data in this structure
  is not static and varies from game to game, and is absent for most.
  Here you can find detailed descriptions for:
  * [Arma 3 keywords][]
  * [DayZ keywords][]
* `.ID` - Server identifier (from configuration file)
* `.Host` - Server host address (from configuration file)
* `.Port` - Server port (from configuration file)

> [!TIP]  
> In `.Extra` currently, here is contains additionally processed data
> for DayZ and Arma 3 games

More details can be found in the [template.go](cli/template.go) source code.

In addition to the data set, there are also various built-in functions for
processing data. For example, how we previously displayed the
time of day using emoji.

### Templating functions

<!-- omit in toc -->
#### `AppID`

If you pass it an application ID in steam, it will convert it to a full name.

> [!TIP]  
> This is not a complete database, it supports several hundred popular games

```go
{{ AppID .Info.ID }}
// for example for ID 252490 it will return the string Rust
```

<!-- omit in toc -->
#### `DurationEmoji` and `TimeEmoji`

Converts duration and time respectively into hours and outputs

* 🌙 for night (0-7 hours or after 20 hours)
* 🌞 for day (7-20 hours)

```go
{{ DurationEmoji .Extra.Time }}
// for example for 13:15 it will return the 🌞
```

<!-- omit in toc -->
#### `OSEmoji`

For lines corresponding to different OS it will output

* 🍎 - MacOS
* 🐧 - Linux
* 🪟 - Windows
* 😈 - FreeBSD (default for others)

```go
{{ OSEmoji .Info.Environment }}
// for example for win it will return the 🪟
```

<!-- omit in toc -->
#### `CountryEmoji` and `CodeEmoji`

`CountryEmoji` will try to find a short code from the country name, and
`CodeEmoji` immediately accepts a short code as input and in both cases they
try to return an emoji with the country flag, otherwise 🏳️

```go
{{ CountryEmoji .Extra.Language }}
// for example for "south korea" it will return the 🇰🇷
{{ CodeEmoji .Extra.Language }}
// for example for "AU" it will return the 🇦🇺
```

<!-- omit in toc -->
#### `ValueColorEmoji`

Returns color emoji based on the current meaning and maximum.
Need pass two value, current value and limit, 0 its start point.

* 🟣 — 0
* 🔵 — <10%
* 🟢 — <50%
* 🟡 — <75%
* 🟠 — <90%
* 🔴 — 100%
* 🚫 — the value is less than 0 or exceeds the maximum

```go
{{ ValueColorEmoji .Info.Players .Info.MaxPlayers }}
// for example for 43 players on 100 slots it will return the 🟢
{{ ValueColorEmoji .Extra.PlayersQueue 20 }}
// for example for 19 players in queue and 20 limit it will return the 🔴
{{ ValueColorEmoji -10 20 }}
{{ ValueColorEmoji 10 5 }}
// for example for its broken values it will return the 🚫
```

### Example template for learning

Now that you have read this, it will not be difficult for you to read and
adjust this template from the example for a detailed description of the
channel to suit your taste.

```go
{{ if .Info -}}
🟢 {{ .Info.Name }}
📜 {{ .Info.Game }}
{{ if eq .Info.ID 107410 }}{{ if .Extra.GameType -}}
🎯 {{ .Extra.GameType }}
{{ end }}{{ end -}}
👥 {{ .Info.Players }}/{{ .Info.MaxPlayers }}{{ if eq .Info.ID 221100 }}{{ if .Extra.PlayersQueue }} ({{ .Extra.PlayersQueue }}){{ end }}{{ end }}
🌍 {{ .Info.Map }}
📡 {{ .Host }}:{{ .Info.Port }}
⚙️ {{ .Info.Environment }} {{ AppID .Info.ID }} {{ .Info.Version }} server
{{ else -}}
🔴 Server {{ .ID }} offline
📡 {{ .Host }}:{{ .Port }}"
{{ end -}}
```

## Setup

### Obtaining the Discord Bot Token and Setting Permissions

To set up your Discord bot, follow these steps:

* Create a Discord Application
  * Navigate to the Discord Developer Portal.
    <https://discord.com/developers/applications>
  * Click on "New Application".
  * Enter a name for your application and click "Create".
* Add a Bot to Your Application
  * In your application's dashboard, go to the "Bot" tab.
  * Click "Add Bot" and confirm by clicking "Yes, do it!".
* Configure the Bot:
  * Username: Set the bot's name.
  * Avatar: Upload an avatar image for the bot.
  * Banner: (Optional) Set a banner image.
  * Public Bot:
    Disable this option to prevent others from adding your bot to their servers.
  * Token: Click "Reset Token" to generate a new token.
    Keep this token secure as it grants control over your bot.

In the "OAuth2" tab you can select "URL Generator".
In "Scope" check the `bot` option and under "Bot Permissions",
select: `Manage Channels` and `View Channels`

Generated OAuth2 URL will appear at the bottom. It will look similar to:

```txt
https://discord.com/oauth2/authorize?client_id=YOUR_CLIENT_ID&permissions=1040&scope=bot
```

Or just use this link, replace `YOUR_CLIENT_ID` with your bot id and follow it,
specify the server where you want to add your bot.

### Configure the application launch

#### Container Image

The images are published to two container registries:

* [`docker pull ghcr.io/woozymasta/discord-a2s-bot:latest`][ghcr]
* [`docker pull docker.io/woozymasta/discord-a2s-bot:latest`][docker]

Quick start:

```bash
# Pull the image
docker pull ghcr.io/woozymasta/discord-a2s-bot:latest
# Generate an example YAML config
docker run --rm -ti ghcr.io/woozymasta/discord-a2s-bot:latest --example > discord-a2s-bot.yaml
# Edit the config file
editor discord-a2s-bot.yaml
# Run the container with the mounted config
docker run --name discord-a2s-bot -d \
  -v "$PWD/discord-a2s-bot.yaml:/config.yaml" \
  ghcr.io/woozymasta/discord-a2s-bot:latest
```

You can also use environment variables instead of a configuration file
by running `--get-env` to get an example and passing them
as container environment variables.

> [!TIP]  
> When running in Kubernetes or other container orchestrators, use
> `/health/liveness` and `/health/readiness` endpoints to check the
> health and readiness of the containerized application.

#### Systemd service

To run the Discord A2S Bot as a systemd service, use the following example
configuration. This ensures the exporter runs on system startup.

```ini
[Unit]
Description=Discord A2S Bot
Documentation=https://woozymasta.github.io/discord-a2s-bot/
Wants=network-online.target
After=network-online.target dayz-server.target

[Service]
ExecStart=/usr/bin/discord-a2s-bot /etc/discord-a2s-bot.yaml
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

> [!WARNING]  
> Do not use the `root` user for production environments.
> It's recommended to create a dedicated user for this purpose.

Save this as `/etc/systemd/system/discord-a2s-bot.service`
and enable it using

```bash
systemctl enable discord-a2s-bot
systemctl start discord-a2s-bot
```

#### Windows service

You can run the exporter using any method that suits you, but it's
recommended to use a Windows service for better management and reliability.

To register the service, assuming the application and configuration are
already downloaded and set up in the `C:\discord-a2s-bot` directory,
use the following commands:

```powershell
sc.exe create discord-a2s-bot `
  binPath= "C:\discord-a2s-bot\discord-a2s-bot.exe C:\discord-a2s-bot\config.yaml" `
  DisplayName= "Discord A2S Bot" `
  start= auto

sc.exe start discord-a2s-bot
sc.exe query discord-a2s-bot
```

> [!TIP]  
> You can specify a more descriptive `DisplayName`, especially if you have
> multiple servers or exporters running, to make management easier.

For uninstall use:

```powershell
sc.exe stop discord-a2s-bot
sc.exe query discord-a2s-bot
sc.exe delete discord-a2s-bot
```

## Other

The Discord API is not always very responsive, I tried to load it as little
as possible and I check every time that I don't send it any changes in vain.

But the API is still very loaded periodically, and if it was not possible to
wait for a response between `update_interval`, this data change will not be
taken into account.

All API calls are non-blocking, and if a call hangs somewhere, it will not
disrupt the update of other servers.

Don't use very frequent `update_interval`, it is unlikely that anyone really
needs such frequent updates, and have pity on Discord, there are many of
you. If you want to collect detailed statistics about your server, take a
look at this [discord-a2s-bot] project for DayZ, I think you can find something
similar for your game.

### YAML anchors merging

In order to avoid complex configuration and provide flexibility of
configuration, where each server can have its own personal template, the
configuration example uses YAML anchors and their merging, which allows
repeating configuration blocks to be reused between sections.

The example has a configuration key `base-template` which is not actually
used by the application itself, it is just an intermediate storage for a
common configuration for all servers.

You can read more about YAML in
<https://yaml.org/spec/1.2.2/#322-serialization-tree> and
<https://yaml.org/spec/1.2.2/#chapter-7-flow-style-productions>, and in
simpler language in the article
<https://medium.com/@kinghuang/docker-compose-anchors-aliases-extensions-a1e4105d70bd>

## Support me ☕

If you enjoy my projects and want to support further development,
feel free to donate! Every contribution helps to keep the work going.
Thank you!

<!-- omit in toc -->
### Crypto Donations

<!-- cSpell:disable -->
* **BTC**: `1Jb6vZAMVLQ9wwkyZfx2XgL5cjPfJ8UU3c`
* **USDT (TRC20)**: `TN99xawQTZKraRyvPAwMT4UfoS57hdH8Kz`
* **TON**: `UQBB5D7cL5EW3rHM_44rur9RDMz_fvg222R4dFiCAzBO_ptH`
<!-- cSpell:enable -->

Your support is greatly appreciated!

<!-- Links -->
[logo]: assets/discord-a2s.png
[example]: assets/example.jpg
[example.config.yaml]: cli/example.config.yaml
[text/template]: https://pkg.go.dev/text/template

[.Info]: https://github.com/WoozyMasta/a2s/blob/master/pkg/a2s/a2s_info.go#L11
[Arma 3 keywords]: https://github.com/WoozyMasta/a2s/blob/master/pkg/keywords/arma3.go#L11
[DayZ keywords]: https://github.com/WoozyMasta/a2s/blob/master/pkg/keywords/dayz.go#L9

[discord-a2s-bot]: https://github.com/WoozyMasta/discord-a2s-bot

[MacOS arm64]: https://github.com/WoozyMasta/discord-a2s-bot/releases/latest/download/discord-a2s-bot-darwin-arm64 "MacOS arm64 file"
[MacOS amd64]: https://github.com/WoozyMasta/discord-a2s-bot/releases/latest/download/discord-a2s-bot-darwin-amd64 "MacOS amd64 file"
[Linux i386]: https://github.com/WoozyMasta/discord-a2s-bot/releases/latest/download/discord-a2s-bot-linux-386 "Linux i386 file"
[Linux amd64]: https://github.com/WoozyMasta/discord-a2s-bot/releases/latest/download/discord-a2s-bot-linux-amd64 "Linux amd64 file"
[Linux arm]: https://github.com/WoozyMasta/discord-a2s-bot/releases/latest/download/discord-a2s-bot-linux-arm "Linux arm file"
[Linux arm64]: https://github.com/WoozyMasta/discord-a2s-bot/releases/latest/download/discord-a2s-bot-linux-arm64 "Linux arm64 file"
[Windows i386]: https://github.com/WoozyMasta/discord-a2s-bot/releases/latest/download/discord-a2s-bot-windows-386.exe "Windows i386 file"
[Windows amd64]: https://github.com/WoozyMasta/discord-a2s-bot/releases/latest/download/discord-a2s-bot-windows-amd64.exe "Windows amd64 file"
[Windows arm64]: https://github.com/WoozyMasta/discord-a2s-bot/releases/latest/download/discord-a2s-bot-windows-arm64.exe "Windows arm64 file"

[A2S]: https://developer.valvesoftware.com/wiki/Server_queries
[yq]: https://github.com/mikefarah/yq/releases/latest
