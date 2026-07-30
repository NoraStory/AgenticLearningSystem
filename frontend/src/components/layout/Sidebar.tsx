'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import {
  Home,
  Code2,
  FileCode2,
  Database,
  Binary,
  Bot,
  Terminal,
  Map,
  User,
  MessageSquare,
  Wrench,
  Brain,
  Network,
  Workflow,
  GraduationCap,
  FolderKanban,
  FileText,
} from 'lucide-react';

const menuItems = [
  { href: '/', icon: Home, label: '首页' },
  { href: '/python', icon: Code2, label: 'Python 编程' },
  { href: '/cpp', icon: FileCode2, label: 'C++ 编程' },
  { href: '/database', icon: Database, label: '数据库' },
  { href: '/algorithm', icon: Binary, label: '数据结构与算法' },
];

const agentItems = [
  { href: '/agent', icon: Bot, label: 'AI Agent' },
  { href: '/agent/chat', icon: MessageSquare, label: '智能对话' },
  { href: '/agent/tools', icon: Wrench, label: 'Tool 管理' },
  { href: '/agent/profile', icon: Brain, label: '用户画像' },
];

const secondaryItems = [
  { href: '/practice', icon: Terminal, label: '在线练习' },
  { href: '/interview', icon: GraduationCap, label: '笔试模拟' },
  { href: '/project', icon: FolderKanban, label: '项目实战' },
  { href: '/resume', icon: FileText, label: '简历分析优化' },
  { href: '/learning-path', icon: Map, label: '学习路径' },
  { href: '/profile', icon: User, label: '个人中心' },
];

export function Sidebar() {
  const pathname = usePathname();

  const isActive = (href: string) => {
    if (href === '/') return pathname === '/';
    return pathname.startsWith(href);
  };

  return (
    <aside className="w-52 shrink-0 bg-surface-container-lowest/50 border-r border-outline/10 overflow-y-auto">
      <nav className="p-3 space-y-1">
        {menuItems.map((item) => {
          const Icon = item.icon;
          const active = isActive(item.href);
          return (
            <Link
              key={item.href}
              href={item.href}
              className={`flex items-center gap-2.5 px-3 py-2 rounded-md text-sm transition-colors ${
                active
                  ? 'bg-primary/10 text-primary font-medium'
                  : 'text-muted-foreground hover:bg-surface-container hover:text-foreground'
              }`}
              aria-current={active ? 'page' : undefined}
            >
              <Icon className="w-4 h-4" />
              {item.label}
            </Link>
          );
        })}
        <div className="my-2 mx-3 border-t border-outline/10" />
        <div className="px-3 py-1 text-xs font-medium text-muted-foreground/60 uppercase tracking-wider">AI 助手</div>
        {agentItems.map((item) => {
          const Icon = item.icon;
          const active = isActive(item.href);
          return (
            <Link
              key={item.href}
              href={item.href}
              className={`flex items-center gap-2.5 px-3 py-2 rounded-md text-sm transition-colors ${
                active
                  ? 'bg-primary/10 text-primary font-medium'
                  : 'text-muted-foreground hover:bg-surface-container hover:text-foreground'
              }`}
              aria-current={active ? 'page' : undefined}
            >
              <Icon className="w-4 h-4" />
              {item.label}
            </Link>
          );
        })}
        <div className="my-2 mx-3 border-t border-outline/10" />
        {secondaryItems.map((item) => {
          const Icon = item.icon;
          const active = isActive(item.href);
          return (
            <Link
              key={item.href}
              href={item.href}
              className={`flex items-center gap-2.5 px-3 py-2 rounded-md text-sm transition-colors ${
                active
                  ? 'bg-primary/10 text-primary font-medium'
                  : 'text-muted-foreground hover:bg-surface-container hover:text-foreground'
              }`}
              aria-current={active ? 'page' : undefined}
            >
              <Icon className="w-4 h-4" />
              {item.label}
            </Link>
          );
        })}
      </nav>
      {/* 小板报：学习日历 */}
      <div className="mx-3 mt-4 p-3 bg-surface rounded-lg border border-outline/10">
        <div className="text-xs font-medium text-muted-foreground mb-2">
          学习小板报
        </div>
        <div className="space-y-1.5 text-xs">
          <div className="flex justify-between">
            <span className="text-muted-foreground">今日</span>
            <span className="text-primary font-medium">已学 45 分钟</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">本周</span>
            <span className="text-foreground">5/7 天</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">本月</span>
            <span className="text-foreground">18/30 天</span>
          </div>
          <div className="flex justify-between">
            <span className="text-muted-foreground">连续打卡</span>
            <span className="text-destructive font-medium">15 天</span>
          </div>
        </div>
      </div>
    </aside>
  );
}
