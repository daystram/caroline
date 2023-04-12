# caroline

[![Github Actions](https://github.com/daystram/caroline/actions/workflows/ci.yml/badge.svg)](https://github.com/daystram/caroline/actions/workflows/ci.yml)
[![Docker Pulls](https://img.shields.io/docker/pulls/daystram/caroline)](https://hub.docker.com/r/daystram/caroline)
[![MIT License](https://img.shields.io/github/license/daystram/caroline)](https://github.com/daystram/caroline/blob/master/LICENSE)

_The_ Discord singer.

## Installation

### Go version < 1.16

```shell
$ go get -u github.com/daystram/caroline/cmd/caroline
```

### Go 1.16+

```shell
$ go install github.com/daystram/caroline/cmd/caroline@latest
```

## Requirements

To function properly, the following tools and library must be installed:

- [ffmpeg](https://ffmpeg.org/)
- [Opus](https://opus-codec.org/) development library
- [yt-dlp](https://github.com/yt-dlp/yt-dlp)

## Usage

After providing the required configuration, the bot can simply be run as follows:

```shell
$ caroline
```

### Docker

Instead of installing the command itself, you can run the bot via Docker:

```shell
$ docker run --name caroline --env-file ./.env -d daystram/caroline
```

## Configuration

The bot could be configured by setting the following environment variables.

| Name               | Description            | Default | Required |
| ------------------ | ---------------------- | ------- | -------- |
| `BOT_TOKEN`        | Discord Bot token      | `""`    | ✅       |
| `SP_CLIENT_ID`     | Spotify client ID      | `""`    | ✅       |
| `SP_CLIENT_SECRET` | Spotify client secret  | `""`    | ✅       |
| `DEBUG_GUILD_ID`   | Discord debug Guild ID | `""`    | ⬜       |

## License

This project is licensed under the [MIT license](./LICENSE).
