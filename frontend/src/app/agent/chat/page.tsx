'use client';

import { useState, useEffect, useRef } from 'react';
import { Send, Bot, User, Sparkles, Code, BookOpen, MessageSquare, Loader2, Paperclip, Image, X, File, GitBranch, Plus, MessageCircle, Trash2 } from 'lucide-react';
import { apiFetch } from '@/lib/api';

// Agent 类型
const agents = [
  { id: 'learning', name: '学习助手', icon: BookOpen, desc: '概念问答，RAG检索', color: 'bg-blue-500' },
  { id: 'reviewer', name: '代码审查', icon: Code, desc: '代码审查，优化建议', color: 'bg-green-500' },
  { id: 'tutor', name: '题目讲解', icon: MessageSquare, desc: '题目讲解，解题引导', color: 'bg-orange-500' },
  { id: 'mentor', name: '项目导师', icon: Sparkles, desc: '项目实战，架构建议', color: 'bg-purple-500' },
  { id: 'coach', name: '面试教练', icon: Bot, desc: '面试模拟，出题评分', color: 'bg-rose-500' },
  { id: 'community', name: '社区助手', icon: MessageSquare, desc: '博客/README生成', color: 'bg-cyan-500' },
];

// 工作流步骤类型
type WorkflowStep = {
  id?: string;
  name: string;
  status: 'pending' | 'running' | 'done' | 'failed';
  agent?: string;
  tool?: string;
  reason?: string;
  result?: string;
};

type ChatMessage = {
  role: 'user' | 'assistant';
  agent: string;
  content: string;
  workflow: WorkflowStep[] | null;
  artifacts?: string[];
};

// 历史会话摘要
type SessionSummary = {
  session_id: string;
  title: string;
  agent: string;
  message_count: number;
  created_at: string;
  updated_at: string;
};

// 初始消息
const mockMessages: ChatMessage[] = [
  { role: 'assistant', agent: 'learning', content: '你好！我是学习助手，可以帮你解答编程概念、检索课程资源。有什么可以帮你的吗？', workflow: null as WorkflowStep[] | null },
];


function extractMermaidBlocks(content: string): string[] {
  const blocks: string[] = [];
  const pattern = /```mermaid\s*([\s\S]*?)```/gi;
  let match: RegExpExecArray | null;
  while ((match = pattern.exec(content)) !== null) {
    const value = match[1].trim();
    if (value && !blocks.includes(value)) blocks.push(value);
  }
  return blocks;
}

type MermaidNode = { id: string; label: string };
type MermaidEdge = { from: string; to: string };

function parseMermaidFlowchart(source: string): { nodes: MermaidNode[]; edges: MermaidEdge[] } {
  const nodes = new Map<string, MermaidNode>();
  const edges: MermaidEdge[] = [];
  const nodePattern = /([A-Za-z0-9_-]+)\s*(?:\[([^\]]+)\]|\{([^}]+)\})/g;
  const edgePattern = /([A-Za-z0-9_-]+)\s*-->(?:\|[^|]*\|)?\s*([A-Za-z0-9_-]+)/g;
  let match: RegExpExecArray | null;
  while ((match = nodePattern.exec(source)) !== null) nodes.set(match[1], { id: match[1], label: match[2] || match[3] });
  while ((match = edgePattern.exec(source)) !== null) {
    edges.push({ from: match[1], to: match[2] });
    if (!nodes.has(match[1])) nodes.set(match[1], { id: match[1], label: match[1] });
    if (!nodes.has(match[2])) nodes.set(match[2], { id: match[2], label: match[2] });
  }
  return { nodes: Array.from(nodes.values()), edges };
}

function MermaidPreview({ source }: { source: string }) {
  if (!/^\s*(flowchart|graph)\s+/i.test(source)) {
    return <pre className="mt-3 overflow-x-auto rounded-lg bg-slate-950 p-3 text-xs text-slate-100">{source}</pre>;
  }
  const { nodes, edges } = parseMermaidFlowchart(source);
  if (!nodes.length) return <pre className="mt-3 overflow-x-auto rounded-lg bg-slate-950 p-3 text-xs text-slate-100">{source}</pre>;
  const width = 560;
  const nodeWidth = 440;
  const nodeHeight = 48;
  const gap = 36;
  const height = Math.max(120, nodes.length * (nodeHeight + gap) + 24);
  const index = new Map(nodes.map((node, i) => [node.id, i]));
  const y = (id: string) => (index.get(id) ?? 0) * (nodeHeight + gap) + 12;
  return (
    <div className="mt-3 overflow-x-auto rounded-lg border border-border bg-background p-2">
      <svg viewBox={`0 0 ${width} ${height}`} className="min-w-[520px]" role="img" aria-label="Mermaid flowchart">
        <defs>
          <marker id="codeforge-arrow" markerWidth="8" markerHeight="8" refX="7" refY="3" orient="auto" markerUnits="strokeWidth">
            <path d="M0,0 L0,6 L7,3 z" fill="currentColor" />
          </marker>
        </defs>
        {edges.map((edge, i) => {
          const fromY = y(edge.from) + nodeHeight;
          const toY = y(edge.to);
          return <line key={`${edge.from}-${edge.to}-${i}`} x1={width / 2} y1={fromY} x2={width / 2} y2={toY} stroke="currentColor" strokeWidth="2" markerEnd="url(#codeforge-arrow)" className="text-primary" />;
        })}
        {nodes.map((node) => (
          <g key={node.id}>
            <rect x={(width - nodeWidth) / 2} y={y(node.id)} width={nodeWidth} height={nodeHeight} rx="12" className="fill-primary/10 stroke-primary" strokeWidth="2" />
            <text x={width / 2} y={y(node.id) + 29} textAnchor="middle" className="fill-foreground text-[14px]">{node.label}</text>
          </g>
        ))}
      </svg>
    </div>
  );
}

function formatRelativeTime(iso: string): string {
  const date = new Date(iso);
  if (isNaN(date.getTime())) return '';
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  const sameDay = date.toDateString() === now.toDateString();
  if (sameDay) return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
  const yesterday = new Date(now);
  yesterday.setDate(now.getDate() - 1);
  if (date.toDateString() === yesterday.toDateString()) return '昨天';
  return date.toLocaleDateString('zh-CN', { month: 'numeric', day: 'numeric' });
}

export default function AgentChatPage() {
  const agentApiBase = process.env.NEXT_PUBLIC_API_BASE_URL || '';
  const [messages, setMessages] = useState<ChatMessage[]>(mockMessages);
  const [sessionId, setSessionId] = useState('');
  const [input, setInput] = useState('');
  const [selectedAgent, setSelectedAgent] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [attachments, setAttachments] = useState<Array<{ name: string; type: string; preview?: string; file: globalThis.File }>>([]);
  const [connectionError, setConnectionError] = useState<string | null>(null);
  const [sessions, setSessions] = useState<SessionSummary[]>([]);
  const [loadingSessions, setLoadingSessions] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const imageInputRef = useRef<HTMLInputElement>(null);

  // 把后端返回的历史消息还原成前端可渲染的结构
  const restoreMessages = (items: Array<{ role: string; agent: string; content: string; workflow?: unknown }>): ChatMessage[] =>
    items
      .filter((item) => item.role === 'user' || item.role === 'assistant')
      .map((item) => {
        const workflow = Array.isArray(item.workflow) ? (item.workflow as WorkflowStep[]) : null;
        const artifacts = workflow?.flatMap((step) => extractMermaidBlocks(step.result || '')) || [];
        return {
          role: item.role as ChatMessage['role'],
          agent: item.agent || 'learning',
          content: item.content || '',
          workflow,
          artifacts,
        };
      });

  // 拉取当前用户的全部历史会话（按最近活跃倒序）
  const loadSessions = () => {
    setLoadingSessions(true);
    apiFetch<{ items: SessionSummary[] }>(`${agentApiBase}/api/v1/agent/sessions`)
      .then((data) => setSessions(data.items || []))
      .catch(() => undefined)
      .finally(() => setLoadingSessions(false));
  };

  // 加载某个会话的完整消息历史
  const loadHistory = (sid: string) => {
    if (!sid) {
      setMessages(mockMessages);
      return;
    }
    apiFetch<{ items: Array<{ role: string; agent: string; content: string; workflow?: unknown }> }>(
      `${agentApiBase}/api/v1/agent/history?session_id=${encodeURIComponent(sid)}`,
    )
      .then((data) => {
        const restored = restoreMessages(data.items || []);
        setMessages(restored.length ? restored : mockMessages);
      })
      .catch(() => setMessages(mockMessages));
  };

  // 开启新对话：生成全新会话 ID，清空消息，回到欢迎语
  const startNewConversation = () => {
    const sid = crypto.randomUUID();
    localStorage.setItem('codeforge_agent_session_id', sid);
    setSessionId(sid);
    setMessages(mockMessages);
    setAttachments([]);
    setSelectedAgent(null);
  };

  // 切换到某条历史对话
  const switchSession = (sid: string) => {
    if (sid === sessionId) return;
    localStorage.setItem('codeforge_agent_session_id', sid);
    setSessionId(sid);
    setAttachments([]);
    loadHistory(sid);
  };

  // 删除历史对话
  const deleteSession = (sid: string, e: React.MouseEvent) => {
    e.stopPropagation();
    apiFetch(`${agentApiBase}/api/v1/agent/sessions/${encodeURIComponent(sid)}`, { method: 'DELETE' })
      .then(() => {
        setSessions((prev) => prev.filter((item) => item.session_id !== sid));
        if (sid === sessionId) startNewConversation();
      })
      .catch(() => undefined);
  };

  useEffect(() => {
    if (typeof window === 'undefined') return;
    const storageKey = 'codeforge_agent_session_id';
    const existingSessionId = localStorage.getItem(storageKey) || crypto.randomUUID();
    localStorage.setItem(storageKey, existingSessionId);
    setSessionId(existingSessionId);

    apiFetch<{ tools: Array<{ id: string; enabled: boolean }> }>(`${agentApiBase}/api/v1/agent/tools`).catch((error) => {
      setConnectionError(error instanceof Error ? error.message : 'Agent 服务不可用');
    });
    loadSessions();
    loadHistory(existingSessionId);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [agentApiBase]);

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (!files) return;
    
    const newAttachments = Array.from(files).map(file => ({
      name: file.name,
      type: file.type,
      preview: file.type.startsWith('image/') ? URL.createObjectURL(file) : undefined,
      file,
    }));
    
    setAttachments(prev => [...prev, ...newAttachments]);
    e.target.value = '';
  };

  // 移除附件
  const removeAttachment = (index: number) => {
    setAttachments(prev => prev.filter((_, i) => i !== index));
  };

  const handleSend = async () => {
    const message = input.trim();
    if ((!message && attachments.length === 0) || isLoading) return;
    const activeSessionId = sessionId || crypto.randomUUID();
    if (!sessionId) {
      setSessionId(activeSessionId);
      localStorage.setItem('codeforge_agent_session_id', activeSessionId);
    }
    const userMsg: ChatMessage = { role: 'user', agent: 'user', content: message || '请分析我上传的附件。', workflow: null };
    const assistantMsg: ChatMessage = { role: 'assistant', agent: selectedAgent || 'learning', content: '', workflow: [], artifacts: [] };
    setMessages((prev) => [...prev, userMsg, assistantMsg]);
    setInput('');
    setIsLoading(true);

    try {
      let uploaded: string[] = [];
      if (attachments.length) {
        const form = new FormData();
        attachments.forEach((attachment) => form.append('files', attachment.file));
        const upload = await apiFetch<{ file_urls: string[] }>(`${agentApiBase}/api/v1/agent/chat/upload`, { method: 'POST', body: form });
        uploaded = upload.file_urls;
      }
      const response = await fetch(`${agentApiBase}/api/v1/agent/chat`, {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ session_id: activeSessionId, message: message || '请分析我上传的图片或附件。', agent_type: selectedAgent || undefined, attachments: uploaded, collaboration_mode: 'dynamic', context: { current_page: '/agent/chat' } }),
      });
      if (!response.ok || !response.body) throw new Error(`Agent 服务不可用（HTTP ${response.status}）`);
      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let buffer = '';
      const updateLast = (updater: (value: typeof assistantMsg) => typeof assistantMsg) => {
        setMessages((current) => current.map((item, index) => index === current.length - 1 ? updater(item as typeof assistantMsg) : item));
      };
      while (true) {
        const { value, done } = await reader.read();
        buffer += decoder.decode(value || new Uint8Array(), { stream: !done });
        const blocks = buffer.split('\n\n');
        buffer = blocks.pop() || '';
        for (const block of blocks) {
          const event = block.split('\n').find((line) => line.startsWith('event:'))?.slice(6).trim();
          const dataLine = block.split('\n').find((line) => line.startsWith('data:'))?.slice(5).trim();
          if (!event || !dataLine) continue;
          const data = JSON.parse(dataLine) as { id?: string; agent?: string; name?: string; status?: string; content?: string; tool?: string; reason?: string; result?: string; session_id?: string };
          if (event === 'agent_route') updateLast((item) => ({ ...item, agent: data.agent || item.agent }));
          if (event === 'done' && data.session_id) {
            setSessionId(data.session_id);
            localStorage.setItem('codeforge_agent_session_id', data.session_id);
            // 回答完成后刷新历史会话列表，让新对话立刻出现在侧栏
            loadSessions();
          }
          if (event === 'tool_result' && data.result) {
            const diagrams = extractMermaidBlocks(data.result);
            if (diagrams.length) updateLast((item) => ({ ...item, artifacts: [...new Set([...(item.artifacts || []), ...diagrams])] }));
          }
          if (event === 'workflow_step') updateLast((item) => {
            const step: WorkflowStep = { id: data.id, name: data.name || '工作流步骤', status: data.status === 'completed' ? 'done' : data.status === 'failed' || data.status === 'unavailable' ? 'failed' : data.status === 'running' ? 'running' : 'pending', agent: data.agent, tool: data.tool, reason: data.reason, result: data.result };
            const current = item.workflow || [];
            const existingIndex = current.findIndex((value) => (step.id && value.id === step.id) || (!step.id && value.name === step.name));
            return { ...item, workflow: existingIndex >= 0 ? current.map((value, index) => index === existingIndex ? { ...value, ...step } : value) : [...current, step] };
          });
          if (event === 'token') updateLast((item) => ({ ...item, content: item.content + String(data.content || '') }));
        }
        if (done) break;
      }
      setAttachments([]);
    } catch (error) {
      setMessages((current) => current.map((item, index) => index === current.length - 1 ? { ...item, content: error instanceof Error ? error.message : '请求失败' } : item));
    } finally {
      setIsLoading(false);
    }
  };

  const activeWorkflow = messages[messages.length - 1]?.workflow || [];

  return (
    <div className="flex h-[calc(100vh-8rem)]">
      {/* 左侧侧栏：新对话 + 对话历史 + Agent 选择 */}
      <div className="w-72 border-r border-border bg-surface-container-lowest flex flex-col">
        {/* 新对话按钮 */}
        <div className="p-3 border-b border-border">
          <button
            onClick={startNewConversation}
            className="w-full flex items-center justify-center gap-2 px-3 py-2.5 bg-primary text-primary-foreground rounded-xl hover:opacity-90 transition-opacity text-sm font-medium"
          >
            <Plus className="w-4 h-4" />
            新对话
          </button>
        </div>

        {/* 对话历史 */}
        <div className="flex-1 overflow-y-auto px-2 py-2 min-h-0">
          <div className="flex items-center justify-between px-1 py-1.5">
            <h4 className="text-xs font-medium text-muted-foreground">对话历史</h4>
            {loadingSessions && <Loader2 className="w-3 h-3 text-muted-foreground animate-spin" />}
          </div>
          <div className="space-y-0.5">
            {sessions.length === 0 ? (
              <div className="text-xs text-muted-foreground/60 px-2 py-4 text-center">
                {loadingSessions ? '加载中…' : '暂无历史对话'}
              </div>
            ) : (
              sessions.map((session) => (
                <div
                  key={session.session_id}
                  onClick={() => switchSession(session.session_id)}
                  className={`group flex items-start gap-2 px-2 py-2 rounded-lg cursor-pointer transition-colors ${
                    session.session_id === sessionId ? 'bg-primary/10 border border-primary/30' : 'hover:bg-surface-container'
                  }`}
                >
                  <MessageCircle className={`w-4 h-4 mt-0.5 flex-shrink-0 ${session.session_id === sessionId ? 'text-primary' : 'text-muted-foreground'}`} />
                  <div className="flex-1 min-w-0">
                    <div className="text-sm text-foreground truncate">{session.title || '新对话'}</div>
                    <div className="text-xs text-muted-foreground/70 flex items-center gap-1.5">
                      <span>{session.message_count} 条</span>
                      <span>·</span>
                      <span>{formatRelativeTime(session.updated_at)}</span>
                    </div>
                  </div>
                  <button
                    onClick={(e) => deleteSession(session.session_id, e)}
                    className="opacity-0 group-hover:opacity-100 text-muted-foreground hover:text-destructive transition-opacity flex-shrink-0"
                    title="删除对话"
                  >
                    <Trash2 className="w-3.5 h-3.5" />
                  </button>
                </div>
              ))
            )}
          </div>
        </div>

        {/* Agent 选择 */}
        <div className="p-3 border-t border-border max-h-[40%] overflow-y-auto">
          <h3 className="text-sm font-semibold text-foreground mb-2">选择 Agent</h3>
          <div className="space-y-1">
            {agents.map(agent => {
              const Icon = agent.icon;
              return (
                <button
                  key={agent.id}
                  onClick={() => setSelectedAgent(agent.id === selectedAgent ? null : agent.id)}
                  className={`w-full flex items-center gap-2 px-2 py-1.5 rounded-lg text-left transition-all ${
                    selectedAgent === agent.id
                      ? 'bg-primary/10 border border-primary/30'
                      : 'hover:bg-surface-container'
                  }`}
                >
                  <div className={`w-6 h-6 rounded-md ${agent.color} flex items-center justify-center flex-shrink-0`}>
                    <Icon className="w-3.5 h-3.5 text-white" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="text-xs font-medium text-foreground truncate">{agent.name}</div>
                    <div className="text-[11px] text-muted-foreground truncate">{agent.desc}</div>
                  </div>
                </button>
              );
            })}
          </div>

          <div className="mt-3 pt-3 border-t border-border">
            <h4 className="text-xs font-medium text-muted-foreground mb-2">协作模式</h4>
            <div className="space-y-1">
              <button className="w-full text-left px-2 py-1 text-xs rounded hover:bg-surface-container text-foreground">串行接力</button>
              <button className="w-full text-left px-2 py-1 text-xs rounded hover:bg-surface-container text-foreground">并行合并</button>
              <button className="w-full text-left px-2 py-1 text-xs rounded hover:bg-surface-container text-foreground">辩论模式</button>
              <button className="w-full text-left px-2 py-1 text-xs rounded hover:bg-surface-container text-foreground">师徒模式</button>
            </div>
          </div>
        </div>
      </div>

      {/* 中间聊天区域 */}
      <div className="flex-1 flex flex-col">
        {/* 消息列表 */}
        <div className="flex-1 overflow-y-auto p-6 space-y-4">
          {messages.map((msg, i) => {
            const agent = msg.agent ? agents.find(a => a.id === msg.agent) : null;
            const Icon = agent?.icon || Bot;
            
            return (
              <div key={i} className={`flex gap-3 ${msg.role === 'user' ? 'justify-end' : ''}`}>
                {msg.role === 'assistant' && (
                  <div className={`w-8 h-8 rounded-lg ${agent?.color || 'bg-primary'} flex items-center justify-center flex-shrink-0`}>
                    <Icon className="w-4 h-4 text-white" />
                  </div>
                )}
                <div className={`max-w-[70%] px-4 py-2.5 rounded-2xl ${
                  msg.role === 'user'
                    ? 'bg-primary text-primary-foreground'
                    : 'bg-surface-container text-foreground'
                }`}>
                  {msg.role === 'assistant' && agent && (
                    <div className="text-xs font-medium mb-1 opacity-70">{agent.name}</div>
                  )}
                  <div className="text-sm whitespace-pre-wrap">{msg.content}</div>
                  {msg.role === 'assistant' && [...(msg.artifacts || []), ...extractMermaidBlocks(msg.content)].filter((value, index, values) => values.indexOf(value) === index).map((diagram, index) => (
                    <MermaidPreview key={`${i}-diagram-${index}`} source={diagram} />
                  ))}
                </div>
                {msg.role === 'user' && (
                  <div className="w-8 h-8 rounded-lg bg-muted flex items-center justify-center flex-shrink-0">
                    <User className="w-4 h-4 text-muted-foreground" />
                  </div>
                )}
              </div>
            );
          })}
          
          {isLoading && (
            <div className="space-y-3">
              {/* 工作流执行状态 */}
              <div className="max-w-4xl mx-auto">
                <div className="bg-surface-container-lowest border border-border rounded-xl p-4">
                  <div className="flex items-center gap-2 mb-3">
                    <GitBranch className="w-4 h-4 text-primary" />
                    <span className="text-sm font-medium text-foreground">工作流执行中</span>
                    <span className="text-xs text-muted-foreground">动态 Tool 规划</span>
                  </div>
                  <div className="space-y-2">
                    {activeWorkflow.map((step, index) => (
                      <div key={step.id || index} className="flex flex-wrap items-center gap-3">
                        {step.status === 'done' ? (
                          <div className="w-5 h-5 rounded-full bg-primary/20 flex items-center justify-center">
                            <svg className="w-3 h-3 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                            </svg>
                          </div>
                        ) : step.status === 'failed' ? (
                          <div className="w-5 h-5 rounded-full bg-destructive/20 flex items-center justify-center text-xs text-destructive">!</div>
                        ) : step.status === 'running' ? (
                          <div className="w-5 h-5 rounded-full bg-primary/20 flex items-center justify-center">
                            <Loader2 className="w-3 h-3 text-primary animate-spin" />
                          </div>
                        ) : (
                          <div className="w-5 h-5 rounded-full bg-surface-container flex items-center justify-center">
                            <div className="w-2 h-2 rounded-full bg-muted-foreground/30" />
                          </div>
                        )}
                        <span className={`text-sm ${step.status === 'done' ? 'text-foreground' : step.status === 'failed' ? 'text-destructive' : step.status === 'running' ? 'text-primary' : 'text-muted-foreground'}`}>
                          {step.name}
                        </span>
                        {step.agent && (
                          <span className="text-xs px-2 py-0.5 rounded bg-primary/10 text-primary">{step.agent}</span>
                        )}
                        {step.tool && (
                          <span className="text-xs px-2 py-0.5 rounded bg-secondary/50 text-secondary-foreground">{step.tool}</span>
                        )}
                        {step.reason && <span className="w-full pl-8 text-xs text-muted-foreground">选择原因：{step.reason}</span>}
                        {step.result && <span className="w-full pl-8 text-xs text-muted-foreground">执行结果：{step.result}</span>}
                      </div>
                    ))}
                  </div>
                </div>
              </div>
              
              {/* 加载动画 */}
              <div className="flex gap-3 max-w-4xl mx-auto">
                <div className="w-8 h-8 rounded-lg bg-primary flex items-center justify-center">
                  <Loader2 className="w-4 h-4 text-white animate-spin" />
                </div>
                <div className="bg-surface-container px-4 py-2.5 rounded-2xl">
                  <div className="flex gap-1">
                    <span className="w-2 h-2 bg-muted-foreground/50 rounded-full animate-bounce" style={{ animationDelay: '0ms' }}></span>
                    <span className="w-2 h-2 bg-muted-foreground/50 rounded-full animate-bounce" style={{ animationDelay: '150ms' }}></span>
                    <span className="w-2 h-2 bg-muted-foreground/50 rounded-full animate-bounce" style={{ animationDelay: '300ms' }}></span>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* 输入区域 */}
        <div className="border-t border-border p-4">
          {connectionError && <div className="max-w-4xl mx-auto mb-3 rounded-lg bg-destructive/10 px-3 py-2 text-xs text-destructive">{connectionError}</div>}
          <div className="max-w-4xl mx-auto">
            {/* 附件预览 */}
            {attachments.length > 0 && (
              <div className="flex flex-wrap gap-2 mb-3">
                {attachments.map((file, index) => (
                  <div key={index} className="relative group">
                    {file.preview ? (
                      <div className="w-16 h-16 rounded-lg overflow-hidden border border-border">
                        <img src={file.preview} alt={file.name} className="w-full h-full object-cover" />
                      </div>
                    ) : (
                      <div className="w-16 h-16 rounded-lg bg-surface-container border border-border flex items-center justify-center">
                        <File className="w-6 h-6 text-muted-foreground" />
                      </div>
                    )}
                    <button
                      onClick={() => removeAttachment(index)}
                      className="absolute -top-1 -right-1 w-5 h-5 bg-destructive text-white rounded-full flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity"
                    >
                      <X className="w-3 h-3" />
                    </button>
                    <div className="text-xs text-muted-foreground truncate w-16 mt-1">{file.name}</div>
                  </div>
                ))}
              </div>
            )}
            
            {/* 输入框和按钮 */}
            <div className="flex gap-3">
              {/* 隐藏的文件输入 */}
              <input
                ref={fileInputRef}
                type="file"
                multiple
                onChange={handleFileUpload}
                className="hidden"
                accept=".txt,.md,.json,.js,.ts,.py,.rs,.cpp,.c,.java,.pdf,.doc,.docx"
              />
              <input
                ref={imageInputRef}
                type="file"
                multiple
                onChange={handleFileUpload}
                className="hidden"
                accept="image/png,image/jpeg,image/gif,image/webp,image/svg+xml"
              />
              
              {/* 上传按钮 */}
              <div className="flex gap-1">
                <button
                  onClick={() => imageInputRef.current?.click()}
                  className="p-2.5 text-muted-foreground hover:text-foreground hover:bg-surface-container rounded-xl transition-colors"
                  title="上传图片"
                >
                  <Image className="w-4 h-4" />
                </button>
                <button
                  onClick={() => fileInputRef.current?.click()}
                  className="p-2.5 text-muted-foreground hover:text-foreground hover:bg-surface-container rounded-xl transition-colors"
                  title="上传文件"
                >
                  <Paperclip className="w-4 h-4" />
                </button>
              </div>
              
              <input
                type="text"
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && !e.shiftKey && handleSend()}
                placeholder={selectedAgent ? `向${agents.find(a => a.id === selectedAgent)?.name}提问...` : '输入问题，智能路由将自动分发到合适的 Agent...'}
                className="flex-1 px-4 py-2.5 bg-surface-container border-none rounded-xl text-sm text-foreground placeholder:text-muted-foreground/40 focus:outline-none focus:ring-2 focus:ring-primary/30"
              />
              <button
                onClick={handleSend}
                disabled={(!input.trim() && attachments.length === 0) || isLoading}
                className="px-4 py-2.5 bg-primary text-primary-foreground rounded-xl hover:opacity-90 disabled:opacity-50 transition-opacity"
              >
                <Send className="w-4 h-4" />
              </button>
            </div>
            <div className="text-xs text-muted-foreground text-center mt-2">
              按 Enter 发送 · 支持上传图片/文件 · 支持智能路由自动分发 · SSE 流式输出
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}