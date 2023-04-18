# elevenlabs
Unofficial [elevenlabs.io](https://beta.elevenlabs.io/) ([11.ai](11.ai)) voice synthesis client

This library is not affiliated with, not associated with ElevenLabs in any way.

ElevenLabs' official api documentation, upon which this client has been
derived, [can be found here](https://api.elevenlabs.io/docs).

## Purpose
This go client provides an easy interface to create synthesized voices and
make TTS (text-to-speech) requests to elevenlabs.io


As a prerequisite, you must already have an account with elevenlabs.io.
After creating your account, you can get you API key [from here](https://help.elevenlabs.io/hc/en-us/articles/14599447207697-How-to-authorize-yourself-using-your-xi-api-key-).


To use this library, create a new client and send a TTS request to a voice.
