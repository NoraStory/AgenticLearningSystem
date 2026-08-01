'use client';

import Link from 'next/link';
import { Flame, Clock, Users, BookOpen } from 'lucide-react';
import { useEffect, useState } from 'react';
import { apiFetch } from '@/lib/api';

// 文章数据
const fallbackArticles = [
  {
    id: 1,
    category: 'rust',
    categoryLabel: 'Rust',
    categoryColor: 'bg-orange-500/10 text-orange-600',
    date: '2024 年 1 月 15 日',
    title: 'Rust 所有权机制深度解析',
    summary:
      '所有权系统是 Rust 最独特的特性，它让 Rust 无需垃圾回收器即可保障内存安全。本文深入探讨所有权、借用和生命周期的核心概念，通过大量实例帮助你建立直觉。',
    tags: ['入门', '所有权', '内存安全'],
    image: 'https://images.unsplash.com/photo-1555066931-4365d14bab8c?w=800&h=400&fit=crop',
  },
  {
    id: 2,
    category: 'algorithm',
    categoryLabel: '算法',
    categoryColor: 'bg-blue-500/10 text-blue-600',
    date: '2024 年 1 月 13 日',
    title: '动态规划入门：从递归到记忆化搜索',
    summary:
      '动态规划是算法面试中的高频考点。本文将从最简单的递归解法出发，逐步引入记忆化搜索和状态转移方程，帮你建立完整的 DP 思维框架。',
    tags: ['DP', '入门', '递归'],
    image: 'https://images.unsplash.com/photo-1509228468518-180dd4864904?w=800&h=400&fit=crop',
  },
  {
    id: 3,
    category: 'agent',
    categoryLabel: 'Agent',
    categoryColor: 'bg-purple-500/10 text-purple-600',
    date: '2024 年 1 月 11 日',
    title: '构建你的第一个 AI Agent：从原理到实践',
    summary:
      'AI Agent 正在改变软件开发的方式。本文介绍 Agent 的核心架构——感知、规划、行动循环，并手把手带你用 LangChain 构建一个能调用工具的简单 Agent。',
    tags: ['LLM', '工具调用', 'LangChain'],
    image: 'https://images.unsplash.com/photo-1677442136019-21780ecad995?w=800&h=400&fit=crop',
  },
  {
    id: 4,
    category: 'rust',
    categoryLabel: 'Rust',
    categoryColor: 'bg-orange-500/10 text-orange-600',
    date: '2024 年 1 月 9 日',
    title: 'Rust 并发编程：无畏并发的实践指南',
    summary:
      'Rust 的类型系统在编译期就能防止数据竞争。本文详解线程、消息传递、共享状态等并发模式，以及 async/await 异步编程的实战技巧。',
    tags: ['并发', 'async', '线程'],
    image: 'https://images.unsplash.com/photo-1517694712202-14dd9538aa97?w=800&h=400&fit=crop',
  },
  {
    id: 5,
    category: 'algorithm',
    categoryLabel: '算法',
    categoryColor: 'bg-blue-500/10 text-blue-600',
    date: '2024 年 1 月 7 日',
    title: '图算法实战：最短路径与网络流',
    summary:
      '图论算法在实际工程中应用广泛。本文以 Dijkstra 和 Ford-Fulkerson 为核心，结合地图导航和资源分配两个场景，讲解图算法的实现与优化策略。',
    tags: ['图论', 'Dijkstra', '进阶'],
    image: 'https://images.unsplash.com/photo-1551288049-bebda4e38f71?w=800&h=400&fit=crop',
  },
];

// 热门标签
const fallbackHotTags = [
  'Rust', '所有权', '算法', 'DP', 'LLM', 'LangChain',
  '并发', '图论', 'Agent', '工具调用',
];

// 推荐课程
const fallbackRecommendedCourses: { title: string; category: string; reason?: string }[] = [
  { title: 'Rust 入门到精通', category: 'Rust' },
  { title: '算法面试突破 100 题', category: '算法' },
  { title: 'LangChain 实战教程', category: 'Agent' },
];

export default function HomePage() {
  const [filter, setFilter] = useState<string>('all');
  const [articles, setArticles] = useState(fallbackArticles);
  const [hotTags, setHotTags] = useState(fallbackHotTags);
  const [recommendedCourses, setRecommendedCourses] = useState(fallbackRecommendedCourses);
  const [userName, setUserName] = useState('小初');
  const [streakDays, setStreakDays] = useState(15);

  useEffect(() => {
    Promise.all([
      apiFetch<{ items: typeof fallbackArticles }>('/api/v1/courses?page_size=20'),
      apiFetch<{ tags: string[] }>('/api/v1/tags/hot'),
      apiFetch<{ items: typeof fallbackRecommendedCourses }>('/api/v1/courses/recommended/weak'),
      apiFetch<{ username: string }>('/api/v1/users/me'),
      apiFetch<{ streak_days: number }>('/api/v1/users/me/streak'),
    ]).then(([courseData, tagData, recommendedData, user, streak]) => {
      setArticles(courseData.items.length ? courseData.items : fallbackArticles);
      setHotTags(tagData.tags);
      setRecommendedCourses(recommendedData.items);
      setUserName(user.username);
      setStreakDays(streak.streak_days);
    }).catch(() => undefined);
  }, []);

  const filteredArticles =
    filter === 'all'
      ? articles
      : articles.filter((a) => a.category === filter);

  return (
    <>
      {/* 欢迎横幅 */}
      <div className="px-8 pt-8 pb-6">
        <h1 className="text-2xl font-bold text-foreground">欢迎回来，{userName}</h1>
        <p className="text-sm text-muted-foreground mt-1.5 flex items-center gap-1.5">
          <Flame className="w-4 h-4 text-destructive" />
          已连续学习 {streakDays} 天，继续保持！
        </p>
      </div>

      {/* 内容区：博客流 + 右侧边栏 */}
      <div className="flex gap-8 px-8 pb-8">
        {/* 左侧：博客文章流 */}
        <div className="flex-1 min-w-0">
          {/* 分区标题 + 筛选标签 */}
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-lg font-semibold text-foreground">最新内容</h2>
            <div className="flex items-center gap-1">
              {[
                { key: 'all', label: '全部' },
                { key: 'python', label: 'Python' },
                { key: 'cpp', label: 'C++' },
                { key: 'database', label: '数据库' },
                { key: 'algorithm', label: '算法' },
                { key: 'agent', label: 'Agent' },
              ].map((tab) => (
                <button
                  key={tab.key}
                  onClick={() => setFilter(tab.key)}
                  className={`px-3 py-1.5 text-sm rounded-md transition-colors ${
                    filter === tab.key
                      ? 'font-medium text-primary bg-primary/10'
                      : 'text-muted-foreground hover:text-foreground hover:bg-surface-container'
                  }`}
                >
                  {tab.label}
                </button>
              ))}
            </div>
          </div>

          {/* 文章列表 */}
          <div>
            {filteredArticles.map((article) => (
              <div
                key={article.id}
                className="pb-8 mb-8 border-b border-outline/10"
              >
                <Link href={`/course/${article.id}`} className="group block">
                  <div className="overflow-hidden rounded-lg mb-4">
                    <img
                      src={article.image}
                      alt={article.title}
                      className="w-full h-52 object-cover group-hover:scale-[1.02] transition-transform duration-300"
                    />
                  </div>
                  <div className="space-y-2.5">
                    <div className="flex items-center gap-2">
                      <span
                        className={`inline-flex items-center px-2 py-0.5 rounded-sm text-xs font-medium ${article.categoryColor}`}
                      >
                        {article.categoryLabel}
                      </span>
                      <span className="text-xs text-muted-foreground">
                        {article.date}
                      </span>
                    </div>
                    <h2 className="text-xl font-bold text-foreground group-hover:text-primary transition-colors leading-snug">
                      {article.title}
                    </h2>
                    <p className="text-sm text-muted-foreground leading-relaxed">
                      {article.summary}
                    </p>
                    <div className="flex items-center gap-2 pt-1">
                      {article.tags.map((tag) => (
                        <span
                          key={tag}
                          className="text-xs text-muted-foreground bg-surface-container px-2 py-0.5 rounded-sm"
                        >
                          {tag}
                        </span>
                      ))}
                    </div>
                  </div>
                </Link>
              </div>
            ))}
          </div>

          {/* 查看更多 */}
          <div className="text-center pt-2 pb-4">
            <button className="px-6 py-2.5 text-sm font-medium text-muted-foreground bg-surface-container hover:bg-surface-container-high rounded-md transition-colors">
              查看更多文章
            </button>
          </div>
        </div>

        {/* 右侧边栏 */}
        <div className="w-[280px] shrink-0">
          <div className="sticky top-4 space-y-5">
            {/* 学习进度 */}
            <div className="bg-surface rounded-lg shadow-card p-5">
              <h3 className="text-sm font-semibold text-foreground mb-4">
                学习进度
              </h3>
              <div className="space-y-4">
                {[
                  { name: 'Python 编程', progress: 45, color: 'bg-primary' },
                  { name: 'C++ 编程', progress: 32, color: 'bg-info' },
                  { name: '数据库', progress: 28, color: 'bg-success' },
                  { name: '数据结构与算法', progress: 42, color: 'bg-warning' },
                  { name: 'AI Agent', progress: 25, color: 'bg-destructive' },
                ].map((item) => (
                  <div key={item.name}>
                    <div className="flex justify-between text-xs mb-1.5">
                      <span className="text-foreground font-medium">
                        {item.name}
                      </span>
                      <span className="text-muted-foreground">
                        {item.progress}%
                      </span>
                    </div>
                    <div className="h-1.5 bg-surface-container rounded-full overflow-hidden">
                      <div
                        className={`h-full ${item.color} rounded-full transition-all`}
                        style={{ width: `${item.progress}%` }}
                      />
                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* 每日一题 */}
            <div className="bg-surface rounded-lg shadow-card p-5 border-l-4 border-primary">
              <div className="flex items-center gap-2 mb-2">
                <BookOpen className="w-4 h-4 text-primary" />
                <h3 className="text-sm font-semibold text-foreground">
                  每日一题
                </h3>
              </div>
              <p className="text-xs text-muted-foreground mb-2">
                今日推荐挑战
              </p>
              <Link
                href="/practice"
                className="text-sm font-medium text-foreground hover:text-primary transition-colors block mb-3"
              >
                #146. LRU 缓存机制
              </Link>
              <div className="flex items-center gap-3 text-xs text-muted-foreground mb-3">
                <span className="px-2 py-0.5 bg-orange-500/10 text-orange-600 rounded-sm font-medium">
                  中等
                </span>
                <span className="flex items-center gap-1">
                  <Users className="w-3 h-3" />
                  通过率 52.3%
                </span>
              </div>
              <Link
                href="/practice"
                className="block w-full text-center py-2 text-xs font-medium text-primary bg-primary/10 hover:bg-primary/20 rounded-md transition-colors"
              >
                开始挑战
              </Link>
            </div>

            {/* 热门标签 */}
            <div className="bg-surface rounded-lg shadow-card p-5">
              <h3 className="text-sm font-semibold text-foreground mb-4">
                热门标签
              </h3>
              <div className="flex flex-wrap gap-2">
                {hotTags.map((tag) => (
                  <span
                    key={tag}
                    className="text-xs text-muted-foreground bg-surface-container hover:bg-surface-container-high px-2.5 py-1 rounded-sm cursor-pointer transition-colors"
                  >
                    {tag}
                  </span>
                ))}
              </div>
            </div>

            {/* 推荐课程 */}
            <div className="bg-surface rounded-lg shadow-card p-5">
              <h3 className="text-sm font-semibold text-foreground mb-4">
                推荐课程
              </h3>
              <div className="space-y-3">
                {recommendedCourses.map((course) => (
                  <Link
                    key={course.title}
                    href="/course/1"
                    className="flex items-center gap-3 group"
                  >
                    <div className="w-10 h-10 rounded-md bg-surface-container flex items-center justify-center text-xs font-medium text-muted-foreground group-hover:bg-primary/10 group-hover:text-primary transition-colors">
                      {course.category.slice(0, 2)}
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-foreground group-hover:text-primary transition-colors truncate">
                        {course.title}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {course.category}
                      </p>
                      {course.reason && (
                        <p className="text-[11px] text-primary/70 mt-0.5 truncate">
                          {course.reason}
                        </p>
                      )}
                    </div>
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
