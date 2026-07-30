'use client';

import Link from 'next/link';
import { useEffect, useState } from 'react';
import { apiFetch } from '@/lib/api';
import {
  BookOpen,
  Clock,
  Users,
  CheckCircle2,
  PlayCircle,
  Lock,
  Code2,
  GitBranch,
  Layers,
  FileText,
  Settings,
  Package,
  Globe,
  Database,
} from 'lucide-react';

const fallbackChapters = [
  {
    id: 1,
    title: 'Python 环境搭建与基础语法',
    level: '入门',
    status: 'completed',
    lessons: 4,
    hours: 6,
    tags: ['环境配置', '解释器', 'REPL'],
    desc: '安装 Python 解释器，配置开发环境（VS Code / PyCharm），学习交互式解释器和脚本执行方式，编写第一个 Python 程序。',
    icon: Settings,
    gradient: 'from-blue-500 to-cyan-500',
  },
  {
    id: 2,
    title: '变量、数据类型与运算符',
    level: '入门',
    status: 'completed',
    lessons: 5,
    hours: 8,
    tags: ['变量', '类型', '运算符'],
    desc: '理解 Python 的动态类型系统，掌握整数、浮点数、字符串、布尔值等基本数据类型，学习算术、比较、逻辑运算符。',
    icon: Code2,
    gradient: 'from-green-500 to-emerald-500',
  },
  {
    id: 3,
    title: '控制流：条件与循环',
    level: '入门',
    status: 'completed',
    lessons: 4,
    hours: 7,
    tags: ['if', 'for', 'while'],
    desc: '学习 if/elif/else 条件判断、for 循环和 while 循环、break/continue 控制、列表推导式等 Python 特有的控制流语法。',
    icon: GitBranch,
    gradient: 'from-purple-500 to-violet-500',
  },
  {
    id: 4,
    title: '函数与模块',
    level: '进阶',
    status: 'current',
    progress: 60,
    lessons: 5,
    hours: 10,
    tags: ['函数', '参数', '模块', '包'],
    desc: '掌握函数定义与调用、参数传递（位置/关键字/默认/可变参数）、返回值、作用域、lambda 表达式、模块导入与自定义模块。',
    icon: Layers,
    gradient: 'from-orange-500 to-amber-500',
  },
  {
    id: 5,
    title: '面向对象编程',
    level: '进阶',
    status: 'locked',
    lessons: 6,
    hours: 12,
    tags: ['类', '继承', '多态', '封装'],
    desc: '学习类的定义与实例化、属性和方法、构造函数、继承与方法重写、多态、特殊方法（__str__、__repr__等）、装饰器。',
    icon: FileText,
    gradient: 'from-rose-500 to-pink-500',
  },
  {
    id: 6,
    title: '文件操作与异常处理',
    level: '进阶',
    status: 'locked',
    lessons: 4,
    hours: 8,
    tags: ['文件', '异常', 'try-except'],
    desc: '掌握文件读写操作、上下文管理器（with 语句）、异常捕获与处理、自定义异常、日志记录基础。',
    icon: FileText,
    gradient: 'from-indigo-500 to-blue-500',
  },
  {
    id: 7,
    title: '标准库与第三方库',
    level: '高级',
    status: 'locked',
    lessons: 5,
    hours: 10,
    tags: ['os', 'sys', 'requests', 'pip'],
    desc: '探索 Python 丰富的标准库（os、sys、json、datetime、re 等），学习 pip 包管理和常用第三方库（requests、numpy、pandas）。',
    icon: Package,
    gradient: 'from-cyan-500 to-teal-500',
  },
  {
    id: 8,
    title: 'Web 开发入门（Flask/Django）',
    level: '高级',
    status: 'locked',
    lessons: 6,
    hours: 15,
    tags: ['Flask', 'Django', 'REST', 'API'],
    desc: '使用 Flask 构建轻量级 Web 应用，了解 Django 全栈框架，学习 RESTful API 设计、路由、模板、数据库集成。',
    icon: Globe,
    gradient: 'from-teal-500 to-green-500',
  },
];

const fallbackResources = [
  { name: 'Python 官方文档', url: 'https://docs.python.org/zh-cn/3/' },
  { name: 'Real Python', url: 'https://realpython.com/' },
  { name: 'Python Cookbook', url: '#' },
  { name: 'LeetCode Python', url: '#' },
  { name: 'Python 教程（廖雪峰）', url: 'https://www.liaoxuefeng.com/wiki/1016959663602400' },
];

const fallbackTagCloud = [
  '变量', '函数', '类', '模块', '装饰器', '生成器',
  '迭代器', '异常', '文件', '正则', '多线程', '协程',
  'Flask', 'Django', 'numpy', 'pandas',
];

const levelColor: Record<string, string> = {
  '入门': 'bg-green-500/10 text-green-600',
  '进阶': 'bg-orange-500/10 text-orange-600',
  '高级': 'bg-red-500/10 text-red-600',
};

export default function PythonPage() {
  const [levelFilter, setLevelFilter] = useState('全部');
  const [statusFilter, setStatusFilter] = useState('全部');
  const [chapters, setChapters] = useState(fallbackChapters);
  const [resources, setResources] = useState(fallbackResources);
  const [tagCloud, setTagCloud] = useState(fallbackTagCloud);

  useEffect(() => {
    Promise.all([
      apiFetch<{ items: Array<Record<string, unknown>> }>('/api/v1/courses?category=python'),
      apiFetch<{ items: Array<{ name: string; url: string }> }>('/api/v1/courses/resources?category=python'),
      apiFetch<{ items: Array<{ name: string }> }>('/api/v1/courses/tags?category=python'),
    ]).then(([courseData, resourceData, tagData]) => {
      setChapters(courseData.items.map((course, index) => ({
        ...fallbackChapters[index % fallbackChapters.length],
        id: Number(course.id), title: String(course.title), level: String(course.level),
        status: String(course.status).replace('_', '-'), lessons: Number(course.lessons_count),
        hours: Number(course.estimated_hours), tags: course.tags as string[], desc: String(course.description),
      })));
      setResources(resourceData.items.length ? resourceData.items : fallbackResources);
      setTagCloud(tagData.items.length ? tagData.items.map((tag) => tag.name) : fallbackTagCloud);
    }).catch(() => undefined);
  }, []);

  const filtered = chapters.filter((ch) => {
    if (levelFilter !== '全部' && ch.level !== levelFilter) return false;
    if (statusFilter !== '全部' && ch.status !== statusFilter.toLowerCase()) return false;
    return true;
  });

  const completedCount = chapters.filter((c) => c.status === 'completed').length;

  return (
    <div className="flex-1 min-w-0 overflow-y-auto">
      <div className="max-w-5xl mx-auto px-6 py-8">
        {/* 页面标题 */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-foreground mb-2">Python 编程学习</h1>
          <p className="text-muted-foreground text-base">从入门到精通，掌握 Python 核心语法与实战应用</p>
          <div className="flex gap-4 mt-3 text-xs text-muted-foreground">
            <span>{chapters.length} 章课程</span>
            <span>{completedCount} 章已完成</span>
            <span>{chapters.reduce((s, c) => s + c.lessons, 0)} 节子课程</span>
          </div>
        </div>

        {/* 筛选器 */}
        <div className="flex items-center gap-3 mb-6">
          <select
            value={levelFilter}
            onChange={(e) => setLevelFilter(e.target.value)}
            className="px-3 py-1.5 text-sm bg-surface-container-lowest border border-outline/20 rounded-md text-foreground"
          >
            <option>全部</option>
            <option>入门</option>
            <option>进阶</option>
            <option>高级</option>
          </select>
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="px-3 py-1.5 text-sm bg-surface-container-lowest border border-outline/20 rounded-md text-foreground"
          >
            <option>全部</option>
            <option>已完成</option>
            <option>学习中</option>
            <option>未开始</option>
          </select>
          <span className="text-xs text-muted-foreground">显示 {filtered.length} 个章节</span>
        </div>

        <div className="flex gap-6">
          {/* 课程列表 */}
          <div className="flex-1 space-y-4">
            {filtered.map((ch) => {
              const Icon = ch.icon;
              return (
                <Link
                  key={ch.id}
                  href={`/course/python-${ch.id}`}
                  className="block bg-surface rounded-lg border border-outline/15 overflow-hidden hover:shadow-md transition-shadow"
                >
                  <div className={`h-28 bg-gradient-to-br ${ch.gradient} flex items-center justify-center relative`}>
                    <Icon className="w-12 h-12 text-white/80" />
                    <span className="absolute top-3 left-3 text-xs font-mono text-white/60">
                      #{String(ch.id).padStart(2, '0')}
                    </span>
                    {ch.status === 'completed' && (
                      <span className="absolute top-3 right-3 bg-white/20 text-white text-xs px-2 py-0.5 rounded-full">
                        已完成
                      </span>
                    )}
                    {ch.status === 'current' && (
                      <span className="absolute top-3 right-3 bg-white/30 text-white text-xs px-2 py-0.5 rounded-full">
                        学习中 {ch.progress}%
                      </span>
                    )}
                  </div>
                  <div className="p-4">
                    <div className="flex items-center gap-2 mb-2">
                      <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${levelColor[ch.level]}`}>
                        {ch.level}
                      </span>
                      <span className="text-xs text-muted-foreground">{ch.lessons} 节 · {ch.hours}h</span>
                    </div>
                    <h2 className="text-lg font-semibold text-foreground mb-1.5 hover:text-primary transition-colors">
                      {ch.title}
                    </h2>
                    <p className="text-sm text-muted-foreground line-clamp-2 mb-3">{ch.desc}</p>
                    <div className="flex flex-wrap gap-1.5">
                      {ch.tags.map((tag) => (
                        <span
                          key={tag}
                          className="text-xs px-2 py-0.5 bg-surface-container rounded text-muted-foreground"
                        >
                          {tag}
                        </span>
                      ))}
                    </div>
                  </div>
                </Link>
              );
            })}
            {filtered.length === 0 && (
              <div className="text-center py-12 text-muted-foreground text-sm">没有匹配的课程</div>
            )}
          </div>

          {/* 右侧边栏 */}
          <div className="w-64 shrink-0 space-y-4">
            <div className="bg-surface rounded-lg border border-outline/15 p-4">
              <h3 className="text-sm font-medium text-foreground mb-3">学习进度</h3>
              <div className="text-2xl font-bold text-primary mb-1">{Math.round((completedCount / chapters.length) * 100)}%</div>
              <div className="w-full h-2 bg-surface-container rounded-full overflow-hidden">
                <div
                  className="h-full bg-primary rounded-full"
                  style={{ width: `${(completedCount / chapters.length) * 100}%` }}
                />
              </div>
              <div className="text-xs text-muted-foreground mt-1">{completedCount}/{chapters.length} 章完成</div>
            </div>

            <div className="bg-surface rounded-lg border border-outline/15 p-4">
              <h3 className="text-sm font-medium text-foreground mb-3">推荐资源</h3>
              <ul className="space-y-2">
                {resources.map((r) => (
                  <li key={r.name}>
                    <a href={r.url} target="_blank" rel="noopener noreferrer" className="text-sm text-primary hover:underline">
                      {r.name}
                    </a>
                  </li>
                ))}
              </ul>
            </div>

            <div className="bg-surface rounded-lg border border-outline/15 p-4">
              <h3 className="text-sm font-medium text-foreground mb-3">知识标签</h3>
              <div className="flex flex-wrap gap-1.5">
                {tagCloud.map((tag) => (
                  <span
                    key={tag}
                    className="text-xs px-2 py-0.5 bg-surface-container rounded text-muted-foreground hover:text-primary hover:bg-primary/10 cursor-pointer transition-colors"
                  >
                    {tag}
                  </span>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
