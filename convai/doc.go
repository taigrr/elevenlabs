// Package convai is the entry point to the ElevenLabs Agents platform
// (conversational AI). Service wires the agent, conversation, knowledge base,
// tool, test, telephony, MCP and triage sub-services and holds the workspace
// level endpoints: settings, secrets, dashboard, analytics, LLM listing and
// usage, conversation users and realtime client secrets.
//
// Service.Connect opens a realtime conversation with an agent over a
// WebSocket and returns a Session that streams audio and events in both
// directions, optionally dispatching client tool calls to ClientTools.
package convai
