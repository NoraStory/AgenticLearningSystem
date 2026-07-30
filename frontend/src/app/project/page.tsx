'use client';

import { useEffect, useState } from 'react';
import { apiFetch } from '@/lib/api';
import { Folder, Code, CheckCircle, Clock, ArrowRight, GitBranch, Star, MessageSquare, Upload, Sparkles, Plus, X, FileCode, AlertCircle } from 'lucide-react';

// 任务类型
type TaskStatus = 'pending' | 'in_progress' | 'completed' | 'failed';

interface Task {
  id: string | number;
  title: string;
  description: string;
  status: TaskStatus;
  aiComment?: string;
}

interface Project {
  id: string;
  name: string;
  description: string;
  techStack: string[];
  tasks: Task[];
  uploadedFiles: string[];
  analysisResult?: {
    completedTasks: Array<string | number>;
    pendingTasks: Array<string | number>;
    suggestions: string[];
  };
}

// 预设项目模板
const projectTemplates = [
  {
    id: 'weather-agent',
    name: '天气查询 Agent',
    description: '使用 LangChain 构建一个能查询天气、提供建议的 AI Agent',
    techStack: ['Python', 'LangChain', 'OpenAI API'],
  },
  {
    id: 'news-agent',
    name: '新闻摘要 Agent',
    description: '构建能自动获取新闻并生成摘要的 AI Agent',
    techStack: ['Python', 'LangChain', 'RAG'],
  },
  {
    id: 'custom',
    name: '自定义项目',
    description: '创建你自己的项目',
    techStack: [],
  },
];

// AI 生成的任务清单示例
const generateTasks = (projectDesc: string): Task[] => {
  // 模拟 AI 根据项目描述生成任务
  if (projectDesc.includes('天气')) {
    return [
      { id: 1, title: '定义 Agent 工具（天气查询 API）', description: '实现调用天气 API 的工具函数', status: 'pending' },
      { id: 2, title: '实现 Agent 主逻辑', description: '使用 LangChain 创建 Agent 并绑定工具', status: 'pending' },
      { id: 3, title: '添加对话记忆功能', description: '使用 ConversationBufferMemory 实现多轮对话', status: 'pending' },
      { id: 4, title: '错误处理与重试机制', description: '处理 API 调用失败的情况', status: 'pending' },
      { id: 5, title: '单元测试', description: '编写测试用例验证 Agent 功能', status: 'pending' },
    ];
  }
  if (projectDesc.includes('新闻')) {
    return [
      { id: 1, title: '新闻数据源接入', description: '实现新闻 API 或 RSS 订阅获取', status: 'pending' },
      { id: 2, title: '文档向量化存储', description: '使用 ChromaDB 存储新闻向量', status: 'pending' },
      { id: 3, title: 'RAG 检索实现', description: '实现检索增强生成流程', status: 'pending' },
      { id: 4, title: '摘要生成逻辑', description: '使用 LLM 生成新闻摘要', status: 'pending' },
      { id: 5, title: '结果展示', description: '格式化输出摘要结果', status: 'pending' },
    ];
  }
  // 默认任务
  return [
    { id: 1, title: '需求分析与设计', description: '明确项目需求和架构设计', status: 'pending' },
    { id: 2, title: '核心功能实现', description: '实现项目的主要功能', status: 'pending' },
    { id: 3, title: '测试与调试', description: '编写测试并修复问题', status: 'pending' },
    { id: 4, title: '代码优化', description: '优化代码质量和性能', status: 'pending' },
    { id: 5, title: '文档编写', description: '编写项目文档', status: 'pending' },
  ];
};

export default function ProjectPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [selectedProject, setSelectedProject] = useState<Project | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [newProject, setNewProject] = useState({ name: '', description: '', techStack: '' });
  const [isGenerating, setIsGenerating] = useState(false);
  const [uploading, setUploading] = useState(false);

  useEffect(() => {
    apiFetch<{ projects: Project[] }>('/api/v1/projects')
      .then((data) => setProjects(data.projects.map((project) => ({ ...project, uploadedFiles: project.uploadedFiles || [] }))))
      .catch(() => undefined);
  }, []);

  // 创建新项目
  const handleCreateProject = async () => {
    if (!newProject.name || !newProject.description) return;
    setIsGenerating(true);
    try {
      const data = await apiFetch<{ success: boolean; project: Project }>('/api/v1/projects', {
        method: 'POST',
        body: JSON.stringify({
          name: newProject.name,
          description: newProject.description,
          tech_stack: newProject.techStack.split(',').map((value) => value.trim()).filter(Boolean),
        }),
      });
      const project = { ...data.project, uploadedFiles: [] };
      setProjects((current) => [project, ...current]);
      setSelectedProject(project);
      setShowCreateModal(false);
      setNewProject({ name: '', description: '', techStack: '' });
    } finally {
      setIsGenerating(false);
    }
  };

  // 上传源码并分析
  const handleUpload = async () => {
    if (!selectedProject) return;
    setUploading(true);
    try {
      const uploadFormData = new FormData();
      uploadFormData.append('projectId', selectedProject.id);
      uploadFormData.append('files', new Blob(['print("CodeForge project")'], { type: 'text/plain' }), 'main.py');
      const uploadResult = await apiFetch<{ task_id: string; files: Array<{ name: string }> }>('/api/v1/projects/upload', {
        method: 'POST', body: uploadFormData,
      });
      const analysisResult = await apiFetch<{ analysis: { files: string[]; completed_tasks: string[]; pending_tasks: string[]; suggestions: string[] } }>('/api/v1/projects/analyze', {
        method: 'POST', body: JSON.stringify({ project_id: selectedProject.id, task_id: uploadResult.task_id }),
      });
      const updatedProject: Project = {
        ...selectedProject,
        uploadedFiles: [...selectedProject.uploadedFiles, ...uploadResult.files.map((file) => file.name)],
        analysisResult: {
          completedTasks: analysisResult.analysis.completed_tasks,
          pendingTasks: analysisResult.analysis.pending_tasks,
          suggestions: analysisResult.analysis.suggestions,
        },
      };
      setSelectedProject(updatedProject);
      setProjects((current) => current.map((project) => project.id === updatedProject.id ? updatedProject : project));
    } catch (error) {
      console.error(error);
    } finally {
      setUploading(false);
    }
  };
  return (
    <div className="max-w-6xl mx-auto">
      {/* 页面标题 */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-foreground">项目实战</h1>
        <p className="text-muted-foreground mt-2">
          创建项目，AI 自动生成任务清单，上传源码智能评估进度
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        {/* 左侧：项目列表 */}
        <div className="lg:col-span-1">
          <h3 className="text-sm font-semibold text-foreground mb-3">我的项目</h3>
          <div className="space-y-2">
            {projects.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground text-sm">
                <Folder className="w-8 h-8 mx-auto mb-2 opacity-50" />
                <p>还没有项目</p>
                <p className="text-xs mt-1">点击下方按钮创建</p>
              </div>
            ) : (
              projects.map(project => (
                <button
                  key={project.id}
                  onClick={() => setSelectedProject(project)}
                  className={`w-full text-left p-3 rounded-xl transition-all ${
                    selectedProject?.id === project.id
                      ? 'bg-primary/10 border border-primary/30'
                      : 'bg-surface border border-border hover:border-primary/20'
                  }`}
                >
                  <div className="flex items-center gap-2 mb-1">
                    <Folder className="w-4 h-4 text-primary" />
                    <span className="text-sm font-medium text-foreground truncate">{project.name}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-xs text-muted-foreground">
                      {project.tasks.filter(t => t.status === 'completed').length}/{project.tasks.length} 任务
                    </span>
                  </div>
                </button>
              ))
            )}
          </div>

          <button 
            onClick={() => setShowCreateModal(true)}
            className="w-full mt-4 py-2.5 border border-dashed border-border rounded-xl text-sm text-muted-foreground hover:text-foreground hover:border-primary/30 transition-colors flex items-center justify-center gap-2"
          >
            <Plus className="w-4 h-4" />
            新建项目
          </button>
        </div>

        {/* 右侧：项目详情 */}
        <div className="lg:col-span-3 space-y-6">
          {!selectedProject ? (
            <div className="text-center py-16 bg-surface border border-border rounded-xl">
              <Folder className="w-12 h-12 mx-auto mb-4 text-muted-foreground/50" />
              <h3 className="text-lg font-medium text-foreground mb-2">选择一个项目或创建新项目</h3>
              <p className="text-sm text-muted-foreground">AI 将根据你的描述自动生成任务清单</p>
            </div>
          ) : (
            <>
              {/* 项目概览 */}
              <div className="p-6 bg-surface border border-border rounded-xl">
                <div className="flex items-start justify-between mb-4">
                  <div>
                    <h2 className="text-xl font-bold text-foreground">{selectedProject.name}</h2>
                    <p className="text-sm text-muted-foreground mt-1">{selectedProject.description}</p>
                  </div>
                </div>

                {/* 进度条 */}
                <div className="mb-4">
                  <div className="flex justify-between text-sm mb-1">
                    <span className="text-muted-foreground">任务完成进度</span>
                    <span className="font-medium text-foreground">
                      {selectedProject.tasks.filter(t => t.status === 'completed').length}/{selectedProject.tasks.length}
                    </span>
                  </div>
                  <div className="h-2 bg-surface-container rounded-full overflow-hidden">
                    <div 
                      className="h-full bg-primary rounded-full transition-all"
                      style={{ width: `${(selectedProject.tasks.filter(t => t.status === 'completed').length / selectedProject.tasks.length) * 100}%` }}
                    />
                  </div>
                </div>

                {/* 技术栈 */}
                {selectedProject.techStack.length > 0 && (
                  <div className="flex flex-wrap gap-2">
                    {selectedProject.techStack.map((tech, i) => (
                      <span key={i} className="px-2 py-1 text-xs bg-surface-container text-foreground rounded">
                        {tech}
                      </span>
                    ))}
                  </div>
                )}
              </div>

              {/* 上传源码 */}
              <div className="p-6 bg-surface border border-border rounded-xl">
                <h3 className="font-semibold text-foreground mb-4 flex items-center gap-2">
                  <Upload className="w-4 h-4 text-primary" />
                  上传源码
                </h3>
                <p className="text-sm text-muted-foreground mb-4">
                  上传你的项目源码，AI 将自动分析代码并评估任务完成情况
                </p>
                
                {selectedProject.uploadedFiles.length > 0 ? (
                  <div className="space-y-3">
                    <div className="flex flex-wrap gap-2">
                      {selectedProject.uploadedFiles.map((file, i) => (
                        <span key={i} className="flex items-center gap-1 px-2 py-1 text-xs bg-surface-container text-foreground rounded">
                          <FileCode className="w-3 h-3" />
                          {file}
                        </span>
                      ))}
                    </div>
                    
                    {/* AI 分析结果 */}
                    {selectedProject.analysisResult && (
                      <div className="p-4 bg-surface-container rounded-lg space-y-3">
                        <h4 className="text-sm font-medium text-foreground flex items-center gap-2">
                          <Sparkles className="w-4 h-4 text-primary" />
                          AI 分析结果
                        </h4>
                        <div className="grid grid-cols-2 gap-4 text-sm">
                          <div>
                            <span className="text-green-600 font-medium">✓ 已完成</span>
                            <span className="text-muted-foreground ml-2">
                              {selectedProject.analysisResult.completedTasks.length} 个任务
                            </span>
                          </div>
                          <div>
                            <span className="text-orange-600 font-medium">○ 待完成</span>
                            <span className="text-muted-foreground ml-2">
                              {selectedProject.analysisResult.pendingTasks.length} 个任务
                            </span>
                          </div>
                        </div>
                        <div className="space-y-2">
                          {selectedProject.analysisResult.suggestions.map((s, i) => (
                            <div key={i} className="flex items-start gap-2 text-sm">
                              <AlertCircle className="w-4 h-4 text-blue-500 mt-0.5 flex-shrink-0" />
                              <span className="text-muted-foreground">{s}</span>
                            </div>
                          ))}
                        </div>
                      </div>
                    )}
                    
                    <button 
                      onClick={handleUpload}
                      disabled={uploading}
                      className="w-full py-2 border border-border rounded-lg text-sm text-muted-foreground hover:text-foreground hover:border-primary/30 transition-colors disabled:opacity-50"
                    >
                      {uploading ? '分析中...' : '重新上传并分析'}
                    </button>
                  </div>
                ) : (
                  <button 
                    onClick={handleUpload}
                    disabled={uploading}
                    className="w-full py-8 border-2 border-dashed border-border rounded-xl text-sm text-muted-foreground hover:text-foreground hover:border-primary/30 transition-colors flex flex-col items-center gap-2 disabled:opacity-50"
                  >
                    {uploading ? (
                      <>
                        <div className="w-6 h-6 border-2 border-primary border-t-transparent rounded-full animate-spin" />
                        <span>AI 正在分析代码...</span>
                      </>
                    ) : (
                      <>
                        <Upload className="w-6 h-6" />
                        <span>点击上传源码文件</span>
                        <span className="text-xs">支持 .py, .js, .ts, .java, .cpp 等</span>
                      </>
                    )}
                  </button>
                )}
              </div>

              {/* 任务列表 */}
              <div className="p-6 bg-surface border border-border rounded-xl">
                <h3 className="font-semibold text-foreground mb-4 flex items-center gap-2">
                  <GitBranch className="w-4 h-4 text-primary" />
                  任务清单
                  <span className="text-xs text-muted-foreground font-normal ml-auto">
                    AI 自动生成
                  </span>
                </h3>
                <div className="space-y-2">
                  {selectedProject.tasks.map(task => (
                    <div key={task.id} className="flex items-start gap-3 p-3 bg-surface-container rounded-lg">
                      {task.status === 'completed' ? (
                        <CheckCircle className="w-5 h-5 text-green-500 mt-0.5" />
                      ) : task.status === 'in_progress' ? (
                        <div className="w-5 h-5 border-2 border-primary rounded-full animate-pulse mt-0.5" />
                      ) : (
                        <div className="w-5 h-5 border-2 border-border rounded-full mt-0.5" />
                      )}
                      <div className="flex-1">
                        <span className={`text-sm font-medium ${
                          task.status === 'completed' ? 'text-muted-foreground line-through' : 'text-foreground'
                        }`}>
                          {task.title}
                        </span>
                        <p className="text-xs text-muted-foreground mt-0.5">{task.description}</p>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </>
          )}
        </div>
      </div>

      {/* 创建项目弹窗 */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-foreground/20 backdrop-blur-sm flex items-center justify-center z-50">
          <div className="w-full max-w-md bg-surface border border-border rounded-2xl p-6 shadow-lg">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-foreground">新建项目</h3>
              <button onClick={() => setShowCreateModal(false)} className="text-muted-foreground hover:text-foreground">
                <X className="w-5 h-5" />
              </button>
            </div>
            
            <div className="space-y-4">
              <div>
                <label className="text-sm font-medium text-foreground mb-1.5 block">项目名称</label>
                <input
                  type="text"
                  value={newProject.name}
                  onChange={e => setNewProject({ ...newProject, name: e.target.value })}
                  placeholder="如：天气查询 Agent"
                  className="w-full px-4 py-2.5 bg-surface-container border-none rounded-lg text-foreground placeholder:text-muted-foreground/40 focus:outline-none focus:ring-2 focus:ring-primary/30"
                />
              </div>
              
              <div>
                <label className="text-sm font-medium text-foreground mb-1.5 block">项目描述</label>
                <textarea
                  value={newProject.description}
                  onChange={e => setNewProject({ ...newProject, description: e.target.value })}
                  placeholder="描述你想要实现的功能，AI 将根据描述生成任务清单..."
                  rows={3}
                  className="w-full px-4 py-2.5 bg-surface-container border-none rounded-lg text-foreground placeholder:text-muted-foreground/40 focus:outline-none focus:ring-2 focus:ring-primary/30 resize-none"
                />
              </div>
              
              <div>
                <label className="text-sm font-medium text-foreground mb-1.5 block">技术栈（可选）</label>
                <input
                  type="text"
                  value={newProject.techStack}
                  onChange={e => setNewProject({ ...newProject, techStack: e.target.value })}
                  placeholder="如：Python, LangChain, OpenAI"
                  className="w-full px-4 py-2.5 bg-surface-container border-none rounded-lg text-foreground placeholder:text-muted-foreground/40 focus:outline-none focus:ring-2 focus:ring-primary/30"
                />
              </div>
              
              <button
                onClick={handleCreateProject}
                disabled={!newProject.name || !newProject.description || isGenerating}
                className="w-full py-3 bg-primary text-primary-foreground rounded-lg font-medium hover:opacity-90 transition-opacity disabled:opacity-50 flex items-center justify-center gap-2"
              >
                {isGenerating ? (
                  <>
                    <div className="w-4 h-4 border-2 border-primary-foreground border-t-transparent rounded-full animate-spin" />
                    AI 正在生成任务清单...
                  </>
                ) : (
                  <>
                    <Sparkles className="w-4 h-4" />
                    创建并生成任务
                  </>
                )}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
