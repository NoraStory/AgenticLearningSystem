'use client';

import Link from 'next/link';
import { useEffect, useState } from 'react';
import { apiFetch } from '@/lib/api';
import {
  Database,
  Table2,
  Layers,
  PenTool,
  Zap,
  Lock,
  Server,
  HardDrive,
  FileJson,
  ShieldCheck,
  BookOpen,
  Clock,
  CheckCircle2,
  Circle,
  PlayCircle,
  ExternalLink,
  TrendingUp,
  Award,
} from 'lucide-react';

// 课程章节数据
const fallbackChapters = [
  {
    id: 1,
    title: '数据库基础概念与分类',
    level: '入门',
    status: 'completed',
    summary:
      '了解数据库的核心概念，包括关系型数据库与非关系型数据库的区别、数据模型、DBMS 架构，以及常见的数据库产品对比。',
    tags: ['数据模型', 'RDBMS', 'NoSQL', 'ER 模型'],
    icon: Database,
    gradient: 'from-blue-500 to-cyan-500',
    lessons: 6,
    duration: '3 小时',
  },
  {
    id: 2,
    title: 'SQL 基础（增删改查）',
    level: '入门',
    status: 'completed',
    summary:
      '掌握 SQL 语言的基本语法，包括 SELECT、INSERT、UPDATE、DELETE 四大操作，以及 WHERE 条件过滤、ORDER BY 排序、LIMIT 分页等实用技巧。',
    tags: ['SQL', 'SELECT', 'CRUD', 'WHERE'],
    icon: Table2,
    gradient: 'from-emerald-500 to-teal-500',
    lessons: 8,
    duration: '5 小时',
  },
  {
    id: 3,
    title: '高级 SQL（子查询、连接、窗口函数）',
    level: '进阶',
    status: 'in-progress',
    summary:
      '深入学习 SQL 高级特性，包括 JOIN 多表连接、子查询与嵌套查询、窗口函数（ROW_NUMBER、RANK、LAG/LEAD）、CTE 公用表表达式等。',
    tags: ['JOIN', '子查询', '窗口函数', 'CTE'],
    icon: Layers,
    gradient: 'from-violet-500 to-purple-500',
    lessons: 10,
    duration: '8 小时',
  },
  {
    id: 4,
    title: '数据库设计（范式、ER 图）',
    level: '进阶',
    status: 'not-started',
    summary:
      '学习数据库设计的理论与实践，掌握三大范式（1NF/2NF/3NF）、BCNF、ER 图绘制方法，以及反范式化设计策略。',
    tags: ['范式', 'ER 图', '数据建模', '规范化'],
    icon: PenTool,
    gradient: 'from-orange-500 to-amber-500',
    lessons: 7,
    duration: '6 小时',
  },
  {
    id: 5,
    title: '索引与查询优化',
    level: '进阶',
    status: 'not-started',
    summary:
      '深入理解数据库索引的工作原理，包括 B+ 树索引、哈希索引、复合索引、覆盖索引，以及查询执行计划分析与 SQL 调优技巧。',
    tags: ['索引', 'B+ 树', '查询优化', '执行计划'],
    icon: Zap,
    gradient: 'from-rose-500 to-pink-500',
    lessons: 8,
    duration: '7 小时',
  },
  {
    id: 6,
    title: '事务与并发控制',
    level: '高级',
    status: 'not-started',
    summary:
      '掌握数据库事务的 ACID 特性、隔离级别（读未提交/读已提交/可重复读/串行化）、锁机制、MVCC 以及死锁处理策略。',
    tags: ['事务', 'ACID', '隔离级别', 'MVCC', '锁'],
    icon: Lock,
    gradient: 'from-indigo-500 to-blue-600',
    lessons: 6,
    duration: '5 小时',
  },
  {
    id: 7,
    title: 'PostgreSQL 实战',
    level: '进阶',
    status: 'not-started',
    summary:
      '通过实际项目学习 PostgreSQL 的高级特性，包括 JSON/JSONB 数据类型、全文搜索、分区表、扩展插件（PostGIS、pg_trgm）等。',
    tags: ['PostgreSQL', 'JSON', '全文搜索', '分区表'],
    icon: Server,
    gradient: 'from-cyan-500 to-teal-500',
    lessons: 9,
    duration: '8 小时',
  },
  {
    id: 8,
    title: 'Redis 缓存数据库',
    level: '进阶',
    status: 'not-started',
    summary:
      '学习 Redis 的核心数据结构与应用场景，包括 String/Hash/List/Set/ZSet、持久化策略（RDB/AOF）、缓存穿透/击穿/雪崩解决方案。',
    tags: ['Redis', '缓存', '数据结构', '持久化'],
    icon: HardDrive,
    gradient: 'from-red-500 to-orange-500',
    lessons: 8,
    duration: '6 小时',
  },
  {
    id: 9,
    title: 'MongoDB 文档数据库',
    level: '进阶',
    status: 'not-started',
    summary:
      '掌握 MongoDB 文档模型设计、CRUD 操作、聚合管道、索引策略，以及分片集群与副本集的高可用架构。',
    tags: ['MongoDB', '文档模型', '聚合管道', '分片'],
    icon: FileJson,
    gradient: 'from-green-500 to-emerald-600',
    lessons: 7,
    duration: '6 小时',
  },
  {
    id: 10,
    title: '数据库运维与备份',
    level: '高级',
    status: 'not-started',
    summary:
      '学习数据库生产环境运维技能，包括主从复制、读写分离、备份恢复策略、性能监控、慢查询分析、容量规划与故障排查。',
    tags: ['运维', '备份', '主从复制', '监控'],
    icon: ShieldCheck,
    gradient: 'from-slate-500 to-gray-600',
    lessons: 8,
    duration: '7 小时',
  },
];

// 推荐资源
const fallbackResources = [
  { name: 'PostgreSQL 官方文档', type: '文档', url: '#' },
  { name: 'SQLZoo 交互式教程', type: '练习', url: '#' },
  { name: 'Redis 官方教程', type: '文档', url: '#' },
  { name: 'MongoDB University', type: '课程', url: '#' },
  { name: 'Use The Index, Luke!', type: '书籍', url: '#' },
  { name: 'Database Internals', type: '书籍', url: '#' },
];

// 标签云
const fallbackTagCloud = [
  { name: 'SQL', count: 12 },
  { name: '索引', count: 8 },
  { name: '事务', count: 6 },
  { name: 'PostgreSQL', count: 9 },
  { name: 'Redis', count: 7 },
  { name: 'MongoDB', count: 5 },
  { name: '范式', count: 4 },
  { name: 'JOIN', count: 10 },
  { name: '备份', count: 3 },
  { name: '性能优化', count: 6 },
  { name: 'ACID', count: 4 },
  { name: 'NoSQL', count: 5 },
  { name: 'ER 图', count: 3 },
  { name: '窗口函数', count: 4 },
  { name: '缓存', count: 7 },
];

const levelColors: Record<string, string> = {
  入门: 'bg-green-500/10 text-green-600',
  进阶: 'bg-orange-500/10 text-orange-600',
  高级: 'bg-red-500/10 text-red-600',
};

const statusConfig: Record<
  string,
  { label: string; icon: typeof CheckCircle2; color: string }
> = {
  completed: {
    label: '已完成',
    icon: CheckCircle2,
    color: 'text-success',
  },
  'in-progress': {
    label: '学习中',
    icon: PlayCircle,
    color: 'text-primary',
  },
  'not-started': {
    label: '未开始',
    icon: Circle,
    color: 'text-muted-foreground',
  },
};

export default function DatabasePage() {
  const [levelFilter, setLevelFilter] = useState('all');
  const [statusFilter, setStatusFilter] = useState('all');
  const [chapters, setChapters] = useState(fallbackChapters);
  const [resources, setResources] = useState(fallbackResources);
  const [tagCloud, setTagCloud] = useState(fallbackTagCloud);

  useEffect(() => {
    Promise.all([
      apiFetch<{ items: Array<Record<string, unknown>> }>('/api/v1/courses?category=database'),
      apiFetch<{ items: typeof fallbackResources }>('/api/v1/courses/resources?category=database'),
      apiFetch<{ items: typeof fallbackTagCloud }>('/api/v1/courses/tags?category=database'),
    ]).then(([courseData, resourceData, tagData]) => {
      setChapters(courseData.items.map((course, index) => ({
        ...fallbackChapters[index % fallbackChapters.length], id: Number(course.id), title: String(course.title),
        level: String(course.level), summary: String(course.description), tags: course.tags as string[],
        status: String(course.status).replace('_', '-'), lessons: Number(course.lessons_count),
        duration: String(Number(course.estimated_hours)) + ' 小时',
      })));
      setResources(resourceData.items.length ? resourceData.items : fallbackResources);
      setTagCloud(tagData.items.length ? tagData.items : fallbackTagCloud);
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
  const progressPercent = Math.round(
    ((completedCount + inProgressCount * 0.4) / chapters.length) * 100
  );

  return (
    <>
      {/* 页面标题 */}
      <div className="px-8 pt-8 pb-6 border-b border-outline/10">
        <h1 className="text-2xl font-bold text-foreground">数据库学习</h1>
        <p className="text-sm text-muted-foreground mt-1.5">
          掌握关系型与非关系型数据库，从 SQL 到数据库设计
        </p>
        <div className="flex items-center gap-4 mt-3 text-sm">
          <span className="text-muted-foreground">
            共 <span className="font-medium text-foreground">{chapters.length}</span> 章节
          </span>
          <span className="text-muted-foreground">
            已完成 <span className="font-medium text-success">{completedCount}</span>
          </span>
          <span className="text-muted-foreground">
            学习中 <span className="font-medium text-primary">{inProgressCount}</span>
          </span>
          <span className="text-muted-foreground">
            总课时 <span className="font-medium text-foreground">
              {chapters.reduce((acc, c) => acc + c.lessons, 0)}
            </span>
          </span>
        </div>
      </div>

      {/* 内容区 */}
      <div className="flex gap-8 px-8 py-8">
        {/* 左侧：课程卡片列表 */}
        <div className="flex-1 min-w-0">
          {/* 筛选器 */}
          <div className="flex items-center gap-3 mb-6">
            <select
              value={levelFilter}
              onChange={(e) => setLevelFilter(e.target.value)}
              className="text-sm border border-outline rounded-md px-3 py-1.5 bg-surface text-foreground focus:outline-none focus:ring-2 focus:ring-ring/20"
            >
              <option value="all">全部难度</option>
              <option value="入门">入门</option>
              <option value="进阶">进阶</option>
              <option value="高级">高级</option>
            </select>
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="text-sm border border-outline rounded-md px-3 py-1.5 bg-surface text-foreground focus:outline-none focus:ring-2 focus:ring-ring/20"
            >
              <option value="all">全部状态</option>
              <option value="completed">已完成</option>
              <option value="in-progress">学习中</option>
              <option value="not-started">未开始</option>
            </select>
            <span className="text-xs text-muted-foreground ml-auto">
              显示 {filtered.length} / {chapters.length} 章节
            </span>
          </div>

          {/* 课程卡片列表 */}
          <div className="space-y-5">
            {filtered.map((chapter) => {
              const Icon = chapter.icon;
              const StatusIcon = statusConfig[chapter.status].icon;
              return (
                <Link
                  key={chapter.id}
                  href={`/course/${chapter.id}`}
                  className="group block"
                >
                  <article className="overflow-hidden rounded-lg border border-outline/10 bg-surface hover:shadow-card transition-shadow">
                    <div className="flex">
                      {/* 渐变色封面区域 */}
                      <div
                        className={`w-44 shrink-0 bg-gradient-to-br ${chapter.gradient} flex items-center justify-center relative overflow-hidden`}
                      >
                        {/* 装饰性背景图案 */}
                        <div className="absolute inset-0 opacity-10">
                          <div className="absolute top-3 left-3 w-16 h-16 border-2 border-white rounded-lg rotate-12" />
                          <div className="absolute bottom-4 right-4 w-12 h-12 border-2 border-white rounded-full" />
                          <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-24 h-24 border border-white rounded-full" />
                        </div>
                        <Icon className="w-12 h-12 text-white drop-shadow-md relative z-10" />
                        {/* 章节编号 */}
                        <div className="absolute top-3 left-3 text-xs font-bold text-white/70">
                          #{String(chapter.id).padStart(2, '0')}
                        </div>
                      </div>

                      {/* 内容区域 */}
                      <div className="flex-1 p-5">
                        <div className="flex items-center gap-2 mb-2">
                          <span
                            className={`inline-flex items-center px-2 py-0.5 rounded-sm text-xs font-medium ${levelColors[chapter.level]}`}
                          >
                            {chapter.level}
                          </span>
                          <span
                            className={`inline-flex items-center gap-1 text-xs ${statusConfig[chapter.status].color}`}
                          >
                            <StatusIcon className="w-3.5 h-3.5" />
                            {statusConfig[chapter.status].label}
                          </span>
                          <span className="text-xs text-muted-foreground ml-auto flex items-center gap-1">
                            <BookOpen className="w-3 h-3" />
                            {chapter.lessons} 课时
                          </span>
                          <span className="text-xs text-muted-foreground flex items-center gap-1">
                            <Clock className="w-3 h-3" />
                            {chapter.duration}
                          </span>
                        </div>

                        <h3 className="text-base font-bold text-foreground group-hover:text-primary transition-colors mb-2">
                          {chapter.title}
                        </h3>

                        <p className="text-sm text-muted-foreground leading-relaxed mb-3 line-clamp-2">
                          {chapter.summary}
                        </p>

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

          {filtered.length === 0 && (
            <div className="text-center py-16">
              <Database className="w-12 h-12 text-muted-foreground/30 mx-auto mb-3" />
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
              <div className="flex items-center gap-2 mb-3">
                <TrendingUp className="w-4 h-4 text-primary" />
                <h3 className="text-sm font-semibold text-foreground">
                  学习进度
                </h3>
              </div>
              <div className="text-2xl font-bold text-primary mb-1">
                {progressPercent}%
              </div>
              <p className="text-xs text-muted-foreground mb-3">
                已完成 {completedCount}/{chapters.length} 章节
              </p>
              <div className="h-2 bg-surface-container rounded-full overflow-hidden">
                <div
                  className="h-full bg-primary rounded-full transition-all"
                  style={{ width: `${progressPercent}%` }}
                />
              </div>
              <div className="grid grid-cols-3 gap-2 mt-4 pt-4 border-t border-outline/10">
                <div className="text-center">
                  <div className="text-lg font-bold text-success">{completedCount}</div>
                  <div className="text-xs text-muted-foreground">已完成</div>
                </div>
                <div className="text-center">
                  <div className="text-lg font-bold text-primary">{inProgressCount}</div>
                  <div className="text-xs text-muted-foreground">学习中</div>
                </div>
                <div className="text-center">
                  <div className="text-lg font-bold text-muted-foreground">
                    {chapters.length - completedCount - inProgressCount}
                  </div>
                  <div className="text-xs text-muted-foreground">未开始</div>
                </div>
              </div>
            </div>

            {/* 成就徽章 */}
            <div className="bg-surface rounded-lg shadow-card p-5">
              <div className="flex items-center gap-2 mb-3">
                <Award className="w-4 h-4 text-warning" />
                <h3 className="text-sm font-semibold text-foreground">
                  学习成就
                </h3>
              </div>
              <div className="space-y-3">
                <div className="flex items-center gap-3">
                  <div className="w-8 h-8 rounded-full bg-green-500/10 flex items-center justify-center">
                    <CheckCircle2 className="w-4 h-4 text-success" />
                  </div>
                  <div>
                    <p className="text-xs font-medium text-foreground">SQL 新手</p>
                    <p className="text-xs text-muted-foreground">完成 SQL 基础章节</p>
                  </div>
                </div>
                <div className="flex items-center gap-3">
                  <div className="w-8 h-8 rounded-full bg-green-500/10 flex items-center justify-center">
                    <CheckCircle2 className="w-4 h-4 text-success" />
                  </div>
                  <div>
                    <p className="text-xs font-medium text-foreground">数据库入门</p>
                    <p className="text-xs text-muted-foreground">完成基础概念章节</p>
                  </div>
                </div>
                <div className="flex items-center gap-3 opacity-50">
                  <div className="w-8 h-8 rounded-full bg-surface-container flex items-center justify-center">
                    <Circle className="w-4 h-4 text-muted-foreground" />
                  </div>
                  <div>
                    <p className="text-xs font-medium text-muted-foreground">高级 SQL 大师</p>
                    <p className="text-xs text-muted-foreground">完成高级 SQL 章节</p>
                  </div>
                </div>
                <div className="flex items-center gap-3 opacity-50">
                  <div className="w-8 h-8 rounded-full bg-surface-container flex items-center justify-center">
                    <Circle className="w-4 h-4 text-muted-foreground" />
                  </div>
                  <div>
                    <p className="text-xs font-medium text-muted-foreground">全能 DBA</p>
                    <p className="text-xs text-muted-foreground">完成所有章节</p>
                  </div>
                </div>
              </div>
            </div>

            {/* 推荐资源 */}
            <div className="bg-surface rounded-lg shadow-card p-5">
              <h3 className="text-sm font-semibold text-foreground mb-4">
                推荐资源
              </h3>
              <div className="space-y-2">
                {resources.map((resource) => (
                  <a
                    key={resource.name}
                    href={resource.url}
                    className="flex items-center justify-between group py-1.5"
                  >
                    <div className="flex items-center gap-2">
                      <ExternalLink className="w-3 h-3 text-muted-foreground group-hover:text-primary transition-colors" />
                      <span className="text-xs text-muted-foreground group-hover:text-primary transition-colors">
                        {resource.name}
                      </span>
                    </div>
                    <span className="text-xs text-muted-foreground bg-surface-container px-1.5 py-0.5 rounded-sm">
                      {resource.type}
                    </span>
                  </a>
                ))}
              </div>
            </div>

            {/* 标签云 */}
            <div className="bg-surface rounded-lg shadow-card p-5">
              <h3 className="text-sm font-semibold text-foreground mb-4">
                标签云
              </h3>
              <div className="flex flex-wrap gap-2">
                {tagCloud.map((tag) => (
                  <span
                    key={tag.name}
                    className="text-xs text-muted-foreground bg-surface-container hover:bg-surface-container-high px-2.5 py-1 rounded-sm cursor-pointer transition-colors"
                    style={{
                      fontSize: `${Math.min(Math.max(0.65 + tag.count * 0.04, 0.65), 0.95)}rem`,
                    }}
                  >
                    {tag.name}
                  </span>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
