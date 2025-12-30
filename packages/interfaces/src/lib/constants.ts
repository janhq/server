/**
 * Shared constants for UI components
 */

// Message roles
export const MESSAGE_ROLE = {
  TOOL: "tool",
  USER: "user",
  ASSISTANT: "assistant",
  SYSTEM: "system",
} as const;

export type MessageRoleValue = (typeof MESSAGE_ROLE)[keyof typeof MESSAGE_ROLE];

// Chat Status (matches ChatStatus type from 'ai' SDK)
export const CHAT_STATUS = {
  SUBMITTED: "submitted",
  STREAMING: "streaming",
  READY: "ready",
  ERROR: "error",
} as const;

export type ChatStatusValue = (typeof CHAT_STATUS)[keyof typeof CHAT_STATUS];

// Tool invocation states (matches ToolUIPart states from 'ai' SDK)
export const TOOL_STATE = {
  // Input states
  INPUT_STREAMING: "input-streaming",
  INPUT_AVAILABLE: "input-available",
  // Approval states
  APPROVAL_REQUESTED: "approval-requested",
  APPROVAL_RESPONDED: "approval-responded",
  // Output states
  OUTPUT_AVAILABLE: "output-available",
  OUTPUT_ERROR: "output-error",
  OUTPUT_DENIED: "output-denied",
} as const;

export type ToolStateValue = (typeof TOOL_STATE)[keyof typeof TOOL_STATE];

// Upload status
export const UPLOAD_STATUS = {
  PENDING: "pending",
  UPLOADING: "uploading",
  COMPLETED: "completed",
  FAILED: "failed",
} as const;

export type UploadStatusValue =
  (typeof UPLOAD_STATUS)[keyof typeof UPLOAD_STATUS];

// Browser/WebSocket connection states
export const CONNECTION_STATE = {
  DISCONNECTED: "disconnected",
  CONNECTING: "connecting",
  CONNECTED: "connected",
  ERROR: "error",
} as const;

export type ConnectionStateValue =
  (typeof CONNECTION_STATE)[keyof typeof CONNECTION_STATE];

// Session storage keys and prefixes
export const SESSION_STORAGE_KEY = {
  INITIAL_MESSAGE_TEMPORARY: "initialMessage_temporary",
} as const;

export const SESSION_STORAGE_PREFIX = {
  INITIAL_MESSAGE: "initialMessage_",
} as const;
