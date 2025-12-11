/**
 * Media proxy utilities for handling anti-hotlinking and caching
 */

/**
 * Convert a media URL to use the proxy endpoint
 * @param url Original media URL
 * @param referer Optional referer URL for anti-hotlinking
 * @returns Proxied URL
 */
export function getProxiedMediaUrl(url: string, referer?: string): string {
  if (!url) return '';

  // Don't proxy data URLs or blob URLs
  if (url.startsWith('data:') || url.startsWith('blob:')) {
    return url;
  }

  // Don't proxy local URLs
  if (
    url.startsWith('/') ||
    url.startsWith('http://localhost') ||
    url.startsWith('http://127.0.0.1')
  ) {
    return url;
  }

  // Build proxy URL
  const params = new URLSearchParams();
  params.set('url', url);
  if (referer) {
    params.set('referer', referer);
  }

  return `/api/media/proxy?${params.toString()}`;
}

/**
 * Check if media caching is enabled
 * @returns Promise<boolean>
 */
export async function isMediaCacheEnabled(): Promise<boolean> {
  try {
    const response = await fetch('/api/settings');
    if (response.ok) {
      const settings = await response.json();
      return settings.media_cache_enabled === 'true' || settings.media_cache_enabled === true;
    }
  } catch (error) {
    console.error('Failed to check media cache status:', error);
  }
  return false;
}

/**
 * Process HTML content to proxy image URLs
 * @param html HTML content
 * @param referer Optional referer URL
 * @returns HTML with proxied image URLs
 */
export function proxyImagesInHtml(html: string, referer?: string): string {
  if (!html) return html;

  // Replace img src attributes
  return html.replace(/<img([^>]+)src="([^"]+)"/gi, (match, attrs, src) => {
    const proxiedUrl = getProxiedMediaUrl(src, referer);
    return `<img${attrs}src="${proxiedUrl}"`;
  });
}
