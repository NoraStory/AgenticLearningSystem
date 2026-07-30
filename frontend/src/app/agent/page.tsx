'use client';

import Link from 'next/link';

// 课程模块数据
const modules = [
  {
    id: 1,
    title: 'Prompt Engineering',
    level: '入门',
    courses: 4,
    summary: '掌握提示词工程的核心技巧，学会与 LLM 高效沟通，构建高质量的提示词。',
    tags: ['提示词', 'ChatGPT', '技巧'],
    image: 'https://images.unsplash.com/photo-1677442136019-21780ecad995?w=800&h=400&fit=crop',
  },
  {
    id: 2,
    title: 'LangChain 框架',
    level: '进阶',
    courses: 6,
    summary: '深入学习 LangChain 框架，掌握链式调用、记忆管理、工具集成等核心功能。',
    tags: ['LangChain', '链式调用', '工具'],
    image: 'https://images.unsplash.com/photo-1555949963-aa79dcee981c?w=800&h=400&fit=crop',
  },
  {
    id: 3,
    title: 'LangGraph 工作流',
    level: '进阶',
    courses: 5,
    summary: '学习使用 LangGraph 构建复杂的 Agent 工作流，实现多步骤任务和状态管理。',
    tags: ['LangGraph', '工作流', '状态管理'],
    image: 'https://images.unsplash.com/photo-1551288049-bebda4e38f71?w=800&h=400&fit=crop',
  },
  {
    id: 4,
    title: 'RAG 检索增强生成',
    level: '进阶',
    courses: 5,
    summary: '掌握 RAG 技术，让 LLM 能够访问外部知识库，生成更准确、更有依据的回答。',
    tags: ['RAG', '向量数据库', '知识库'],
    image: 'https://images.unsplash.com/photo-1516321318423-f06f85e504b3?w=800&h=400&fit=crop',
  },
  {
    id: 5,
    title: '多 Agent 协作',
    level: '高级',
    courses: 4,
    summary: '学习多 Agent 系统设计，掌握 Agent 之间的协作、通信和任务分配机制。',
    tags: ['多Agent', '协作', 'AutoGen'],
    image: 'https://images.unsplash.com/photo-1519389950473-47ba0277781c?w=800&h=400&fit=crop',
  },
  {
    id: 6,
    title: '工具调用与函数调用',
    level: '进阶',
    courses: 3,
    summary: '实现 LLM 与外部工具的交互，掌握 Function Calling 和 Tool Use 的实现方式。',
    tags: ['工具调用', 'Function Call', 'API'],
    image: 'https://images.unsplash.com/photo-1504639725590-34d0984388bd?w=800&h=400&fit=crop',
  },
  {
    id: 7,
    title: '记忆管理系统',
    level: '进阶',
    courses: 3,
    summary: '为 Agent 添加记忆能力，实现短期记忆、长期记忆和上下文管理。',
    tags: ['记忆', '上下文', '状态'],
    image: 'https://images.unsplash.com/photo-1551288049-bebda4e38f71?w=800&h=400&fit=crop',
  },
  {
    id: 8,
    title: 'Agent 部署与工程化',
    level: '高级',
    courses: 4,
    summary: '将 Agent 部署到生产环境，掌握性能优化、监控、日志和错误处理。',
    tags: ['部署', '工程化', '监控'],
    image: 'https://images.unsplash.com/photo-1460925895917-afdab827c52f?w=800&h=400&fit=crop',
  },
];

// 实战项目
const projects = [
  { title: '天气查询 Agent', level: '初级', tags: ['LangChain', 'API'] },
  { title: '新闻摘要 Agent', level: '中级', tags: ['RAG', '摘要'] },
  { title: '办公助手 Agent', level: '高级', tags: ['多Agent', '工具'] },
];

// 技术栈
const techStack = [
  'LangChain', 'LangGraph', 'AutoGen', 'Dify', 'ChromaDB',
  'FAISS', 'OpenAI API', 'HuggingFace', 'Pinecone', 'LlamaIndex',
];

const levelColors: Record<string, string> = {
  '入门': 'bg-green-500/10 text-green-600',
  '进阶': 'bg-orange-500/10 text-orange-600',
  '高级': 'bg-red-500/10 text-red-600',
};

export default function AgentPage() {
  return (
    <>
      {/* 页面标题 */}
      <div className="px-8 pt-8 pb-6 border-b border-outline/10">
        <h1 className="text-2xl font-bold text-foreground">AI Agent 开发学习</h1>
        <p className="text-sm text-muted-foreground mt-1.5">
          从 LLM 基础到 Agent 工程化，系统掌握 AI Agent 开发全流程
        </p>
      </div>

      {/* 内容区 */}
      <div className="flex gap-8 px-8 py-8">
        {/* 左侧：课程模块 */}
        <div className="flex-1 min-w-0">
          <h2 className="text-lg font-semibold text-foreground mb-6">
            课程模块
          </h2>

          {/* 模块卡片列表 */}
          <div className="space-y-6 mb-10">
            {modules.map((module) => (
              <Link
                key={module.id}
                href={`/course/${module.id}`}
                className="group block"
              >
                <article className="overflow-hidden rounded-lg border border-outline/10 hover:shadow-card transition-shadow">
                  <div className="flex">
                    <div className="w-48 shrink-0 overflow-hidden">
                      <img
                        src={module.image}
                        alt={module.title}
                        className="w-full h-full object-cover group-hover:scale-[1.02] transition-transform duration-300"
                      />
                    </div>
                    <div className="flex-1 p-5">
                      <div className="flex items-center gap-2 mb-2">
                        <span
                          className={`inline-flex items-center px-2 py-0.5 rounded-sm text-xs font-medium ${levelColors[module.level]}`}
                        >
                          {module.level}
                        </span>
                        <span className="text-xs text-muted-foreground">
                          {module.courses} 节课程
                        </span>
                      </div>
                      <h3 className="text-lg font-bold text-foreground group-hover:text-primary transition-colors mb-2">
                        {module.title}
                      </h3>
                      <p className="text-sm text-muted-foreground leading-relaxed mb-3">
                        {module.summary}
                      </p>
                      <div className="flex items-center gap-2">
                        {module.tags.map((tag) => (
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
            ))}
          </div>

          {/* 实战项目 */}
          <h2 className="text-lg font-semibold text-foreground mb-4">
            实战项目
          </h2>
          <div className="grid grid-cols-3 gap-4">
            {projects.map((project) => (
              <Link
                key={project.title}
                href="/practice"
                className="p-4 bg-surface rounded-lg border border-outline/10 hover:shadow-card transition-shadow group"
              >
                <h4 className="text-sm font-medium text-foreground group-hover:text-primary transition-colors mb-2">
                  {project.title}
                </h4>
                <span
                  className={`inline-flex items-center px-2 py-0.5 rounded-sm text-xs font-medium ${levelColors[project.level]} mb-2`}
                >
                  {project.level}
                </span>
                <div className="flex flex-wrap gap-1 mt-2">
                  {project.tags.map((tag) => (
                    <span
                      key={tag}
                      className="text-xs text-muted-foreground bg-surface-container px-1.5 py-0.5 rounded-sm"
                    >
                      {tag}
                    </span>
                  ))}
                </div>
              </Link>
            ))}
          </div>
        </div>

        {/* 右侧边栏 */}
        <div className="w-[280px] shrink-0">
          <div className="sticky top-4 space-y-5">
            {/* 学习进度 */}
            <div className="bg-surface rounded-lg shadow-card p-5">
              <h3 className="text-sm font-semibold text-foreground mb-3">
                学习进度
              </h3>
              <div className="text-2xl font-bold text-primary mb-1">18%</div>
              <p className="text-xs text-muted-foreground">
                已完成 1/8 模块
              </p>
              <div className="h-1.5 bg-surface-container rounded-full overflow-hidden mt-3">
                <div
                  className="h-full bg-primary rounded-full"
                  style={{ width: '18%' }}
                />
              </div>
            </div>

            {/* 技术栈 */}
            <div className="bg-surface rounded-lg shadow-card p-5">
              <h3 className="text-sm font-semibold text-foreground mb-4">
                技术栈
              </h3>
              <div className="flex flex-wrap gap-2">
                {techStack.map((tech) => (
                  <span
                    key={tech}
                    className="text-xs text-muted-foreground bg-surface-container hover:bg-surface-container-high px-2.5 py-1 rounded-sm cursor-pointer transition-colors"
                  >
                    {tech}
                  </span>
                ))}
              </div>
            </div>

            {/* 推荐资源 */}
            <div className="bg-surface rounded-lg shadow-card p-5">
              <h3 className="text-sm font-semibold text-foreground mb-4">
                推荐资源
              </h3>
              <div className="space-y-2">
                {[
                  'LangChain 官方文档',
                  'LangGraph 教程',
                  'OpenAI Cookbook',
                  'HuggingFace 课程',
                  'Agent 论文合集',
                ].map((name) => (
                  <a
                    key={name}
                    href="#"
                    className="block text-sm text-muted-foreground hover:text-primary transition-colors py-1"
                  >
                    {name}
                  </a>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
