'use client';

import Link from 'next/link';
import { Award, Clock, BookOpen, Code, Flame } from 'lucide-react';
import { useEffect, useState } from 'react';
import { apiFetch } from '@/lib/api';

// 用户数据
const fallbackUser = {
  name: '小初',
  avatar: '初',
  level: 'Lv.5',
  joinDate: '2024 年 3 月',
  days: 128,
  stats: {
    hours: 256,
    courses: 23,
    problems: 81,
    streak: 15,
  },
  progress: {
    python: 45,
    cpp: 32,
    database: 28,
    algorithm: 32,
    agent: 18,
  },
};

// 最近活动
const fallbackActivities = [
  { type: 'complete', text: '完成了 "Python 第3章：函数与模块"', time: '2小时前' },
  { type: 'solve', text: '解决了算法题 "#42 两数之和"', time: '5小时前' },
  { type: 'start', text: '开始学习 "LangChain 基础"', time: '昨天' },
  { type: 'achievement', text: '获得成就 "连续打卡7天"', time: '昨天' },
  { type: 'note', text: '新增笔记 "Python 装饰器总结"', time: '2天前' },
  { type: 'complete', text: '完成了 "C++ 第2章：指针与引用"', time: '3天前' },
];

// 成就
const fallbackAchievements = [
  { name: '初学乍练', desc: '完成第一个课程', unlocked: true },
  { name: '算法新星', desc: '解决10道算法题', unlocked: true },
  { name: '连续打卡7天', desc: '连续学习7天', unlocked: true },
  { name: 'Rust 入门', desc: '完成 Rust 基础课程', unlocked: true },
  { name: '学无止境', desc: '学习时长超过100小时', unlocked: true },
  { name: '速解达人', desc: '5分钟内解决一道简单题', unlocked: false },
];

// 收藏课程
const fallbackFavorites = [
  { id: 1, title: 'Rust 所有权机制深度解析', category: 'Rust' },
  { id: 2, title: '动态规划入门指南', category: '算法' },
  { id: 3, title: 'LangChain 链式调用', category: 'Agent' },
  { id: 4, title: '并发编程实战', category: 'Rust' },
];

// 笔记
const fallbackNotes = [
  { id: 1, title: '所有权规则总结', course: 'Rust 所有权机制', date: '2天前' },
  { id: 2, title: '动态规划解题思路', course: '动态规划入门', date: '3天前' },
  { id: 3, title: 'LangChain 核心概念', course: 'LangChain 基础', date: '5天前' },
];

const activityIcons: Record<string, string> = {
  complete: '✓',
  solve: '◉',
  start: '▶',
  achievement: '★',
  note: '✎',
};

const activityColors: Record<string, string> = {
  complete: 'text-success',
  solve: 'text-primary',
  start: 'text-muted-foreground',
  achievement: 'text-warning',
  note: 'text-info',
};

export default function ProfilePage() {
  const [user, setUser] = useState(fallbackUser);
  const [activities, setActivities] = useState(fallbackActivities);
  const [achievements, setAchievements] = useState(fallbackAchievements);
  const [favorites, setFavorites] = useState(fallbackFavorites);
  const [notes, setNotes] = useState(fallbackNotes);

  useEffect(() => {
    Promise.all([
      apiFetch<{ username: string; level: number; join_date: string; learning_days: number; stats: { total_hours: number; completed_courses: number; solved_problems: number; current_streak: number } }>('/api/v1/users/me'),
      apiFetch<Record<string, { progress?: number }>>('/api/v1/users/me/progress'),
      apiFetch<{ items: typeof fallbackActivities }>('/api/v1/users/me/activities'),
      apiFetch<{ items: typeof fallbackAchievements }>('/api/v1/users/me/achievements'),
      apiFetch<{ items: typeof fallbackFavorites }>('/api/v1/users/me/favorites'),
      apiFetch<{ items: typeof fallbackNotes }>('/api/v1/users/me/notes'),
    ]).then(([me, progress, activityData, achievementData, favoriteData, noteData]) => {
      setUser({
        name: me.username, avatar: String(me.username || '初')[0], level: 'Lv.' + me.level,
        joinDate: me.join_date, days: me.learning_days,
        stats: { hours: Math.round(me.stats.total_hours), courses: me.stats.completed_courses, problems: me.stats.solved_problems, streak: me.stats.current_streak },
        progress: { python: progress.python?.progress || 0, cpp: progress.cpp?.progress || 0, database: progress.database?.progress || 0, algorithm: progress.algorithm?.progress || 0, agent: progress.agent?.progress || 0 },
      });
      setActivities(activityData.items.length ? activityData.items : fallbackActivities);
      setAchievements(achievementData.items.length ? achievementData.items : fallbackAchievements);
      setFavorites(favoriteData.items);
      setNotes(noteData.items);
    }).catch(() => undefined);
  }, []);
  return (
    <>
      {/* 页面标题 */}
      <div className="px-8 pt-8 pb-6 border-b border-outline/10">
        <h1 className="text-2xl font-bold text-foreground">个人中心</h1>
      </div>

      {/* 内容区 */}
      <div className="flex gap-8 px-8 py-8">
        {/* 左侧：主要内容 */}
        <div className="flex-1 min-w-0">
          {/* 用户信息卡片 */}
          <div className="bg-surface rounded-lg shadow-card p-6 mb-6">
            <div className="flex items-center gap-4">
              <div className="w-16 h-16 rounded-full bg-primary/10 flex items-center justify-center text-xl font-bold text-primary">
                {user.avatar}
              </div>
              <div>
                <div className="flex items-center gap-2">
                  <h2 className="text-xl font-bold text-foreground">{user.name}</h2>
                  <span className="px-2 py-0.5 bg-primary/10 text-primary text-xs font-medium rounded-sm">
                    {user.level}
                  </span>
                </div>
                <p className="text-sm text-muted-foreground mt-1">
                  注册于 {user.joinDate} · 已学习 {user.days} 天
                </p>
              </div>
            </div>
          </div>

          {/* 学习统计 */}
          <div className="grid grid-cols-4 gap-4 mb-6">
            {[
              { icon: Clock, label: '学习时长', value: `${user.stats.hours}h`, color: 'text-primary' },
              { icon: BookOpen, label: '完成课程', value: `${user.stats.courses}个`, color: 'text-success' },
              { icon: Code, label: '解题数量', value: `${user.stats.problems}道`, color: 'text-info' },
              { icon: Flame, label: '连续打卡', value: `${user.stats.streak}天`, color: 'text-warning' },
            ].map((stat) => (
              <div key={stat.label} className="bg-surface rounded-lg shadow-card p-4 text-center">
                <stat.icon className={`w-5 h-5 mx-auto mb-2 ${stat.color}`} />
                <div className="text-lg font-bold text-foreground">{stat.value}</div>
                <div className="text-xs text-muted-foreground">{stat.label}</div>
              </div>
            ))}
          </div>

          {/* 学习进度 */}
          <div className="bg-surface rounded-lg shadow-card p-5 mb-6">
            <h3 className="text-sm font-semibold text-foreground mb-4">学习进度</h3>
            <div className="space-y-4">
              {[
                { name: 'Python 编程', progress: user.progress.python, color: 'bg-primary' },
                { name: 'C++ 编程', progress: user.progress.cpp, color: 'bg-info' },
                { name: '数据库', progress: user.progress.database, color: 'bg-success' },
                { name: '数据结构与算法', progress: user.progress.algorithm, color: 'bg-warning' },
                { name: 'AI Agent', progress: user.progress.agent, color: 'bg-destructive' },
              ].map((item) => (
                <div key={item.name}>
                  <div className="flex justify-between text-sm mb-1">
                    <span className="text-foreground">{item.name}</span>
                    <span className="text-muted-foreground">{item.progress}%</span>
                  </div>
                  <div className="h-2 bg-surface-container rounded-full overflow-hidden">
                    <div
                      className={`h-full ${item.color} rounded-full`}
                      style={{ width: `${item.progress}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* 最近活动 */}
          <div className="bg-surface rounded-lg shadow-card p-5">
            <h3 className="text-sm font-semibold text-foreground mb-4">最近活动</h3>
            <div className="space-y-4">
              {activities.map((activity, idx) => (
                <div key={idx} className="flex items-start gap-3">
                  <div className={`w-6 h-6 rounded-full bg-surface-container flex items-center justify-center text-xs ${activityColors[activity.type]}`}>
                    {activityIcons[activity.type]}
                  </div>
                  <div className="flex-1">
                    <p className="text-sm text-foreground">{activity.text}</p>
                    <p className="text-xs text-muted-foreground mt-0.5">{activity.time}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* 右侧边栏 */}
        <div className="w-[280px] shrink-0">
          <div className="sticky top-4 space-y-5">
            {/* 成就徽章 */}
            <div className="bg-surface rounded-lg shadow-card p-5">
              <h3 className="text-sm font-semibold text-foreground mb-4">
                成就徽章
              </h3>
              <div className="grid grid-cols-3 gap-3">
                {achievements.map((badge) => (
                  <div
                    key={badge.name}
                    className={`text-center p-2 rounded-lg ${
                      badge.unlocked
                        ? 'bg-primary/5'
                        : 'bg-surface-container opacity-50'
                    }`}
                    title={badge.desc}
                  >
                    <Award className={`w-5 h-5 mx-auto mb-1 ${
                      badge.unlocked ? 'text-primary' : 'text-muted-foreground'
                    }`} />
                    <p className="text-xs text-foreground truncate">{badge.name}</p>
                  </div>
                ))}
              </div>
            </div>

            {/* 收藏课程 */}
            <div className="bg-surface rounded-lg shadow-card p-5">
              <h3 className="text-sm font-semibold text-foreground mb-4">
                收藏课程
              </h3>
              <div className="space-y-2">
                {favorites.map((fav) => (
                  <Link
                    key={fav.id}
                    href={`/course/${fav.id}`}
                    className="block text-sm text-muted-foreground hover:text-primary transition-colors py-1"
                  >
                    <span className="text-xs text-primary mr-2">[{fav.category}]</span>
                    {fav.title}
                  </Link>
                ))}
              </div>
            </div>

            {/* 我的笔记 */}
            <div className="bg-surface rounded-lg shadow-card p-5">
              <h3 className="text-sm font-semibold text-foreground mb-4">
                我的笔记
              </h3>
              <div className="space-y-2">
                {notes.map((note) => (
                  <Link
                    key={note.id}
                    href={`/course/1`}
                    className="block py-1"
                  >
                    <p className="text-sm text-foreground hover:text-primary transition-colors">
                      {note.title}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {note.course} · {note.date}
                    </p>
                  </Link>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
