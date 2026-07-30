import type { Metadata } from 'next';
import './globals.css';
import { Header } from '@/components/layout/Header';
import { Sidebar } from '@/components/layout/Sidebar';

export const metadata: Metadata = {
  title: 'CodeForge Academy | 技术学习平台',
  description:
    'CodeForge Academy 是一个面向开发者的技术学习平台，涵盖 Rust 编程、数据结构与算法、AI Agent 开发三大方向。',
  keywords: ['Rust', '算法', 'AI Agent', '编程学习', '数据结构'],
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="zh-CN">
      <body className="bg-background text-foreground font-sans antialiased">
        <Header />
        <div className="flex" style={{ height: 'calc(100vh - 3.5rem)' }}>
          <Sidebar />
          <main className="flex-1 min-w-0 overflow-y-auto bg-background">
            {children}
          </main>
        </div>
      </body>
    </html>
  );
}
