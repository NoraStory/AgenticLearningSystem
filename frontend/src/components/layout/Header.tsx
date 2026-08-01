'use client';

import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { useCallback, useEffect, useState } from 'react';
import { Braces, LogOut, Search } from 'lucide-react';
import { clearAuth, fetchMe, getToken, isLoggedIn } from '@/lib/api';

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
  const router = useRouter();
  const [username, setUsername] = useState<string | null>(null);

  // 登录态变化时刷新用户信息(token 变化后 localStorage 内容已更新)
  const refreshUser = useCallback(() => {
    if (!isLoggedIn()) {
      setUsername(null);
      return;
    }
    fetchMe()
      .then((me) => setUsername(me.username))
      .catch(() => setUsername(null)); // token 失效视为未登录
  }, []);

  useEffect(() => {
    refreshUser();
  }, [pathname, refreshUser]); // 路由变化时刷新(登录/登出后跳转)

  const handleLogout = () => {
    clearAuth();
    setUsername(null);
    router.push('/');
    router.refresh();
  };

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
        {username ? (
          <div className="flex items-center gap-2">
            <Link
              href="/profile"
              className="flex items-center gap-2 px-2 py-1 rounded-lg hover:bg-surface-container transition-colors cursor-pointer"
              title={username}
            >
              <span className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center text-xs font-medium text-primary">
                {username.slice(0, 1).toUpperCase()}
              </span>
              <span className="text-sm text-foreground max-w-[8rem] truncate">{username}</span>
            </Link>
            <button
              onClick={handleLogout}
              className="p-2 text-muted-foreground hover:text-destructive transition-colors"
              title="退出登录"
            >
              <LogOut className="w-4 h-4" />
            </button>
          </div>
        ) : (
          <Link
            href="/auth"
            className="px-4 py-1.5 rounded-lg bg-primary text-primary-foreground text-sm font-medium hover:opacity-90 transition-opacity"
          >
            登录 / 注册
          </Link>
        )}
      </div>
    </header>
  );
}
