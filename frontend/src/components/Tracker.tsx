'use client';

import { usePathname } from 'next/navigation';
import { useEffect, useRef } from 'react';
import { trackPageView } from '@/lib/track';

/** 路由级埋点:监听 pathname 变化,自动上报页面浏览(首屏也会上报)。 */
export function Tracker() {
  const pathname = usePathname();
  const last = useRef<string | null>(null);

  useEffect(() => {
    if (last.current !== pathname) {
      last.current = pathname;
      trackPageView(pathname);
    }
  }, [pathname]);

  return null;
}
