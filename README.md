# Borik

a discord bot, written using [discordgo](https://github.com/bwmarrin/discordgo), for ✨ breaking images ✨.

## Running the bot

_`TODO`_

### MacOS

Imagemagick 6 is required, however the version that installs itself if you run `brew install imagemagick@6` doesn't include the `lqr` library, so using [this](https://github.com/nint8835/homebrew-formulae/blob/main/Formula/imagemagick%406.rb) custom homebrew formula, run:

```sh
brew install nint8835/formulae/imagemagick@6
brew link nint8835/formulae/imagemagick@6 # might not be required, ensures that this formulae is the system imagemagick
```
