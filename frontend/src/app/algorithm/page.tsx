'use client';

import Link from 'next/link';
import { useEffect, useState } from 'react';
import { apiFetch } from '@/lib/api';
import { Target } from 'lucide-react';

// 题目数据
const fallbackProblems = [
  { id: 1, title: '两数之和', category: '数组', difficulty: '简单', passRate: '48.5%', status: 'solved' },
  { id: 2, title: '反转链表', category: '链表', difficulty: '简单', passRate: '72.3%', status: 'solved' },
  { id: 3, title: '有效的括号', category: '栈', difficulty: '简单', passRate: '44.1%', status: 'solved' },
  { id: 4, title: '合并K个升序链表', category: '链表', difficulty: '困难', passRate: '35.2%', status: 'attempted' },
  { id: 5, title: 'LRU 缓存机制', category: '哈希表', difficulty: '中等', passRate: '52.3%', status: 'not-started' },
  { id: 6, title: '二叉树层序遍历', category: '树', difficulty: '中等', passRate: '65.8%', status: 'solved' },
  { id: 7, title: '课程表', category: '图', difficulty: '中等', passRate: '48.9%', status: 'not-started' },
  { id: 8, title: '最长递增子序列', category: '动态规划', difficulty: '中等', passRate: '38.7%', status: 'not-started' },
  { id: 9, title: '接雨水', category: '双指针', difficulty: '困难', passRate: '28.5%', status: 'not-started' },
  { id: 10, title: '无重复字符的最长子串', category: '滑动窗口', difficulty: '中等', passRate: '41.2%', status: 'attempted' },
  { id: 11, title: '三数之和', category: '双指针', difficulty: '中等', passRate: '36.8%', status: 'not-started' },
  { id: 12, title: '全排列', category: '回溯', difficulty: '中等', passRate: '55.4%', status: 'not-started' },
];

const categories = ['全部', '数组', '链表', '栈', '哈希表', '树', '图', '排序', '查找', '动态规划', '贪心', '回溯', '双指针', '滑动窗口'];

const difficultyColors: Record<string, string> = {
  '简单': 'bg-green-500/10 text-green-600',
  '中等': 'bg-orange-500/10 text-orange-600',
  '困难': 'bg-red-500/10 text-red-600',
};

const statusIcons: Record<string, { icon: string; color: string }> = {
  solved: { icon: '✓', color: 'text-success' },
  attempted: { icon: '◐', color: 'text-warning' },
  'not-started': { icon: '○', color: 'text-muted-foreground' },
};

export default function AlgorithmPage() {
  const [categoryFilter, setCategoryFilter] = useState('全部');
  const [difficultyFilter, setDifficultyFilter] = useState('all');
  const [problems, setProblems] = useState(fallbackProblems);

  useEffect(() => {
    apiFetch<{ items: Array<{ id: number; title: string; category: string; difficulty: string; pass_rate: number; status: string }> }>('/api/v1/problems')
      .then((data) => setProblems(data.items.map((problem) => ({
        id: problem.id, title: problem.title, category: problem.category, difficulty: problem.difficulty,
        passRate: `${problem.pass_rate.toFixed(1)}%`, status: problem.status.replace('_', '-'),
      }))))
      .catch(() => undefined);
  }, []);

  const filtered = problems.filter((p) => {
    const matchCategory = categoryFilter === '全部' || p.category === categoryFilter;
    const matchDifficulty = difficultyFilter === 'all' || p.difficulty === difficultyFilter;
    return matchCategory && matchDifficulty;
  });

  return (
    <>
      {/* 页面标题 */}
      <div className="px-8 pt-8 pb-6 border-b border-outline/10">
        <h1 className="text-2xl font-bold text-foreground">数据结构与算法</h1>
        <p className="text-sm text-muted-foreground mt-1.5">
          系统掌握数据结构与算法，提升编程能力和面试竞争力
        </p>
        <div className="flex items-center gap-4 mt-3 text-sm">
          <span className="text-muted-foreground">
            题目总数 <span className="font-medium text-foreground">{problems.length}</span>
          </span>
          <span className="text-muted-foreground">
            已解决 <span className="font-medium text-success">{problems.filter((p) => p.status === 'solved').length}</span>
          </span>
          <span className="text-muted-foreground">
            尝试中 <span className="font-medium text-warning">{problems.filter((p) => p.status === 'attempted').length}</span>
          </span>
        </div>
      </div>

      {/* 内容区 */}
      <div className="flex gap-8 px-8 py-8">
        {/* 左侧：题目列表 */}
        <div className="flex-1 min-w-0">
          {/* 分类标签 */}
          <div className="flex flex-wrap gap-2 mb-6">
            {categories.map((cat) => (
              <button
                key={cat}
                onClick={() => setCategoryFilter(cat)}
                className={`px-3 py-1 text-xs rounded-sm transition-colors ${
                  categoryFilter === cat
                    ? 'font-medium text-primary bg-primary/10'
                    : 'text-muted-foreground hover:text-foreground hover:bg-surface-container'
                }`}
              >
                {cat}
              </button>
            ))}
          </div>

          {/* 筛选器 */}
          <div className="flex items-center gap-3 mb-4">
            <select
              value={difficultyFilter}
              onChange={(e) => setDifficultyFilter(e.target.value)}
              className="text-sm border border-outline rounded-md px-3 py-1.5 bg-surface text-foreground"
            >
              <option value="all">全部难度</option>
              <option value="简单">简单</option>
              <option value="中等">中等</option>
              <option value="困难">困难</option>
            </select>
            <span className="text-xs text-muted-foreground">
              共 {filtered.length} 题
            </span>
          </div>

          {/* 题目列表 */}
          <div className="space-y-3">
            {filtered.map((problem) => (
              <div
                key={problem.id}
                className="flex items-center justify-between p-4 bg-surface rounded-lg border border-outline/10 hover:shadow-card transition-shadow"
              >
                <div className="flex items-center gap-4">
                  <span className={`text-lg ${statusIcons[problem.status].color}`}>
                    {statusIcons[problem.status].icon}
                  </span>
                  <div>
                    <Link
                      href={`/practice?id=${problem.id}`}
                      className="text-sm font-medium text-foreground hover:text-primary transition-colors"
                    >
                      {problem.id}. {problem.title}
                    </Link>
                    <div className="flex items-center gap-2 mt-1">
                      <span className="text-xs text-muted-foreground">
                        {problem.category}
                      </span>
                      <span
                        className={`text-xs px-1.5 py-0.5 rounded-sm ${difficultyColors[problem.difficulty]}`}
                      >
                        {problem.difficulty}
                      </span>
                    </div>
                  </div>
                </div>
                <div className="flex items-center gap-4">
                  <span className="text-xs text-muted-foreground">
                    通过率 {problem.passRate}
                  </span>
                  <Link
                    href={`/practice?id=${problem.id}`}
                    className="px-3 py-1.5 text-xs font-medium text-primary bg-primary/10 hover:bg-primary/20 rounded-md transition-colors"
                  >
                    开始练习
                  </Link>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* 右侧边栏 */}
        <div className="w-[280px] shrink-0">
          <div className="sticky top-4 space-y-5">
            {/* 每日一题 */}
            <div className="bg-surface rounded-lg shadow-card p-5 border-l-4 border-primary">
              <div className="flex items-center gap-2 mb-2">
                <Target className="w-4 h-4 text-primary" />
                <h3 className="text-sm font-semibold text-foreground">
                  每日一题
                </h3>
              </div>
              <Link
                href="/practice?id=200"
                className="text-sm font-medium text-foreground hover:text-primary transition-colors block mb-2"
              >
                #200. 岛屿数量
              </Link>
              <div className="flex items-center gap-2 text-xs">
                <span className="px-2 py-0.5 bg-orange-500/10 text-orange-600 rounded-sm font-medium">
                  中等
                </span>
                <span className="text-muted-foreground">通过率 58.2%</span>
              </div>
            </div>

            {/* 难度统计 */}
            <div className="bg-surface rounded-lg shadow-card p-5">
              <h3 className="text-sm font-semibold text-foreground mb-4">
                解题统计
              </h3>
              <div className="space-y-3">
                {[
                  { label: '简单', solved: 45, total: 120, color: 'bg-green-500' },
                  { label: '中等', solved: 28, total: 180, color: 'bg-orange-500' },
                  { label: '困难', solved: 8, total: 60, color: 'bg-red-500' },
                ].map((item) => (
                  <div key={item.label}>
                    <div className="flex justify-between text-xs mb-1">
                      <span className="text-foreground">{item.label}</span>
                      <span className="text-muted-foreground">
                        {item.solved}/{item.total}
                      </span>
                    </div>
                    <div className="h-1.5 bg-surface-container rounded-full overflow-hidden">
                      <div
                        className={`h-full ${item.color} rounded-full`}
                        style={{
                          width: `${(item.solved / item.total) * 100}%`,
                        }}
                      />
                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* 热门标签 */}
            <div className="bg-surface rounded-lg shadow-card p-5">
              <h3 className="text-sm font-semibold text-foreground mb-4">
                热门标签
              </h3>
              <div className="flex flex-wrap gap-2">
                {['数组', '链表', '树', 'DP', '回溯', '双指针', '图', '栈'].map(
                  (tag) => (
                    <span
                      key={tag}
                      className="text-xs text-muted-foreground bg-surface-container hover:bg-surface-container-high px-2.5 py-1 rounded-sm cursor-pointer transition-colors"
                    >
                      {tag}
                    </span>
                  )
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
