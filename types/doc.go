// Package types holds the models shared by several ElevenLabs services:
// multipart uploads (File), newline-delimited JSON streams (Stream),
// clearable update fields (Clearable), numeric limits that may be
// "unlimited" (LimitOrUnlimited) and the Ptr helper.
//
// Only types used by three or more service packages live here; everything
// else belongs to the package that owns it.
package types
