'use client';

import { memo, useEffect, useRef, useState } from 'react';

// mermaid 库较大(~2MB),客户端懒加载,避免 SSR 水合问题
type MermaidAPI = typeof import('mermaid').default;

async function loadMermaid(): Promise<MermaidAPI> {
  const mod = await import('mermaid');
  const mermaid = mod.default;
  mermaid.initialize({
    startOnLoad: false,
    theme: 'default',
    securityLevel: 'strict',
    fontFamily: 'inherit',
  });
  return mermaid;
}

export function Mermaid({ source }: { source: string }) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [svg, setSvg] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    loadMermaid()
      .then(async (mermaid) => {
        if (cancelled) return;
        try {
          const { svg } = await mermaid.render(`mmd-${crypto.randomUUID().replace(/-/g, '')}`, source);
          if (!cancelled) setSvg(svg);
        } catch (e) {
          if (!cancelled) setError(e instanceof Error ? e.message : String(e));
        }
      })
      .catch(() => setError('图表组件加载失败'));
    return () => {
      cancelled = true;
    };
  }, [source]);

  if (error) {
    return <pre className="mt-3 overflow-x-auto rounded-lg bg-slate-950 p-3 text-xs text-slate-100">{source}</pre>;
  }
  if (!svg) {
    return <div className="mt-3 p-3 text-xs text-muted-foreground animate-pulse">图表渲染中...</div>;
  }
  return (
    <div
      ref={containerRef}
      className="mt-3 overflow-x-auto rounded-lg border border-border bg-background p-2"
      dangerouslySetInnerHTML={{ __html: svg }}
    />
  );
}

export default memo(Mermaid);
