'use client';

import { useEffect, useState, useCallback } from 'react';
import { apiFetch } from '@/lib/api';
import { track } from '@/lib/track';
import { Heart, ChevronLeft, ChevronRight, CheckCircle2, XCircle, Clock, History, Loader2 } from 'lucide-react';
import Editor from '@monaco-editor/react';

type Problem = {
  id: number;
  title: string;
  difficulty: string;
  passRate: number;
  submissions: number;
  description: string;
  examples: { input: string; output: string; explanation?: string }[];
  constraints: string[];
  tags: string[];
};

type CaseResult = {
  input: string;
  expected: string;
  actual: string;
  passed: boolean;
  error?: string;
  time_ms: number;
};

type RunResult = {
  status: string;
  stdout?: string;
  stderr?: string;
  execution_time_ms?: number;
  memory_kb?: number;
  passed_cases?: number;
  total_cases?: number;
  case_results?: CaseResult[];
};

type ProblemListItem = {
  id: number;
  title: string;
  difficulty: string;
  status?: string;
};

const difficultyColors: Record<string, string> = {
  easy: 'text-green-500', simple: 'text-green-500', 简单: 'text-green-500',
  medium: 'text-orange-500', 中等: 'text-orange-500',
  hard: 'text-red-500', 困难: 'text-red-500',
};

const languageMap: Record<string, string> = {
  python: 'python', javascript: 'javascript', cpp: 'cpp', rust: 'rust',
};

const codeKey = (pid: number, lang: string) => `codeforge_code_${pid}_${lang}`;

export default function PracticePage() {
  const [language, setLanguage] = useState('python');
  const [activeTab, setActiveTab] = useState<'result' | 'stdout' | 'info' | 'history'>('result');
  const [isFavorite, setIsFavorite] = useState(false);
  const [problem, setProblem] = useState<Problem | null>(null);
  const [code, setCode] = useState('');
  const [codeTemplates, setCodeTemplates] = useState<Record<string, string>>({});
  const [runResult, setRunResult] = useState<RunResult | null>(null);
  const [isRunning, setIsRunning] = useState(false);
  const [problemList, setProblemList] = useState<ProblemListItem[]>([]);
  const [showList, setShowList] = useState(false);
  const [submissions, setSubmissions] = useState<{ id: string; language: string; status: string; passed_cases: number; total_cases: number; created_at: string }[]>([]);

  const loadProblem = useCallback(async (pid: number) => {
    try {
      const [detail, templates] = await Promise.all([
        apiFetch<Record<string, unknown>>(`/api/v1/problems/${pid}`),
        apiFetch<Record<string, string>>(`/api/v1/problems/${pid}/templates`),
      ]);
      const p: Problem = {
        id: Number(detail.id), title: String(detail.title), difficulty: String(detail.difficulty),
        passRate: Number(detail.pass_rate), submissions: Number(detail.submissions),
        description: String(detail.description),
        examples: detail.examples as Problem['examples'] || [],
        constraints: detail.constraints as string[] || [],
        tags: detail.tags as string[] || [],
      };
      setProblem(p);
      const tpls = { ...templates };
      setCodeTemplates(tpls);
      // 恢复代码：localStorage 优先，否则模板
      const saved = typeof window !== 'undefined' ? localStorage.getItem(codeKey(pid, language)) : null;
      setCode(saved || tpls[language] || '');
    } catch { /* ignore */ }
  }, [language]);

  const loadSubmissions = useCallback(async (pid: number) => {
    try {
      const data = await apiFetch<{ items: typeof submissions }>(`/api/v1/problems/${pid}/submissions`);
      setSubmissions(data.items || []);
    } catch { /* ignore */ }
  }, []);

  useEffect(() => {
    apiFetch<{ items: ProblemListItem[] }>('/api/v1/problems?page_size=100').then(data => {
      setProblemList(data.items || []);
    }).catch(() => undefined);
    loadProblem(1);
    track('problem_start', { problem_id: 1 });
  }, []);

  useEffect(() => {
    if (problem) {
      loadProblem(problem.id);
      loadSubmissions(problem.id);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [language]);

  const changeLanguage = (lang: string) => {
    setLanguage(lang);
    if (problem) {
      const saved = typeof window !== 'undefined' ? localStorage.getItem(codeKey(problem.id, lang)) : null;
      setCode(saved || codeTemplates[lang] || '');
    }
  };

  const onCodeChange = (val: string | undefined) => {
    const v = val || '';
    setCode(v);
    if (problem && typeof window !== 'undefined') {
      localStorage.setItem(codeKey(problem.id, language), v);
    }
  };

  const execute = async (submit: boolean) => {
    if (!problem) return;
    setIsRunning(true);
    setActiveTab(submit ? 'result' : 'stdout');
    try {
      const result = await apiFetch<RunResult>(submit ? '/api/v1/code/submit' : '/api/v1/code/run', {
        method: 'POST', body: JSON.stringify({ problem_id: problem.id, language, code }),
      });
      setRunResult(result);
      if (submit) {
        setActiveTab('result');
        loadSubmissions(problem.id);
      }
    } catch (error) {
      setRunResult({ status: 'error', stderr: error instanceof Error ? error.message : '执行失败' });
    } finally {
      setIsRunning(false);
    }
  };

  const resetCode = () => {
    if (problem && codeTemplates[language]) {
      setCode(codeTemplates[language]);
      if (typeof window !== 'undefined') localStorage.setItem(codeKey(problem.id, language), codeTemplates[language]);
    }
  };

  const toggleFavorite = async () => {
    if (!problem) return;
    try {
      if (isFavorite) {
        await apiFetch(`/api/v1/favorites/${problem.id}`, { method: 'DELETE' });
        setIsFavorite(false);
      } else {
        await apiFetch('/api/v1/favorites', { method: 'POST', body: JSON.stringify({ course_id: problem.id }) });
        setIsFavorite(true);
      }
    } catch { /* ignore */ }
  };

  const switchProblem = (pid: number) => {
    setShowList(false);
    setRunResult(null);
    loadProblem(pid);
    loadSubmissions(pid);
    track('problem_start', { problem_id: pid });
  };

  const caseResults = runResult?.case_results || [];

  return (
    <>
      <div className="px-8 py-4 border-b border-border/10 flex items-center justify-between">
        <div className="flex items-center gap-4">
          <button onClick={() => setShowList(!showList)} className="px-3 py-1.5 text-sm text-muted-foreground hover:text-foreground hover:bg-surface-container rounded transition-colors">
            题目列表
          </button>
          <button className="p-1 text-muted-foreground hover:text-foreground" onClick={() => { const prev = problemList.find(p => p.id < (problem?.id || 0)); if (prev) switchProblem(prev.id); }} disabled={!problemList.some(p => p.id < (problem?.id || 0))}>
            <ChevronLeft className="w-5 h-5" />
          </button>
          <span className="text-sm font-medium text-foreground">{problem?.id}. {problem?.title}</span>
          <button className="p-1 text-muted-foreground hover:text-foreground" onClick={() => { const next = problemList.find(p => p.id > (problem?.id || 0)); if (next) switchProblem(next.id); }} disabled={!problemList.some(p => p.id > (problem?.id || 0))}>
            <ChevronRight className="w-5 h-5" />
          </button>
        </div>
        <div className="flex items-center gap-3">
          {problem && <span className={`text-xs font-medium ${difficultyColors[problem.difficulty.toLowerCase()] || 'text-muted-foreground'}`}>{problem.difficulty}</span>}
          <button onClick={toggleFavorite} className="p-1 text-muted-foreground hover:text-red-500">
            <Heart className={`w-4 h-4 ${isFavorite ? 'fill-red-500 text-red-500' : ''}`} />
          </button>
        </div>
      </div>

      {/* 题目列表抽屉 */}
      {showList && (
        <div className="absolute z-40 left-0 top-16 w-80 max-h-[60vh] overflow-y-auto bg-surface border border-border rounded-r-xl shadow-xl">
          {problemList.map(p => (
            <div key={p.id} onClick={() => switchProblem(p.id)} className={`px-4 py-3 cursor-pointer border-b border-border/50 hover:bg-surface-container ${p.id === problem?.id ? 'bg-primary/10' : ''}`}>
              <div className="flex items-center justify-between">
                <span className="text-sm text-foreground">{p.id}. {p.title}</span>
                {p.status === 'solved' && <CheckCircle2 className="w-4 h-4 text-green-500" />}
              </div>
              <span className={`text-xs ${difficultyColors[p.difficulty?.toLowerCase()] || 'text-muted-foreground'}`}>{p.difficulty}</span>
            </div>
          ))}
        </div>
      )}

      <div className="flex h-[calc(100vh-8rem)]">
        {/* 左侧题目描述 */}
        <div className="w-[40%] overflow-y-auto p-6 border-r border-border/10">
          {problem && (
            <>
              <div className="flex items-center gap-3 mb-4">
                <span className={`px-2 py-0.5 text-xs rounded ${difficultyColors[problem.difficulty.toLowerCase()] || 'text-muted-foreground'} bg-surface-container`}>{problem.difficulty}</span>
                <span className="text-xs text-muted-foreground">通过率 {(problem.passRate * 100).toFixed(1)}%</span>
                <span className="text-xs text-muted-foreground">提交 {problem.submissions.toLocaleString()}</span>
              </div>
              {problem.tags?.map((tag, i) => (
                <span key={i} className="mr-2 mb-2 inline-block px-2 py-0.5 text-xs bg-primary/10 text-primary rounded">{tag}</span>
              ))}
              <div className="mt-4 text-sm text-foreground leading-relaxed whitespace-pre-wrap">{problem.description}</div>
              {problem.examples?.map((ex, i) => (
                <div key={i} className="mt-4 p-4 bg-surface-container rounded-lg">
                  <div className="text-xs font-medium text-muted-foreground mb-2">示例 {i + 1}</div>
                  <div className="text-xs text-foreground"><span className="text-muted-foreground">输入：</span>{ex.input}</div>
                  <div className="text-xs text-foreground mt-1"><span className="text-muted-foreground">输出：</span>{ex.output}</div>
                  {ex.explanation && <div className="text-xs text-muted-foreground mt-1">{ex.explanation}</div>}
                </div>
              ))}
              {problem.constraints?.length > 0 && (
                <div className="mt-4">
                  <div className="text-xs font-medium text-muted-foreground mb-2">约束</div>
                  {problem.constraints.map((c, i) => <div key={i} className="text-xs text-muted-foreground">• {c}</div>)}
                </div>
              )}
            </>
          )}
        </div>

        {/* 右侧编辑器 + 输出 */}
        <div className="flex-1 flex flex-col">
          <div className="flex items-center justify-between px-4 py-2 border-b border-border/10">
            <select value={language} onChange={(e) => changeLanguage(e.target.value)} className="text-sm border border-border rounded px-2 py-1 bg-surface text-foreground">
              <option value="python">Python</option>
              <option value="javascript">JavaScript</option>
              <option value="cpp">C++</option>
              <option value="rust">Rust</option>
            </select>
            <div className="flex items-center gap-2">
              <button onClick={resetCode} className="px-3 py-1.5 text-xs text-muted-foreground hover:text-foreground hover:bg-surface-container rounded transition-colors">重置</button>
              <button disabled={isRunning} onClick={() => execute(false)} className="px-3 py-1.5 text-xs font-medium text-primary bg-primary/10 hover:bg-primary/20 rounded transition-colors disabled:opacity-50">
                {isRunning ? '运行中...' : '运行代码'}
              </button>
              <button disabled={isRunning} onClick={() => execute(true)} className="px-3 py-1.5 text-xs font-medium text-primary-foreground bg-primary hover:bg-primary/90 rounded transition-colors disabled:opacity-50">提交答案</button>
            </div>
          </div>

          {/* Monaco 编辑器 */}
          <div className="flex-1" style={{ minHeight: '300px' }}>
            <Editor
              language={languageMap[language] || 'plaintext'}
              value={code}
              onChange={onCodeChange}
              theme="vs-dark"
              options={{
                fontSize: 14, minimap: { enabled: false }, scrollBeyondLastLine: false,
                wordWrap: 'on', tabSize: 4, automaticLayout: true,
              }}
            />
          </div>

          {/* 输出区域 */}
          <div className="h-56 border-t border-border/10 flex flex-col">
            <div className="flex border-b border-border/10">
              {[
                { key: 'result', label: '测试结果', icon: runResult?.case_results?.length ? CheckCircle2 : null },
                { key: 'stdout', label: '输出' },
                { key: 'info', label: '执行详情' },
                { key: 'history', label: '提交历史' },
              ].map(tab => {
                const Icon = tab.icon as React.ComponentType<{ className?: string }> | null;
                return (
                  <button key={tab.key} onClick={() => setActiveTab(tab.key as typeof activeTab)} className={`px-4 py-2 text-xs font-medium transition-colors flex items-center gap-1 ${activeTab === tab.key ? 'text-primary border-b-2 border-primary' : 'text-muted-foreground hover:text-foreground'}`}>
                    {Icon && <Icon className="w-3 h-3" />}{tab.label}
                  </button>
                );
              })}
            </div>
            <div className="flex-1 overflow-auto p-4">
              {activeTab === 'result' && (
                <div className="space-y-2 text-sm">
                  {caseResults.length > 0 ? caseResults.map((tc, idx) => (
                    <div key={idx} className={`flex items-start gap-3 p-2 rounded ${tc.passed ? 'bg-green-500/5' : 'bg-red-500/5'}`}>
                      {tc.passed ? <CheckCircle2 className="w-4 h-4 text-green-500 flex-shrink-0 mt-0.5" /> : <XCircle className="w-4 h-4 text-red-500 flex-shrink-0 mt-0.5" />}
                      <div className="flex-1 min-w-0">
                        <div className="text-xs text-muted-foreground">输入: {tc.input}</div>
                        <div className="text-xs text-foreground">期望: {tc.expected}</div>
                        <div className={`text-xs ${tc.passed ? 'text-green-600' : 'text-red-600'}`}>实际: {tc.actual || '(空)'}</div>
                        {tc.error && <div className="text-xs text-red-500 mt-1">错误: {tc.error}</div>}
                        <div className="text-[10px] text-muted-foreground mt-0.5"><Clock className="w-3 h-3 inline" /> {tc.time_ms}ms</div>
                      </div>
                    </div>
                  )) : <div className="text-muted-foreground text-xs">提交代码后查看逐条测试结果</div>}
                </div>
              )}
              {activeTab === 'stdout' && (
                <pre className="text-xs text-muted-foreground font-mono whitespace-pre-wrap">{runResult?.stdout || runResult?.stderr || '点击运行代码查看输出。'}</pre>
              )}
              {activeTab === 'info' && (
                <div className="space-y-2 text-sm">
                  <p className={`${runResult?.status === 'accepted' || runResult?.status === 'success' ? 'text-green-500 font-medium' : 'text-red-500 font-medium'}`}>{runResult ? `状态：${runResult.status}` : '尚未提交'}</p>
                  {runResult?.execution_time_ms !== undefined && <p className="text-muted-foreground text-xs">执行时间: {runResult.execution_time_ms}ms</p>}
                  {runResult?.passed_cases !== undefined && <p className="text-muted-foreground text-xs">通过用例: {runResult.passed_cases}/{runResult.total_cases}</p>}
                </div>
              )}
              {activeTab === 'history' && (
                <div className="space-y-1 text-sm">
                  {submissions.length > 0 ? submissions.map((sub) => (
                    <div key={sub.id} className="flex items-center gap-3 p-2 bg-surface-container rounded">
                      {sub.status === 'accepted' ? <CheckCircle2 className="w-4 h-4 text-green-500" /> : <XCircle className="w-4 h-4 text-red-500" />}
                      <span className="text-xs text-foreground">{sub.language}</span>
                      <span className={`text-xs ${sub.status === 'accepted' ? 'text-green-500' : 'text-red-500'}`}>{sub.status}</span>
                      <span className="text-xs text-muted-foreground">{sub.passed_cases}/{sub.total_cases}</span>
                      <span className="text-xs text-muted-foreground ml-auto">{new Date(sub.created_at).toLocaleString('zh-CN')}</span>
                    </div>
                  )) : <div className="text-muted-foreground text-xs">暂无提交记录</div>}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </>
  );
}