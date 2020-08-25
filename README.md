# ErikBotDev

TODO: Add description and usage here later.

## Configuration

Many of the commands in this bot interact with OBS through the [OBS websocket plugin](https://obsproject.com/forum/resources/obs-websocket-remote-control-obs-studio-from-websockets.466/). By default, this bot _requires_ that it can connect to the websocket in order to even run. If it can't it marks itself as offline and won't respond to any commands.

### If you want to force this into "streaming mode"

If you do not have the OBS websocket plugin running, you have two options for forcing the bot to configure itself into streaming mode:

- Pass `-s` or `--streaming-on`
- Mark an individual command `"offline": true` to enable just that command

## Sounds

Any sound you reference in the config file ([sample](./erikbotdev.json)) needs to be a WAV file in the media directory.