# Borik

A discord bot, written using [discordgo](https://github.com/bwmarrin/discordgo), for ✨ breaking images ✨.

## Running the bot

- Build all dependencies with custom permitted cflags, to build the ImageMagick wrapper:
  ```shell
  CGO_CFLAGS_ALLOW=-Xpreprocessor go build -a
  ```
- Copy `.env.dist` to `.env`, and populate it with a token and a prefix
- `go run .`

### MacOS

ImageMagick 6 is required, however the version that installs itself if you run `brew install imagemagick@6` doesn't include the `lqr` library, so using [this](https://github.com/nint8835/homebrew-formulae/blob/main/Formula/imagemagick%406.rb) custom homebrew formula, run:

```sh
brew install nint8835/formulae/imagemagick@6
brew link nint8835/formulae/imagemagick@6
```
