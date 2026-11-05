// Package realtime defines the errors shared by the ElevenLabs WebSocket
// clients (text-to-speech input streaming, Scribe realtime, dubbing and
// translation, agent conversations).
//
// The realtime entry points themselves are methods of the owning services,
// e.g. texttospeech.Service.StreamInput. Server error messages are returned
// as *Error, abnormal closures as *CloseError; both can be matched with
// errors.Is against the sentinels of this package and of package apierror.
package realtime
