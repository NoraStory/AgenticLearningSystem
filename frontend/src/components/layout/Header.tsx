'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { Braces, Search } from 'lucide-react';

const navItems = [
  { href: '/', label: '首页' },
  { href: '/python', label: 'Python' },
  { href: '/cpp', label: 'C++' },
  { href: '/database', label: '数据库' },
  { href: '/algorithm', label: '算法' },
  { href: '/agent', label: 'Agent' },
  { href: '/practice', label: '练习' },
  { href: '/interview', label: '面试' },
  { href: '/project', label: '项目' },
  { href: '/learning-path', label: '路径' },
];

export function Header() {
  const pathname = usePathname();

  return (
    <header className="bg-surface/80 backdrop-blur-sm sticky top-0 z-40 h-14 flex items-center justify-between px-6 border-b border-outline/20">
      <Link href="/" className="flex items-center gap-2.5">
        <Braces className="text-primary w-5 h-5" />
        <span className="font-bold text-lg text-foreground">CodeForge</span>
        <span className="text-[10px] text-primary-foreground bg-primary px-1.5 py-0.5 rounded-sm font-medium">
          Academy
        </span>
      </Link>
      <nav className="flex items-center gap-1">
        {navItems.map((item) => (
          <Link
            key={item.href}
            href={item.href}
            className={`px-3 py-1.5 text-sm transition-colors ${
              pathname === item.href
                ? 'font-medium text-primary'
                : 'text-muted-foreground hover:text-primary'
            }`}
            aria-current={pathname === item.href ? 'page' : undefined}
          >
            {item.label}
          </Link>
        ))}
      </nav>
      <div className="flex items-center gap-3">
        <button className="p-2 text-muted-foreground hover:text-primary transition-colors">
          <Search className="w-4 h-4" />
        </button>
        <Link
          href="/profile"
          className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center text-xs font-medium text-primary cursor-pointer hover:bg-primary/20 transition-colors"
        >
          初
        </Link>
      </div>
    </header>
  );
}
