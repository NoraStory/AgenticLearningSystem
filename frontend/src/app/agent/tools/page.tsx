'use client';

import { useEffect, useState } from 'react';
import { apiFetch } from '@/lib/api';
import { Search, Globe, Code, FileText, Github, BookOpen, Play, Type, CheckCircle, XCircle, Lock } from 'lucide-react';

// 15个 Tool 插件
const fallbackTools = [
  { id: 'web_search', name: '????', desc: 'DuckDuckGo ??', icon: Globe, category: 'search', enabled: true, free: true },
  { id: 'doc_reader', name: '????', desc: '?????????', icon: FileText, category: 'search', enabled: true, free: true },
  { id: 'code_search', name: '????', desc: '????????', icon: Code, category: 'code', enabled: true, free: true },
  { id: 'git_helper', name: 'Git ??', desc: '??????????', icon: Github, category: 'code', enabled: true, free: true },
  { id: 'leetcode_fetch', name: '????', desc: '???????', icon: Code, category: 'search', enabled: true, free: true },
  { id: 'code_execute', name: '????', desc: '??????', icon: Play, category: 'sandbox', enabled: true, free: true, locked: true },
  { id: 'sql_explain', name: 'SQL ??', desc: 'SQL ?????????', icon: CheckCircle, category: 'code', enabled: true, free: true },
  { id: 'diagram_gen', name: '????', desc: 'Mermaid ???????', icon: Type, category: 'generate', enabled: true, free: true },
  { id: 'quiz_gen', name: '????', desc: '?????????', icon: BookOpen, category: 'generate', enabled: true, free: true },
  { id: 'self_heal', name: '???', desc: '???????????', icon: XCircle, category: 'sandbox', enabled: true, free: true, locked: true },
  { id: 'course_search', name: '????', desc: 'PostgreSQL ????', icon: Search, category: 'data', enabled: true, free: true },
  { id: 'progress_query', name: '????', desc: '??????', icon: FileText, category: 'data', enabled: true, free: true },
  { id: 'resume_review', name: '????', desc: '?????????', icon: FileText, category: 'generate', enabled: true, free: false },
  { id: 'project_review', name: '????', desc: '??????', icon: Github, category: 'code', enabled: true, free: false },
  { id: 'mindmap_gen', name: '????', desc: 'Mermaid ??????', icon: Type, category: 'generate', enabled: true, free: true },
]

const categories = [
  { id: 'all', name: '全部' },
  { id: 'search', name: '搜索类' },
  { id: 'code', name: '代码类' },
  { id: 'sandbox', name: '沙箱类' },
  { id: 'data', name: '数据类' },
  { id: 'generate', name: '生成类' },
];

export default function AgentToolsPage() {
  const [toolList, setToolList] = useState(fallbackTools);
  const [activeCategory, setActiveCategory] = useState('all');
  const [toolStats, setToolStats] = useState<Record<string, { usable: boolean; reason: string }>>({});

  useEffect(() => {
    apiFetch<{ tools: Array<{ id: string; name: string; desc: string; category: string; enabled: boolean; locked: boolean; free: boolean; usable: boolean; reason: string }> }>('/api/v1/agent/tools')
      .then((data) => {
        setToolList(data.tools.map((tool, index) => {
          const fallback = fallbackTools.find((item) => item.id === tool.id) || fallbackTools[index % fallbackTools.length];
          const category = tool.id.includes('search') || tool.id.includes('fetch') ? 'search' : tool.id.includes('code') || tool.id.includes('git') || tool.id.includes('sql') || tool.id.includes('project') ? 'code' : tool.id.includes('execute') || tool.id.includes('heal') ? 'sandbox' : tool.id.includes('course') || tool.id.includes('progress') ? 'data' : 'generate';
          return { ...fallback, ...tool, category };
        }));
        setToolStats(Object.fromEntries(data.tools.map((tool) => [tool.id, { usable: tool.usable, reason: tool.reason }])));
      })
      .catch(() => undefined);
  }, []);

  const toggleTool = async (id: string) => {
    const tool = toolList.find((item) => item.id === id);
    if (!tool || tool.locked) return;
    const result = await apiFetch<{ id: string; enabled: boolean }>('/api/v1/agent/tools/' + id, {
      method: 'PATCH', body: JSON.stringify({ enabled: !tool.enabled }),
    });
    setToolList((current) => current.map((item) => item.id === id ? { ...item, enabled: result.enabled } : item));
  };

  const filteredTools = activeCategory === 'all' 
    ? toolList 
    : toolList.filter(t => t.category === activeCategory);

  const enabledCount = toolList.filter(t => t.enabled).length;

  return (
    <div className="max-w-6xl mx-auto">
      {/* 页面标题 */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-foreground">Tool 插件管理</h1>
        <p className="text-muted-foreground mt-2">
          管理 AI Agent 可用的工具插件。沙箱类工具始终启用以确保代码执行安全。
        </p>
        <div className="flex items-center gap-4 mt-4">
          <span className="text-sm text-muted-foreground">
            已启用 <span className="font-semibold text-foreground">{enabledCount}</span> / {toolList.length}
          </span>
          <div className="h-2 flex-1 max-w-xs bg-surface-container rounded-full overflow-hidden">
            <div 
              className="h-full bg-primary rounded-full transition-all"
              style={{ width: `${(enabledCount / toolList.length) * 100}%` }}
            />
          </div>
        </div>
      </div>

      {/* 分类筛选 */}
      <div className="flex gap-2 mb-6 flex-wrap">
        {categories.map(cat => (
          <button
            key={cat.id}
            onClick={() => setActiveCategory(cat.id)}
            className={`px-4 py-1.5 text-sm rounded-full transition-colors ${
              activeCategory === cat.id
                ? 'bg-primary text-primary-foreground'
                : 'bg-surface-container text-muted-foreground hover:text-foreground'
            }`}
          >
            {cat.name}
          </button>
        ))}
      </div>

      {/* Tool 列表 */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {filteredTools.map(tool => {
          const Icon = tool.icon;
          return (
            <div
              key={tool.id}
              className={`p-4 bg-surface border border-border rounded-xl transition-all ${
                tool.enabled ? 'opacity-100' : 'opacity-60'
              }`}
            >
              <div className="flex items-start gap-3">
                <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${
                  tool.enabled ? 'bg-primary/10' : 'bg-surface-container'
                }`}>
                  <Icon className={`w-5 h-5 ${tool.enabled ? 'text-primary' : 'text-muted-foreground'}`} />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <h3 className="font-medium text-foreground">{tool.name}</h3>
                    {tool.locked && (
                      <Lock className="w-3 h-3 text-muted-foreground" />
                    )}
                    <span className="text-xs px-1.5 py-0.5 bg-green-100 text-green-700 rounded">免费</span>
                  </div>
                                    <p className="text-sm text-muted-foreground mt-0.5">{tool.desc}</p>
                  <p className={`text-xs mt-1 ${toolStats[tool.id]?.usable ? 'text-emerald-600' : 'text-destructive'}`}>
                    {toolStats[tool.id]?.usable ? '可用' : '不可用'} · {toolStats[tool.id]?.reason || '检测中...'}
                  </p>
                  <p className="text-xs text-muted-foreground mt-1 font-mono">{tool.id}</p>
                </div>
                <button
                  onClick={() => toggleTool(tool.id)}
                  disabled={tool.locked}
                  className={`relative w-11 h-6 rounded-full transition-colors flex-shrink-0 ${
                    tool.locked
                      ? 'bg-primary cursor-not-allowed'
                      : tool.enabled
                        ? 'bg-primary'
                        : 'bg-surface-container'
                  }`}
                >
                  <span 
                    className="absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full shadow transition-transform"
                    style={{ transform: (tool.enabled || tool.locked) ? 'translateX(20px)' : 'translateX(0)' }}
                  />
                </button>
              </div>
            </div>
          );
        })}
      </div>

      {/* 说明 */}
      <div className="mt-8 p-4 bg-surface-container rounded-xl">
        <h4 className="font-medium text-foreground mb-2">说明</h4>
        <ul className="text-sm text-muted-foreground space-y-1">
          <li>• 沙箱类工具（代码执行、格式化、静态分析、调试、自修复）始终启用，无法关闭</li>
          <li>• Agent 只会调用已启用的 Tool</li>
          <li>• Tool 调用通过串行化避免 API 限流</li>
          <li>• 所有 Tool 均为免费数据源</li>
        </ul>
      </div>
    </div>
  );
}
