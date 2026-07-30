'use client';

import Link from 'next/link';
import { useEffect, useState } from 'react';
import { apiFetch } from '@/lib/api';
import {
  Code2,
  Binary,
  GitBranch,
  Pointer,
  Container,
  Layers,
  Blocks,
  Cpu,
  Network,
  BookOpen,
  CheckCircle2,
  Circle,
  Clock,
} from 'lucide-react';

// 课程章节数据
const fallbackChapters = [
  {
    id: 1,
    title: 'C++ 开发环境与第一个程序',
    level: '入门',
    summary:
      '搭建 C++ 开发环境，安装编译器（GCC/Clang/MSVC），配置 IDE（VS Code / CLion），编写并运行你的第一个 Hello World 程序，理解编译链接过程。',
    tags: ['环境搭建', '编译器', 'Hello World'],
    icon: Code2,
    gradient: 'from-blue-500 to-cyan-500',
    status: 'completed',
    lessons: 4,
    duration: '3 小时',
  },
  {
    id: 2,
    title: '变量、数据类型与运算符',
    level: '入门',
    summary:
      '深入理解 C++ 基本数据类型（int、float、double、char、bool），掌握变量声明与初始化、类型转换、算术运算符、关系运算符和逻辑运算符的使用。',
    tags: ['数据类型', '变量', '运算符'],
    icon: Binary,
    gradient: 'from-emerald-500 to-teal-500',
    status: 'completed',
    lessons: 5,
    duration: '4 小时',
  },
  {
    id: 3,
    title: '控制流与函数',
    level: '入门',
    summary:
      '学习 if/else、switch、for、while 等控制流语句，掌握函数的定义与调用、参数传递方式（值传递、引用传递）、函数重载和默认参数。',
    tags: ['条件语句', '循环', '函数重载'],
    icon: GitBranch,
    gradient: 'from-violet-500 to-purple-500',
    status: 'completed',
    lessons: 6,
    duration: '5 小时',
  },
  {
    id: 4,
    title: '指针与引用',
    level: '进阶',
    summary:
      'C++ 的核心特性。理解指针的本质（内存地址的抽象），掌握指针运算、动态内存分配（new/delete）、引用的概念与使用场景，以及指针与数组的关系。',
    tags: ['指针', '引用', '内存地址', 'new/delete'],
    icon: Pointer,
    gradient: 'from-amber-500 to-orange-500',
    status: 'in-progress',
    lessons: 7,
    duration: '6 小时',
  },
  {
    id: 5,
    title: '数组、字符串与 STL 容器',
    level: '进阶',
    summary:
      '从原生数组到 std::string，再到 STL 标准模板库。深入学习 vector、list、map、set、unordered_map 等常用容器的使用方法和性能特点。',
    tags: ['数组', 'string', 'STL', 'vector', 'map'],
    icon: Container,
    gradient: 'from-pink-500 to-rose-500',
    status: 'not-started',
    lessons: 8,
    duration: '7 小时',
  },
  {
    id: 6,
    title: '面向对象编程（类、继承、多态）',
    level: '进阶',
    summary:
      '掌握 C++ 面向对象三大特性：封装（class/struct）、继承（public/protected/private）、多态（虚函数、纯虚函数、抽象类）。理解构造/析构函数和拷贝语义。',
    tags: ['类', '继承', '多态', '虚函数'],
    icon: Layers,
    gradient: 'from-indigo-500 to-blue-500',
    status: 'not-started',
    lessons: 8,
    duration: '8 小时',
  },
  {
    id: 7,
    title: '模板与泛型编程',
    level: '高级',
    summary:
      '学习函数模板、类模板的定义与使用，理解模板特化与偏特化，掌握可变参数模板（Variadic Templates）和 SFINAE 原则，了解 C++20 Concepts。',
    tags: ['模板', '泛型', '特化', 'Concepts'],
    icon: Blocks,
    gradient: 'from-cyan-500 to-blue-600',
    status: 'not-started',
    lessons: 6,
    duration: '6 小时',
  },
  {
    id: 8,
    title: '内存管理与智能指针',
    level: '高级',
    summary:
      '深入理解 C++ 内存模型（栈、堆、全局区），掌握 RAII 原则，学习智能指针（unique_ptr、shared_ptr、weak_ptr）的使用，避免内存泄漏和悬空指针。',
    tags: ['内存管理', '智能指针', 'RAII', 'unique_ptr'],
    icon: Cpu,
    gradient: 'from-red-500 to-orange-600',
    status: 'not-started',
    lessons: 5,
    duration: '5 小时',
  },
  {
    id: 9,
    title: '多线程与并发',
    level: '高级',
    summary:
      '学习 C++11 多线程支持库，掌握 std::thread、std::mutex、std::condition_variable 的使用，理解原子操作、内存序和现代 C++ 并发编程最佳实践。',
    tags: ['多线程', 'mutex', 'atomic', '并发'],
    icon: Network,
    gradient: 'from-fuchsia-500 to-pink-600',
    status: 'not-started',
    lessons: 6,
    duration: '6 小时',
  },
];

// 推荐资源
const fallbackResources = [
  { name: 'C++ Primer (第5版)', type: '书籍', url: '#' },
  { name: 'cppreference.com', type: '文档', url: '#' },
  { name: 'C++ Core Guidelines', type: '规范', url: '#' },
  { name: 'Effective Modern C++', type: '书籍', url: '#' },
  { name: 'The C++ Programming Language', type: '书籍', url: '#' },
  { name: 'Compiler Explorer (Godbolt)', type: '工具', url: '#' },
];

// 标签云
const fallbackTagCloud = [
  '指针', '内存管理', 'STL', '模板', 'OOP', '多态',
  '智能指针', 'RAII', 'move语义', 'lambda', '并发',
  '容器', '迭代器', '虚函数', '命名空间', '异常处理',
];

// 难度配色
const levelConfig: Record<string, { bg: string; text: string; label: string }> = {
  '入门': { bg: 'bg-emerald-500/10', text: 'text-emerald-600', label: '入门' },
  '进阶': { bg: 'bg-amber-500/10', text: 'text-amber-600', label: '进阶' },
  '高级': { bg: 'bg-red-500/10', text: 'text-red-600', label: '高级' },
};

// 状态配置
const statusConfig: Record<string, { icon: typeof CheckCircle2; color: string; label: string }> = {
  completed: { icon: CheckCircle2, color: 'text-success', label: '已完成' },
  'in-progress': { icon: Clock, color: 'text-primary', label: '学习中' },
  'not-started': { icon: Circle, color: 'text-muted-foreground', label: '未开始' },
};

export default function CppPage() {
  const [levelFilter, setLevelFilter] = useState('all');
  const [statusFilter, setStatusFilter] = useState('all');
  const [chapters, setChapters] = useState(fallbackChapters);
  const [resources, setResources] = useState(fallbackResources);
  const [tagCloud, setTagCloud] = useState(fallbackTagCloud);

  useEffect(() => {
    Promise.all([
      apiFetch<{ items: Array<Record<string, unknown>> }>('/api/v1/courses?category=cpp'),
      apiFetch<{ items: typeof fallbackResources }>('/api/v1/courses/resources?category=cpp'),
      apiFetch<{ items: Array<{ name: string }> }>('/api/v1/courses/tags?category=cpp'),
    ]).then(([courseData, resourceData, tagData]) => {
      setChapters(courseData.items.map((course, index) => ({
        ...fallbackChapters[index % fallbackChapters.length], id: Number(course.id), title: String(course.title),
        level: String(course.level), summary: String(course.description), tags: course.tags as string[],
        status: String(course.status).replace('_', '-'), lessons: Number(course.lessons_count),
        duration: String(Number(course.estimated_hours)) + ' 小时',
      })));
      setResources(resourceData.items.length ? resourceData.items : fallbackResources);
      setTagCloud(tagData.items.length ? tagData.items.map((tag) => tag.name) : fallbackTagCloud);
    }).catch(() => undefined);
  }, []);

  const filtered = chapters.filter((ch) => {
    const matchLevel = levelFilter === 'all' || ch.level === levelFilter;
    const matchStatus = statusFilter === 'all' || ch.status === statusFilter;
    return matchLevel && matchStatus;
  });

  // 统计数据
  const completedCount = chapters.filter((c) => c.status === 'completed').length;
  const inProgressCount = chapters.filter((c) => c.status === 'in-progress').length;
  const progressPercent = Math.round((completedCount / chapters.length) * 100);

  return (
    <>
      {/* 页面标题 */}
      <div className="px-8 pt-8 pb-6 border-b border-outline/10">
        <h1 className="text-2xl font-bold text-foreground">C++ 编程学习</h1>
        <p className="text-sm text-muted-foreground mt-1.5">
          系统级编程利器，从基础语法到高性能开发
        </p>
        <div className="flex items-center gap-4 mt-3 text-sm">
          <span className="text-muted-foreground">
            总章节 <span className="font-medium text-foreground">{chapters.length}</span>
          </span>
          <span className="text-muted-foreground">
            已完成 <span className="font-medium text-success">{completedCount}</span>
          </span>
          <span className="text-muted-foreground">
            学习中 <span className="font-medium text-primary">{inProgressCount}</span>
          </span>
          <span className="text-muted-foreground">
            总课时 <span className="font-medium text-foreground">
              {chapters.reduce((sum, c) => sum + c.lessons, 0)}
            </span>
          </span>
        </div>
      </div>

      {/* 内容区 */}
      <div className="flex gap-8 px-8 py-8">
        {/* 左侧：课程卡片列表 */}
        <div className="flex-1 min-w-0">
          {/* 筛选器 */}
          <div className="flex items-center gap-4 mb-6">
            {/* 难度筛选 */}
            <div className="flex items-center gap-1">
              <span className="text-xs text-muted-foreground mr-2">难度：</span>
              {[
                { key: 'all', label: '全部' },
                { key: '入门', label: '入门' },
                { key: '进阶', label: '进阶' },
                { key: '高级', label: '高级' },
              ].map((tab) => (
                <button
                  key={tab.key}
                  onClick={() => setLevelFilter(tab.key)}
                  className={`px-3 py-1.5 text-xs rounded-md transition-colors ${
                    levelFilter === tab.key
                      ? 'font-medium text-primary bg-primary/10'
                      : 'text-muted-foreground hover:text-foreground hover:bg-surface-container'
                  }`}
                >
                  {tab.label}
                </button>
              ))}
            </div>

            {/* 分隔线 */}
            <div className="h-4 w-px bg-border" />

            {/* 状态筛选 */}
            <div className="flex items-center gap-1">
              <span className="text-xs text-muted-foreground mr-2">状态：</span>
              {[
                { key: 'all', label: '全部' },
                { key: 'completed', label: '已完成' },
                { key: 'in-progress', label: '学习中' },
                { key: 'not-started', label: '未开始' },
              ].map((tab) => (
                <button
                  key={tab.key}
                  onClick={() => setStatusFilter(tab.key)}
                  className={`px-3 py-1.5 text-xs rounded-md transition-colors ${
                    statusFilter === tab.key
                      ? 'font-medium text-primary bg-primary/10'
                      : 'text-muted-foreground hover:text-foreground hover:bg-surface-container'
                  }`}
                >
                  {tab.label}
                </button>
              ))}
            </div>

            <span className="text-xs text-muted-foreground ml-auto">
              共 {filtered.length} 章
            </span>
          </div>

          {/* 课程卡片列表 */}
          <div className="space-y-5">
            {filtered.map((chapter) => {
              const IconComponent = chapter.icon;
              const level = levelConfig[chapter.level];
              const status = statusConfig[chapter.status];
              const StatusIcon = status.icon;

              return (
                <Link
                  key={chapter.id}
                  href={`/course/${chapter.id}`}
                  className="group block"
                >
                  <article className="overflow-hidden rounded-lg border border-outline/10 hover:shadow-card transition-all duration-200">
                    <div className="flex">
                      {/* 封面图区域 - 渐变色背景 + 图标 */}
                      <div className={`w-40 shrink-0 bg-gradient-to-br ${chapter.gradient} flex items-center justify-center`}>
                        <IconComponent className="w-12 h-12 text-white/90" strokeWidth={1.5} />
                      </div>

                      {/* 内容区域 */}
                      <div className="flex-1 p-5">
                        {/* 顶部：分类标签 + 状态 */}
                        <div className="flex items-center justify-between mb-2">
                          <div className="flex items-center gap-2">
                            <span className={`inline-flex items-center px-2 py-0.5 rounded-sm text-xs font-medium ${level.bg} ${level.text}`}>
                              {level.label}
                            </span>
                            <span className="text-xs text-muted-foreground">
                              {chapter.lessons} 课时 · {chapter.duration}
                            </span>
                          </div>
                          <div className={`flex items-center gap-1 text-xs ${status.color}`}>
                            <StatusIcon className="w-3.5 h-3.5" />
                            <span>{status.label}</span>
                          </div>
                        </div>

                        {/* 章节标题 */}
                        <h3 className="text-base font-bold text-foreground group-hover:text-primary transition-colors mb-2">
                          第{chapter.id}章：{chapter.title}
                        </h3>

                        {/* 摘要描述 */}
                        <p className="text-sm text-muted-foreground leading-relaxed mb-3 line-clamp-2">
                          {chapter.summary}
                        </p>

                        {/* 标签 */}
                        <div className="flex items-center gap-2 flex-wrap">
                          {chapter.tags.map((tag) => (
                            <span
                              key={tag}
                              className="text-xs text-muted-foreground bg-surface-container px-2 py-0.5 rounded-sm"
                            >
                              {tag}
                            </span>
                          ))}
                        </div>
                      </div>
                    </div>
                  </article>
                </Link>
              );
            })}
          </div>

          {/* 空状态提示 */}
          {filtered.length === 0 && (
            <div className="text-center py-16">
              <p className="text-sm text-muted-foreground">
                没有符合条件的章节，请调整筛选条件
              </p>
            </div>
          )}
        </div>

        {/* 右侧边栏 */}
        <div className="w-[280px] shrink-0">
          <div className="sticky top-4 space-y-5">
            {/* 学习进度 */}
            <div className="bg-surface rounded-lg shadow-card p-5">
              <h3 className="text-sm font-semibold text-foreground mb-3">
                学习进度
              </h3>
              <div className="text-2xl font-bold text-primary mb-1">{progressPercent}%</div>
              <p className="text-xs text-muted-foreground">
                已完成 {completedCount}/{chapters.length} 章节
              </p>
              <div className="h-1.5 bg-surface-container rounded-full overflow-hidden mt-3">
                <div
                  className="h-full bg-primary rounded-full transition-all"
                  style={{ width: `${progressPercent}%` }}
                />
              </div>

              {/* 各难度进度 */}
              <div className="space-y-3 mt-5">
                {[
                  { label: '入门', done: 3, total: 3, color: 'bg-emerald-500' },
                  { label: '进阶', done: 0, total: 3, color: 'bg-amber-500' },
                  { label: '高级', done: 0, total: 3, color: 'bg-red-500' },
                ].map((item) => (
                  <div key={item.label}>
                    <div className="flex justify-between text-xs mb-1">
                      <span className="text-foreground">{item.label}</span>
                      <span className="text-muted-foreground">
                        {item.done}/{item.total}
                      </span>
                    </div>
                    <div className="h-1.5 bg-surface-container rounded-full overflow-hidden">
                      <div
                        className={`h-full ${item.color} rounded-full`}
                        style={{ width: `${(item.done / item.total) * 100}%` }}
                      />
                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* 推荐资源 */}
            <div className="bg-surface rounded-lg shadow-card p-5">
              <h3 className="text-sm font-semibold text-foreground mb-4">
                推荐资源
              </h3>
              <div className="space-y-2.5">
                {resources.map((res) => (
                  <a
                    key={res.name}
                    href={res.url}
                    className="flex items-center gap-2 text-sm text-muted-foreground hover:text-primary transition-colors py-0.5 group"
                  >
                    <BookOpen className="w-3.5 h-3.5 shrink-0 group-hover:text-primary" />
                    <span className="flex-1 truncate">{res.name}</span>
                    <span className="text-[10px] bg-surface-container px-1.5 py-0.5 rounded-sm text-muted-foreground shrink-0">
                      {res.type}
                    </span>
                  </a>
                ))}
              </div>
            </div>

            {/* 标签云 */}
            <div className="bg-surface rounded-lg shadow-card p-5">
              <h3 className="text-sm font-semibold text-foreground mb-4">
                知识标签
              </h3>
              <div className="flex flex-wrap gap-2">
                {tagCloud.map((tag) => (
                  <span
                    key={tag}
                    className="text-xs text-muted-foreground bg-surface-container hover:bg-surface-container-high px-2.5 py-1 rounded-sm cursor-pointer transition-colors"
                  >
                    {tag}
                  </span>
                ))}
              </div>
            </div>

            {/* 学习提示 */}
            <div className="bg-surface rounded-lg shadow-card p-5 border-l-4 border-primary">
              <div className="flex items-center gap-2 mb-2">
                <BookOpen className="w-4 h-4 text-primary" />
                <h3 className="text-sm font-semibold text-foreground">
                  学习建议
                </h3>
              </div>
              <p className="text-xs text-muted-foreground leading-relaxed">
                C++ 学习曲线较陡，建议先打好基础（前3章），再逐步深入指针、OOP 和模板。多动手写代码，善用 Compiler Explorer 观察编译输出。
              </p>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
