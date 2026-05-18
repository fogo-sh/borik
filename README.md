# Borik

A discord bot, written using [discordgo](https://github.com/bwmarrin/discordgo), for ✨ breaking images ✨.

## Running the bot

- Build all dependencies with custom permitted cflags, to build the ImageMagick wrapper:
  ```shell
  CGO_CFLAGS_ALLOW=-Xpreprocessor go build -a ./cmd/borik ./cmd/borik-worker
  ```
- Copy `.env.dist` to `.env`, and populate it with a token and a prefix
- Run the bot with `go run ./cmd/borik`
- Run the worker with `go run ./cmd/borik-worker`
- For local development, run the bot, worker, and an embedded Temporal dev server with `go run ./cmd/borik-dev`
  - To use an already-running Temporal server instead, set `BORIK_TEMPORAL_HOST_PORT` and run `go run ./cmd/borik-dev --temporal-dev-server=false`

### Nix

If you have Nix installed and Nix Flakes enabled, this repo provides a Flake to streamline the process of running & developing the bot.

Start by following the `.env` instructions above, then do one of the following:

#### Running for usage

If you just want to use Borik and don't intend on working on it yourself, running `nix run .#bot` and `nix run .#worker` should be all that is required to compile and start each component.

#### For development

If you plan on working on Borik, run the below commands to prepare a dev shell and run Borik in it.

- Run `nix develop`
  - This will drop you into a shell with Go & all required dependencies ready to go.
- Run `go run ./cmd/borik-dev` to start the bot, worker, and embedded Temporal dev server together
  - You can still run `go run ./cmd/borik` or `go run ./cmd/borik-worker` individually when working on one component
