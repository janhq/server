import { cn } from "../lib/utils";
import type { ToolUIPart } from "ai";
import {
  CheckCircleIcon,
  CircleIcon,
  ClockIcon,
  SearchIcon,
  WrenchIcon,
  XCircleIcon,
} from "lucide-react";
import type { ComponentProps, ReactNode } from "react";
import { useMemo } from "react";
import { TOOL_STATE } from "../lib/constants";

export type ToolCompactProps = ComponentProps<"button"> & {
  toolName: string;
  state: ToolUIPart["state"];
  isSelected?: boolean;
  input?: Record<string, unknown>;
};

const getCompactStatusIcon = (status: ToolUIPart["state"]): ReactNode => {
  const icons: Record<ToolUIPart["state"], ReactNode> = {
    [TOOL_STATE.INPUT_STREAMING]: (
      <CircleIcon className="size-3 text-muted-foreground" />
    ),
    [TOOL_STATE.INPUT_AVAILABLE]: (
      <ClockIcon className="size-3 animate-pulse text-blue-500" />
    ),
    // @ts-expect-error state only available in AI SDK v6
    [TOOL_STATE.APPROVAL_REQUESTED]: (
      <ClockIcon className="size-3 text-yellow-600" />
    ),
    [TOOL_STATE.APPROVAL_RESPONDED]: (
      <CheckCircleIcon className="size-3 text-blue-600" />
    ),
    [TOOL_STATE.OUTPUT_AVAILABLE]: (
      <CheckCircleIcon className="size-3 text-green-600" />
    ),
    [TOOL_STATE.OUTPUT_ERROR]: <XCircleIcon className="size-3 text-red-600" />,
    [TOOL_STATE.OUTPUT_DENIED]: (
      <XCircleIcon className="size-3 text-orange-600" />
    ),
  };

  return icons[status];
};

// Generate a pseudo-random max-width based on tool name for visual variety
const getRandomMaxWidth = (seed: string): string => {
  let hash = 0;
  for (let i = 0; i < seed.length; i++) {
    hash = (hash << 5) - hash + seed.charCodeAt(i);
    hash |= 0;
  }
  // Generate width between 120px and 180px
  const width = 120 + Math.abs(hash % 60);
  return `${width}px`;
};

// Check if this is a search tool and extract query
const getSearchDisplay = (
  toolName: string,
  input?: Record<string, unknown>,
): { isSearch: boolean; query?: string } => {
  const searchTools = ["google_search", "search", "web_search", "bing_search"];
  const isSearch = searchTools.includes(toolName.toLowerCase());

  if (!isSearch || !input) {
    return { isSearch: false };
  }

  // Try common query parameter names
  const query =
    (input.q as string) ||
    (input.query as string) ||
    (input.search as string) ||
    (input.term as string);

  return { isSearch: true, query };
};

export const ToolCompact = ({
  className,
  toolName,
  state,
  isSelected,
  input,
  ...props
}: ToolCompactProps) => {
  const { isSearch, query } = useMemo(
    () => getSearchDisplay(toolName, input),
    [toolName, input],
  );

  // Use query content for seed if search, otherwise toolName + index
  const maxWidth = useMemo(
    () => getRandomMaxWidth(query || toolName + Math.random().toString()),
    [query, toolName],
  );

  const displayText = isSearch && query ? `Search: ${query}` : toolName;
  const IconComponent = isSearch ? SearchIcon : WrenchIcon;

  return (
    <button
      type="button"
      style={{ maxWidth }}
      className={cn(
        "inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-md border bg-background hover:bg-muted/50 transition-colors cursor-pointer text-left",
        "min-w-[80px]",
        isSelected && "ring-2 ring-primary",
        className,
      )}
      {...props}
    >
      <IconComponent className="size-3 shrink-0 text-muted-foreground" />
      <span className="text-xs font-medium truncate flex-1">{displayText}</span>
      {getCompactStatusIcon(state)}
    </button>
  );
};

export type ToolCompactListProps = ComponentProps<"div"> & {
  children: ReactNode;
};

export const ToolCompactList = ({
  className,
  children,
  ...props
}: ToolCompactListProps) => (
  <div className={cn("flex flex-wrap gap-2 mb-4", className)} {...props}>
    {children}
  </div>
);
