'use client';

import { useState, useEffect, useCallback } from 'react';
import { apiFetch } from '@/lib/api';
import { track } from '@/lib/track';
import { Clock, Target, Trophy, Code, CheckCircle, XCircle, Lightbulb, Play, RotateCcw, Loader2, ChevronLeft } from 'lucide-react';
import Editor from '@monaco-editor/react';

type ExamQuestion = {
  id: string;
  type: string;
  category: string;
  difficulty: string;
  title: string;
  description: string;
  constraints: string[];
  timeLimit: number;
  score: number;
  example?: { input: string; output: string; explanation?: string };
};

type FeedbackItem = {
  question_id: string;
  title: string;
  type: string;
  score: number;
  max_score: number;
  feedback: string;
};

type ExamRecord = {
  exam_id: string;
  date: string;
  score: number;
  direction: string;
  difficulty: string;
};

const codeTemplates: Record<string, string> = {
  python: '# 在这里写你的代码\n',
  javascript: '// 在这里写你的代码\n',
  cpp: '// 在这里写你的代码\n',
  rust: '// 在这里写你的代码\n',
};

const difficultyColors: Record<string, string> = {
  easy: 'bg-green-100 text-green-700', simple: 'bg-green-100 text-green-700', 简单: 'bg-green-100 text-green-700', 初级: 'bg-green-100 text-green-700',
  medium: 'bg-orange-100 text-orange-700', 中等: 'bg-orange-100 text-orange-700', 中级: 'bg-orange-100 text-orange-700',
  hard: 'bg-red-100 text-red-700', 困难: 'bg-red-100 text-red-700', 高级: 'bg-red-100 text-red-700',
};

export default function ExamPage() {
  const [isStarted, setIsStarted] = useState(false);
  const [timeLeft, setTimeLeft] = useState(0);
  const [answers, setAnswers] = useState<Record<string, { answer: string; language: string }>>({});
  const [currentQuestion, setCurrentQuestion] = useState(0);
  const [language, setLanguage] = useState('python');
  const [examQuestions, setExamQuestions] = useState<ExamQuestion[]>([]);
  const [examId, setExamId] = useState('');
  const [direction, setDirection] = useState('Python 后端开发');
  const [difficulty, setDifficulty] = useState('中等');
  const [questionCount, setQuestionCount] = useState(4);
  const [isGenerating, setIsGenerating] = useState(false);
  const [runResult, setRunResult] = useState<{ status: string; stdout?: string; stderr?: string; case_results?: { input: string; expected: string; actual: string; passed: boolean }[] } | null>(null);
  const [isRunning, setIsRunning] = useState(false);
  const [examResult, setExamResult] = useState<{ score: number; feedback: FeedbackItem[] } | null>(null);
  const [examHistory, setExamHistory] = useState<ExamRecord[]>([]);
  const [showHistory, setShowHistory] = useState(false);

  const question = examQuestions[currentQuestion];
  const isCodeQuestion = question?.type === 'code';

  // 倒计时
  useEffect(() => {
    if (!isStarted || timeLeft <= 0) return;
    const timer = setInterval(() => {
      setTimeLeft(prev => {
        if (prev <= 1) {
          clearInterval(timer);
          submitExam();
          return 0;
        }
        return prev - 1;
      });
    }, 1000);
    return () => clearInterval(timer);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isStarted, currentQuestion]);

  // 加载笔试历史
  const loadHistory = useCallback(() => {
    apiFetch<{ items: ExamRecord[] }>('/api/v1/interview/exams').then(data => {
      setExamHistory(data.items || []);
    }).catch(() => undefined);
  }, []);

  useEffect(() => { loadHistory(); }, [loadHistory]);

  const startExam = async () => {
    setIsGenerating(true);
    setRunResult(null);
    setExamResult(null);
    try {
      const qMap: Record<string, number> = { '3 题（快速模拟）': 3, '5 题（标准模拟）': 5, '8 题（完整模拟）': 8 };
      const qc = qMap[questionCount.toString()] || questionCount;
      const data = await apiFetch<{ exam_id: string; questions: ExamQuestion[] }>('/api/v1/interview/exams/generate', {
        method: 'POST', body: JSON.stringify({ direction, difficulty, question_count: qc }),
      });
      setExamId(data.exam_id);
      setExamQuestions(data.questions || []);
      setAnswers({});
      setCurrentQuestion(0);
      setIsStarted(true);
      setTimeLeft((data.questions?.[0]?.timeLimit || 20) * 60);
      track('interview_start', { direction, difficulty, question_count: data.questions?.length || qc });
    } catch { /* ignore */ } finally { setIsGenerating(false); }
  };

  const getCurrentAnswer = () => answers[question?.id]?.answer || '';
  const setCurrentAnswer = (val: string) => {
    if (!question) return;
    setAnswers(prev => ({ ...prev, [question.id]: { answer: val, language } }));
  };

  const runCode = async () => {
    if (!question) return;
    setIsRunning(true);
    setRunResult(null);
    try {
      const result = await apiFetch<{ status: string; stdout?: string; stderr?: string; case_results?: { input: string; expected: string; actual: string; passed: boolean }[] }>('/api/v1/code/run', {
        method: 'POST', body: JSON.stringify({ language, code: getCurrentAnswer() }),
      });
      setRunResult(result);
    } catch (e) {
      setRunResult({ status: 'error', stderr: e instanceof Error ? e.message : '运行失败' });
    } finally { setIsRunning(false); }
  };

  const submitExam = async () => {
    if (!examId) return;
    try {
      const result = await apiFetch<{ score: number; feedback: FeedbackItem[] }>(`/api/v1/interview/exams/${examId}/submit`, {
        method: 'POST', body: JSON.stringify({
          answers: examQuestions.map(q => ({
            question_id: q.id,
            answer: answers[q.id]?.answer || '',
            language: answers[q.id]?.language || language,
          })),
        }),
      });
      setExamResult(result);
      setIsStarted(false);
      loadHistory();
    } catch { /* ignore */ }
  };

  const nextQuestion = () => {
    if (currentQuestion < examQuestions.length - 1) {
      setCurrentQuestion(currentQuestion + 1);
      setTimeLeft(examQuestions[currentQuestion + 1].timeLimit * 60);
      setRunResult(null);
    } else {
      submitExam();
    }
  };

  const getHint = async () => {
    if (!question) return;
    try {
      const response = await fetch('/api/v1/agent/chat', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ message: `给我一个提示：${question.title} - ${question.description}`, collaboration_mode: 'dynamic', context: { current_page: '/interview' } }),
      });
      if (response.ok) {
        const reader = response.body?.getReader();
        const decoder = new TextDecoder();
        let buffer = '';
        let hint = '';
        while (reader) {
          const { value, done } = await reader.read();
          if (done) break;
          buffer += decoder.decode(value || new Uint8Array(), { stream: true });
          const blocks = buffer.split('\n\n');
          buffer = blocks.pop() || '';
          for (const block of blocks) {
            const dataLine = block.split('\n').find(l => l.startsWith('data:'))?.slice(5).trim();
            if (!dataLine) continue;
            const data = JSON.parse(dataLine);
            if (data.content) hint += data.content;
          }
        }
        setRunResult({ status: 'hint', stdout: hint || '无法获取提示' });
      }
    } catch { /* ignore */ }
  };

  // 结果页
  if (examResult) {
    return (
      <div className="max-w-4xl mx-auto py-8">
        <div className="text-center mb-8">
          <Trophy className="w-16 h-16 text-primary mx-auto mb-4" />
          <h1 className="text-3xl font-bold text-foreground">笔试完成</h1>
          <div className="text-5xl font-bold text-primary mt-4">{examResult.score}<span className="text-2xl text-muted-foreground">分</span></div>
        </div>
        <div className="space-y-4">
          {examResult.feedback.map((fb, i) => (
            <div key={i} className="p-4 bg-surface border border-border rounded-xl">
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  {fb.type === 'code' ? <Code className="w-4 h-4 text-primary" /> : <Target className="w-4 h-4 text-primary" />}
                  <span className="font-medium text-foreground">{fb.title}</span>
                </div>
                <span className={`text-lg font-bold ${fb.score >= 80 ? 'text-green-500' : fb.score >= 60 ? 'text-orange-500' : 'text-red-500'}`}>
                  {fb.score}<span className="text-sm text-muted-foreground">/{fb.max_score}</span>
                </span>
              </div>
              <div className="flex items-center gap-2 mb-1">
                <div className="flex-1 h-2 bg-surface-container rounded-full overflow-hidden">
                  <div className={`h-full rounded-full ${fb.score >= 80 ? 'bg-green-500' : fb.score >= 60 ? 'bg-orange-500' : 'bg-red-500'}`} style={{ width: `${(fb.score / fb.max_score) * 100}%` }} />
                </div>
              </div>
              <p className="text-sm text-muted-foreground">{fb.feedback}</p>
            </div>
          ))}
        </div>
        <div className="flex gap-4 mt-8 justify-center">
          <button onClick={() => { setExamResult(null); loadHistory(); }} className="px-6 py-2.5 bg-surface-container text-foreground rounded-xl hover:opacity-80 text-sm font-medium">查看历史</button>
          <button onClick={startExam} className="px-6 py-2.5 bg-primary text-primary-foreground rounded-xl hover:opacity-90 text-sm font-medium">再来一次</button>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-6xl mx-auto">
      <div className="mb-8 flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-foreground">笔试模拟</h1>
          <p className="text-muted-foreground mt-2">AI 出题、自动评分、代码题支持在线编程</p>
        </div>
        {!isStarted && (
          <button onClick={() => { setShowHistory(!showHistory); loadHistory(); }} className="px-4 py-2 text-sm text-muted-foreground hover:text-foreground hover:bg-surface-container rounded-xl flex items-center gap-2">
            <Trophy className="w-4 h-4" /> 笔试历史
          </button>
        )}
      </div>

      {/* 历史记录抽屉 */}
      {!isStarted && showHistory && (
        <div className="mb-6 p-4 bg-surface border border-border rounded-xl">
          <h3 className="font-medium text-foreground mb-3">笔试历史</h3>
          {examHistory.length > 0 ? (
            <div className="space-y-2">
              {examHistory.map((record, i) => (
                <div key={i} className="flex items-center justify-between p-3 bg-surface-container rounded-lg">
                  <div>
                    <div className="text-sm font-medium text-foreground">{record.direction}</div>
                    <div className="text-xs text-muted-foreground">{record.date} · {record.difficulty}</div>
                  </div>
                  <div className={`text-lg font-bold ${record.score >= 80 ? 'text-green-500' : record.score >= 60 ? 'text-orange-500' : 'text-red-500'}`}>{record.score}</div>
                </div>
              ))}
            </div>
          ) : <div className="text-sm text-muted-foreground text-center py-4">暂无笔试记录</div>}
        </div>
      )}

      {!isStarted ? (
        /* 笔试设置 */
        <div className="max-w-lg mx-auto p-8 bg-surface border border-border rounded-xl">
          <h3 className="font-semibold text-foreground mb-6 flex items-center gap-2"><Target className="w-5 h-5 text-primary" /> 笔试设置</h3>
          <div className="space-y-4">
            <div>
              <label className="text-sm text-muted-foreground">笔试方向</label>
              <select value={direction} onChange={e => setDirection(e.target.value)} className="w-full mt-1 px-3 py-2 bg-surface-container border-none rounded-lg text-sm text-foreground">
                <option>Python 后端开发</option>
                <option>C++ 系统开发</option>
                <option>算法与数据结构</option>
                <option>AI Agent 开发</option>
                <option>全栈综合</option>
              </select>
            </div>
            <div>
              <label className="text-sm text-muted-foreground">难度级别</label>
              <select value={difficulty} onChange={e => setDifficulty(e.target.value)} className="w-full mt-1 px-3 py-2 bg-surface-container border-none rounded-lg text-sm text-foreground">
                <option>简单</option>
                <option>中等</option>
                <option>困难</option>
              </select>
            </div>
            <div>
              <label className="text-sm text-muted-foreground">题目数量</label>
              <select value={questionCount} onChange={e => setQuestionCount(Number(e.target.value))} className="w-full mt-1 px-3 py-2 bg-surface-container border-none rounded-lg text-sm text-foreground">
                <option value={3}>3 题（快速模拟）</option>
                <option value={4}>4 题</option>
                <option value={5}>5 题（标准模拟）</option>
                <option value={8}>8 题（完整模拟）</option>
              </select>
            </div>
          </div>
          <button onClick={startExam} disabled={isGenerating} className="w-full mt-6 py-3 bg-primary text-primary-foreground rounded-xl font-medium hover:opacity-90 transition-opacity disabled:opacity-50 flex items-center justify-center gap-2">
            {isGenerating ? <><Loader2 className="w-4 h-4 animate-spin" /> AI 出题中...</> : <><Play className="w-4 h-4" /> 开始笔试</>}
          </button>
        </div>
      ) : (
        /* 笔试进行中 */
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* 左侧：题目 */}
          <div className="space-y-4">
            <div className="bg-surface border border-border rounded-xl p-6">
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-2">
                  <span className="px-2 py-1 text-xs font-medium bg-primary/10 text-primary rounded">第 {currentQuestion + 1} / {examQuestions.length} 题</span>
                  <span className={`px-2 py-1 text-xs rounded ${difficultyColors[question?.difficulty] || 'bg-surface-container'}`}>{question?.difficulty}</span>
                  <span className="px-2 py-1 text-xs bg-surface-container text-muted-foreground rounded">{question?.category}</span>
                  {isCodeQuestion && <span className="px-2 py-1 text-xs bg-purple-100 text-purple-700 rounded flex items-center gap-1"><Code className="w-3 h-3" /> 代码题</span>}
                </div>
                <div className={`flex items-center gap-1 text-sm ${timeLeft < 60 ? 'text-red-500' : 'text-muted-foreground'}`}>
                  <Clock className="w-4 h-4" /> {Math.floor(timeLeft / 60)}:{(timeLeft % 60).toString().padStart(2, '0')}
                </div>
              </div>
              <h2 className="text-xl font-bold text-foreground mb-3">{question?.title}</h2>
              <p className="text-sm text-foreground mb-4 whitespace-pre-wrap">{question?.description}</p>
              {isCodeQuestion && question?.example && (
                <div className="mb-4 p-4 bg-surface-container rounded-lg">
                  <h4 className="text-xs font-medium text-muted-foreground mb-2">示例</h4>
                  <div className="text-xs space-y-1">
                    <div><span className="text-muted-foreground">输入：</span><code className="text-foreground">{question.example.input}</code></div>
                    <div><span className="text-muted-foreground">输出：</span><code className="text-foreground">{question.example.output}</code></div>
                    {question.example.explanation && <div><span className="text-muted-foreground">解释：</span>{question.example.explanation}</div>}
                  </div>
                </div>
              )}
              {question?.constraints && question.constraints.length > 0 && (
                <div>
                  <h4 className="text-xs font-medium text-muted-foreground mb-1">约束</h4>
                  {question.constraints.map((c, i) => <div key={i} className="text-xs text-muted-foreground">• {c}</div>)}
                </div>
              )}
            </div>

            {/* 题目导航 */}
            <div className="bg-surface border border-border rounded-xl p-4">
              <div className="flex flex-wrap gap-2">
                {examQuestions.map((q, i) => (
                  <button key={i} onClick={() => { setCurrentQuestion(i); setTimeLeft(q.timeLimit * 60); setRunResult(null); }} className={`w-10 h-10 rounded-lg text-sm font-medium flex items-center justify-center transition-colors ${i === currentQuestion ? 'bg-primary text-primary-foreground' : answers[q.id]?.answer ? 'bg-green-100 text-green-700' : 'bg-surface-container text-muted-foreground hover:bg-primary/10'}`}>{i + 1}</button>
                ))}
              </div>
            </div>
          </div>

          {/* 右侧：答题区 */}
          <div className="bg-surface border border-border rounded-xl overflow-hidden flex flex-col">
            {isCodeQuestion ? (
              <>
                <div className="flex items-center justify-between px-4 py-2 bg-[#1e1e2e] border-b border-[#313244]">
                  <span className="text-sm text-[#cdd6f4]">代码编辑器</span>
                  <select value={language} onChange={e => setLanguage(e.target.value)} className="px-2 py-1 text-xs bg-[#313244] text-[#cdd6f4] border-none rounded">
                    <option value="python">Python</option>
                    <option value="javascript">JavaScript</option>
                    <option value="cpp">C++</option>
                    <option value="rust">Rust</option>
                  </select>
                </div>
                <div className="flex-1" style={{ minHeight: '250px' }}>
                  <Editor language={language === 'cpp' ? 'cpp' : language === 'javascript' ? 'javascript' : language} value={getCurrentAnswer() || codeTemplates[language]} onChange={(v) => setCurrentAnswer(v || '')} theme="vs-dark" options={{ fontSize: 14, minimap: { enabled: false }, scrollBeyondLastLine: false, wordWrap: 'on', tabSize: 4, automaticLayout: true }} />
                </div>
                <div className="flex items-center justify-between px-4 py-3 bg-[#1e1e2e] border-t border-[#313244]">
                  <div className="flex items-center gap-2">
                    <button onClick={runCode} disabled={isRunning} className="px-3 py-1.5 text-xs bg-[#313244] text-[#cdd6f4] rounded hover:bg-[#45475a] flex items-center gap-1.5 disabled:opacity-50">
                      {isRunning ? <Loader2 className="w-3 h-3 animate-spin" /> : <Play className="w-3 h-3" />} 运行
                    </button>
                    <button onClick={() => setCurrentAnswer(codeTemplates[language])} className="px-3 py-1.5 text-xs bg-[#313244] text-[#cdd6f4] rounded hover:bg-[#45475a] flex items-center gap-1.5"><RotateCcw className="w-3 h-3" /> 重置</button>
                    <button onClick={getHint} className="px-3 py-1.5 text-xs bg-[#313244] text-[#cdd6f4] rounded hover:bg-[#45475a] flex items-center gap-1.5"><Lightbulb className="w-3 h-3" /> 提示</button>
                  </div>
                  <button onClick={nextQuestion} className="px-4 py-1.5 text-xs bg-green-600 text-white rounded hover:bg-green-700 flex items-center gap-1.5">
                    <CheckCircle className="w-3 h-3" /> {currentQuestion < examQuestions.length - 1 ? '下一题' : '提交'}
                  </button>
                </div>
                {/* 运行结果 */}
                {runResult && (
                  <div className="p-3 bg-[#1e1e2e] border-t border-[#313244] max-h-40 overflow-auto">
                    {runResult.status === 'hint' ? (
                      <div className="text-xs text-[#cdd6f4] whitespace-pre-wrap">{runResult.stdout}</div>
                    ) : runResult.case_results && runResult.case_results.length > 0 ? (
                      <div className="space-y-1">
                        {runResult.case_results.map((tc, i) => (
                          <div key={i} className={`flex items-center gap-2 text-xs ${tc.passed ? 'text-green-400' : 'text-red-400'}`}>
                            {tc.passed ? <CheckCircle className="w-3 h-3" /> : <XCircle className="w-3 h-3" />}
                            <span>期望: {tc.expected}</span>
                            <span>实际: {tc.actual || '(空)'}</span>
                          </div>
                        ))}
                      </div>
                    ) : (
                      <pre className="text-xs text-[#cdd6f4] whitespace-pre-wrap">{runResult.stdout || runResult.stderr || '无输出'}</pre>
                    )}
                  </div>
                )}
              </>
            ) : (
              <div className="p-6 flex flex-col h-full">
                <label className="text-sm font-medium text-foreground mb-3">你的回答：</label>
                <textarea value={getCurrentAnswer()} onChange={e => setCurrentAnswer(e.target.value)} placeholder="在这里输入你的答案..." className="flex-1 min-h-[200px] px-4 py-3 bg-surface-container border-none rounded-xl text-sm text-foreground placeholder:text-muted-foreground/40 focus:outline-none focus:ring-2 focus:ring-primary/30 resize-none" />
                <div className="flex justify-between items-center mt-4">
                  <button onClick={getHint} className="px-4 py-2 text-sm text-muted-foreground hover:text-foreground flex items-center gap-2"><Lightbulb className="w-4 h-4" /> 提示</button>
                  <button onClick={nextQuestion} className="px-6 py-2 bg-primary text-primary-foreground rounded-xl text-sm font-medium hover:opacity-90 flex items-center gap-2">
                    {currentQuestion < examQuestions.length - 1 ? '下一题' : '提交笔试'} <CheckCircle className="w-4 h-4" />
                  </button>
                </div>
                {runResult?.status === 'hint' && (
                  <div className="mt-3 p-3 bg-surface-container rounded-lg text-xs text-muted-foreground whitespace-pre-wrap">{runResult.stdout}</div>
                )}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}