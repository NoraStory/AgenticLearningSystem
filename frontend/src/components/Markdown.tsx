'use client';

import { memo, useEffect, useState } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeRaw from 'rehype-raw';

// rehype-pretty-code 是异步插件,需要先初始化 shiki highlighter
async function loadRehypePrettyCode() {
  const { createHighlighter } = await import('shiki');
  const highlighter = await createHighlighter({
    themes: ['github-dark', 'github-light'],
    langs: ['ts', 'tsx', 'js', 'jsx', 'go', 'python', 'cpp', 'c', 'rust', 'sql', 'bash', 'json', 'yaml', 'html', 'css'],
  });
  const { default: rehypePrettyCode } = await import('rehype-pretty-code');
  return rehypePrettyCode({
    theme: 'github-dark',
    keepBackground: false,
    getHighlighter: async () => highlighter,
  });
}

type RehypePlugin = Awaited<ReturnType<typeof loadRehypePrettyCode>>;

const Markdown = memo(function Markdown({ children }: { children: string }) {
  const [rehypePlugin, setRehypePlugin] = useState<RehypePlugin | null>(null);

  useEffect(() => {
    let cancelled = false;
    loadRehypePrettyCode()
      .then((plugin) => {
        if (!cancelled) setRehypePlugin(plugin);
      })
      .catch(() => undefined); // 高亮加载失败时退化为普通 markdown
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <div className="text-sm leading-relaxed">
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[rehypeRaw, ...(rehypePlugin ? [rehypePlugin] : [])] as never}
        components={{
          h1: (props) => <h1 className="text-xl font-bold mt-4 mb-2" {...props} />,
          h2: (props) => <h2 className="text-lg font-bold mt-3 mb-2" {...props} />,
          h3: (props) => <h3 className="text-base font-semibold mt-3 mb-1" {...props} />,
          p: (props) => <p className="my-2" {...props} />,
          ul: (props) => <ul className="list-disc pl-5 my-2 space-y-1" {...props} />,
          ol: (props) => <ol className="list-decimal pl-5 my-2 space-y-1" {...props} />,
          li: (props) => <li className="leading-relaxed" {...props} />,
          a: (props) => (
            <a
              className="text-primary underline underline-offset-2"
              target="_blank"
              rel="noopener noreferrer"
              {...props}
            />
          ),
          strong: (props) => <strong className="font-semibold" {...props} />,
          table: (props) => (
            <div className="my-3 overflow-x-auto">
              <table className="border-collapse text-sm w-full" {...props} />
            </div>
          ),
          th: (props) => <th className="border border-border px-3 py-1.5 text-left bg-muted/50" {...props} />,
          td: (props) => <td className="border border-border px-3 py-1.5" {...props} />,
          blockquote: (props) => (
            <blockquote className="border-l-4 border-primary/40 pl-3 my-2 text-muted-foreground italic" {...props} />
          ),
          code: (props) => {
            const { className } = props as { className?: string };
            const isBlock = /language-/.test(className || '');
            if (isBlock) {
              return (
                <code className="block bg-muted/60 rounded-lg p-4 overflow-x-auto text-[13px] leading-relaxed" {...props} />
              );
            }
            return <code className="px-1.5 py-0.5 bg-muted/60 rounded text-[13px] font-mono" {...props} />;
          },
          pre: (props) => (
            <pre className="my-3 rounded-lg overflow-x-auto p-0 bg-transparent" {...props} />
          ),
        }}
      >
        {children}
      </ReactMarkdown>
    </div>
  );
});

export default Markdown;
