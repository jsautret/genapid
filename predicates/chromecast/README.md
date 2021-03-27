# chromecast

The `chromecast` predicate sends a media file to a Chromecast device.

It can be used to make a Google Home talk using the TTS feature of
[go-chromecast](https://github.com/vishen/go-chromecast#text-to-speech). You
need to setup a Google Service Account (see previous link). Google
Text-to-Speech is free for up to 4 millions of read characters per
month with standard voices.

## Options

| Option                   | Required | Description                                                                    |
| ---                      | ---      | ---                                                                            |
| `tts`                    | yes      | String to say on the Chromecast device                                         |
| `google_service_account` | yes      | Path to your GSA credentials                                                   |
| `language_code`          | yes      | language code                                                                  |
| `voice_name`             | yes      | See https://cloud.google.com/text-to-speech/docs/voices                        |
| `addr`                   | yes      | IP of the Chromecast device                                                    |
| `port`                   |          | Port of the Chromecast device (default to 8009)                                |
| `server_port`            |          | Local port used to server media file. Only needed if you have a local firewall |
| `speaking_rate`          |          | default to 1.0                                                                 |
| `pitch`                  |          | default to 1.0                                                                 |


## Results

Field | Type | Description
---|---|---
`result` | boolean | false if something went wrong

## Example:

See Kodi example in
[../../examples/kodi/kodi.yml](../../examples/kodi/kodi.yml).
