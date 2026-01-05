/* eslint-disable react-hooks/set-state-in-effect */
import { useState, useEffect } from "react";

/**
 * Hook to resolve Jan media URLs to displayable presigned URLs
 * Handles loading state and error handling
 *
 * @param url - The URL to resolve (can be jan media URL or regular URL)
 * @param isJanMediaUrl - Function to check if URL is a Jan media URL
 * @param resolveJanMediaUrl - Function to resolve Jan media URL to presigned URL
 * @returns Object with displayUrl, isLoading state
 */
export function useResolvedMediaUrl(
  url: string | undefined,
  resolveJanMediaUrl: (url: string) => Promise<string>,
) {
  const [displayUrl, setDisplayUrl] = useState<string | undefined>(undefined);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    if (!url) {
      setDisplayUrl(undefined);
      return;
    }

    // If resolver functions are not provided, just use the URL directly
    if (!isJanMediaUrl || !resolveJanMediaUrl) {
      setDisplayUrl(url);
      return;
    }

    // If it's a jan media URL, resolve it to presigned URL
    if (isJanMediaUrl(url)) {
      setIsLoading(true);
      resolveJanMediaUrl(url)
        .then(setDisplayUrl)
        .catch((err) => {
          console.error("Failed to resolve jan media URL:", err);
          setDisplayUrl(undefined);
        })
        .finally(() => setIsLoading(false));
    } else {
      // Regular URL - use directly
      setDisplayUrl(url);
    }
  }, [url, isJanMediaUrl, resolveJanMediaUrl]);

  return { displayUrl, isLoading };
}

/**
 * Check if a URL is a jan media URL format
 * Format: data:image/jpeg;base64,jan_MEDIA_ID
 * @param url - The URL to check
 * @returns true if the URL is a jan media URL
 */
export function isJanMediaUrl(url: string): boolean {
  return url.startsWith("data:") && url.includes(",jan_");
}