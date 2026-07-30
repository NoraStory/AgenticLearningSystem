'use client';

import Link from 'next/link';
import { useEffect, useState } from 'react';
import { apiFetch } from '@/lib/api';
import { Check, Lock, Play } from 'lucide-react';

// 路径数据
const fallbackPaths = [
  { id: 'python', name: 'Python 全栈开发', courses: 30, levels: '入门 → 高级' },
  { id: 'cpp', name: 'C++ 系统编程', courses: 28, levels: '入门 → 高级' },
  { id: 'database', name: '数据库工程师', courses: 22, levels: '入门 → 高级' },
  { id: 'algorithm', name: '算法面试突破', courses: 50, levels: '简单 → 困难' },
  { id: 'agent', name: 'AI Agent 开发', courses: 24, levels: '入门 → 高级' },
];

// 阶段数据（以 Python 路径为例）
const fallbackStages = [
  {
    id: 1,
    name: 'Python 基础语法',
    status: 'completed',
    hours: 8,
    goal: '掌握 Python 基本语法、变量、类型和控制流',
    courses: ['环境搭建', '变量与类型', '控制流', '函数'],
    prerequisite: '无',
  },
  {
    id: 2,
    name: '面向对象编程',
    status: 'completed',
    hours: 10,
    goal: '深入理解类、继承、多态',
    courses: ['类与对象', '继承', '多态', '特殊方法'],
    prerequisite: '阶段 1',
  },
  {
    id: 3,
    name: '标准库与文件操作',
    status: 'in-progress',
    hours: 12,
    goal: '掌握 Python 标准库和文件操作',
    courses: ['文件读写', 'OS 模块', 'JSON 处理', '正则表达式'],
    prerequisite: '阶段 2',
  },
  {
    id: 4,
    name: 'Web 开发',
    status: 'locked',
    hours: 15,
    goal: '学习 Flask/Django Web 框架',
    courses: ['Flask 入门', '路由与视图', '模板引擎', '数据库集成'],
    prerequisite: '阶段 3',
  },
  {
    id: 5,
    name: '数据分析与机器学习',
    status: 'locked',
    hours: 14,
    goal: '学习 NumPy、Pandas、Scikit-learn',
    courses: ['NumPy 基础', 'Pandas 数据处理', '数据可视化', '机器学习入门'],
    prerequisite: '阶段 4',
  },
  {
    id: 6,
    name: '项目部署与优化',
    status: 'locked',
    hours: 8,
    goal: '将项目部署到生产环境',
    courses: ['Docker 容器化', '性能优化', '日志与监控', 'CI/CD'],
    prerequisite: '阶段 5',
  },
];

// 推荐资源
const resources = {
  books: ['Python Crash Course', 'Fluent Python', 'Python Cookbook'],
  videos: ['Python 入门到精通', 'Python Web 开发实战'],
  platforms: ['Rustlings', 'Exercism Rust Track'],
  docs: ['官方文档', 'Rust API 文档', 'Rust by Example'],
};

export default function LearningPathPage() {
  const [selectedPath, setSelectedPath] = useState('python');
  const [paths, setPaths] = useState(fallbackPaths);
  const [stages, setStages] = useState(fallbackStages);

  useEffect(() => {
    apiFetch<{ items: Array<{ id: string; name: string; courses: number; levels: number }> }>('/api/v1/learning-paths')
      .then((data) => setPaths(data.items.map((path) => ({ ...path, levels: path.levels + ' 个阶段' }))))
      .catch(() => undefined);
  }, []);

  useEffect(() => {
    apiFetch<{ items: Array<{ id: string; name: string; status: string; hours: number; goal: string; courses: string[]; prerequisite: string }> }>('/api/v1/learning-paths/' + selectedPath + '/stages')
      .then((data) => setStages(data.items.map((stage, index) => ({ id: index + 1, name: stage.name, status: stage.status.replace('_', '-'), hours: stage.hours, goal: stage.goal, courses: stage.courses, prerequisite: stage.prerequisite }))))
      .catch(() => undefined);
  }, [selectedPath]);

  return (
    <>
      {/* 页面标题 */}
      <div className="px-8 pt-8 pb-6 border-b border-outline/10">
        <h1 className="text-2xl font-bold text-foreground">学习路径规划</h1>
        <p className="text-sm text-muted-foreground mt-1.5">
          系统化的学习路线，帮助你高效掌握技术栈
        </p>
      </div>

      {/* 内容区 */}
      <div className="flex gap-8 px-8 py-8">
        {/* 左侧：路线图 */}
        <div className="flex-1 min-w-0">
          {/* 路径选择 */}
          <div className="grid grid-cols-3 gap-4 mb-8">
            {paths.map((path) => (
              <button
                key={path.id}
                onClick={() => setSelectedPath(path.id)}
                className={`p-4 rounded-lg border text-left transition-all ${
                  selectedPath === path.id
                    ? 'border-primary bg-primary/5 shadow-card'
                    : 'border-outline/10 bg-surface hover:shadow-card'
                }`}
              >
                <h3 className="text-sm font-medium text-foreground mb-1">
                  {path.name}
                </h3>
                <p className="text-xs text-muted-foreground">
                  {path.courses} 节课程 · {path.levels}
                </p>
              </button>
            ))}
          </div>

          {/* 进度概览 */}
          <div className="bg-surface rounded-lg shadow-card p-5 mb-6">
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-sm font-semibold text-foreground">当前进度</h3>
              <span className="text-sm text-primary font-medium">33%</span>
            </div>
            <div className="h-2 bg-surface-container rounded-full overflow-hidden">
              <div className="h-full bg-primary rounded-full" style={{ width: '33%' }} />
            </div>
            <p className="text-xs text-muted-foreground mt-2">
              已完成 2/6 阶段，共 30 课时
            </p>
          </div>

          {/* 时间轴 */}
          <div className="relative">
            {stages.map((stage, index) => (
              <div key={stage.id} className="relative pl-8 pb-8 last:pb-0">
                {/* 连接线 */}
                {index < stages.length - 1 && (
                  <div className="absolute left-[11px] top-6 bottom-0 w-0.5 bg-outline/20" />
                )}

                {/* 状态图标 */}
                <div className={`absolute left-0 top-0 w-6 h-6 rounded-full flex items-center justify-center ${
                  stage.status === 'completed'
                    ? 'bg-success text-on-primary'
                    : stage.status === 'in-progress'
                    ? 'bg-primary text-on-primary'
                    : 'bg-surface-container text-muted-foreground'
                }`}>
                  {stage.status === 'completed' && <Check className="w-3 h-3" />}
                  {stage.status === 'in-progress' && <Play className="w-3 h-3 fill-current" />}
                  {stage.status === 'locked' && <Lock className="w-3 h-3" />}
                </div>

                {/* 内容 */}
                <div className={`rounded-lg border p-4 ${
                  stage.status === 'in-progress'
                    ? 'border-primary bg-primary/5'
                    : 'border-outline/10 bg-surface'
                } ${stage.status === 'locked' ? 'opacity-60' : ''}`}>
                  <div className="flex items-center gap-2 mb-2">
                    <h4 className="text-sm font-semibold text-foreground">
                      阶段 {stage.id}：{stage.name}
                    </h4>
                    <span className="text-xs text-muted-foreground">
                      {stage.hours} 课时
                    </span>
                    {stage.status === 'in-progress' && (
                      <span className="text-xs text-primary font-medium">进行中</span>
                    )}
                  </div>
                  <p className="text-xs text-muted-foreground mb-3">
                    目标：{stage.goal}
                  </p>
                  <div className="flex flex-wrap gap-2 mb-3">
                    {stage.courses.map((course) => (
                      <Link
                        key={course}
                        href="/course/1"
                        className="text-xs text-muted-foreground bg-surface-container hover:bg-surface-container-high px-2 py-1 rounded-sm transition-colors"
                      >
                        {course}
                      </Link>
                    ))}
                  </div>
                  <p className="text-xs text-muted-foreground">
                    前置知识：{stage.prerequisite}
                  </p>
                  {stage.status === 'in-progress' && (
                    <Link
                      href="/course/3"
                      className="inline-flex items-center gap-1 mt-3 text-xs font-medium text-primary hover:underline"
                    >
                      继续学习 →
                    </Link>
                  )}
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* 右侧边栏 */}
        <div className="w-[280px] shrink-0">
          <div className="sticky top-4 space-y-5">
            {/* 推荐资源 */}
            <div className="bg-surface rounded-lg shadow-card p-5">
              <h3 className="text-sm font-semibold text-foreground mb-4">
                推荐资源
              </h3>
              <div className="space-y-4">
                <div>
                  <h4 className="text-xs font-medium text-foreground mb-2">书籍</h4>
                  <div className="space-y-1">
                    {resources.books.map((book) => (
                      <a key={book} href="#" className="block text-xs text-muted-foreground hover:text-primary">
                        {book}
                      </a>
                    ))}
                  </div>
                </div>
                <div>
                  <h4 className="text-xs font-medium text-foreground mb-2">视频课程</h4>
                  <div className="space-y-1">
                    {resources.videos.map((video) => (
                      <a key={video} href="#" className="block text-xs text-muted-foreground hover:text-primary">
                        {video}
                      </a>
                    ))}
                  </div>
                </div>
                <div>
                  <h4 className="text-xs font-medium text-foreground mb-2">练习平台</h4>
                  <div className="space-y-1">
                    {resources.platforms.map((platform) => (
                      <a key={platform} href="#" className="block text-xs text-muted-foreground hover:text-primary">
                        {platform}
                      </a>
                    ))}
                  </div>
                </div>
                <div>
                  <h4 className="text-xs font-medium text-foreground mb-2">文档</h4>
                  <div className="space-y-1">
                    {resources.docs.map((doc) => (
                      <a key={doc} href="#" className="block text-xs text-muted-foreground hover:text-primary">
                        {doc}
                      </a>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
