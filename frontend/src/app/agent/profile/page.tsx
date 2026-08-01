'use client';

import { Brain, Target, TrendingUp, Award, BookOpen, Code, Database, Sparkles, Network, Edit3, X, Flame, Clock, CheckCircle2 } from 'lucide-react';
import { useEffect, useState, useCallback } from 'react';
import { apiFetch } from '@/lib/api';
import KnowledgeGraph from '@/components/profile/KnowledgeGraph';
import InsightsCards, { type Insights } from '@/components/profile/InsightsCards';
import {
  Bar, BarChart, CartesianGrid, PolarAngleAxis, PolarGrid, PolarRadiusAxis,
  Radar, RadarChart, ResponsiveContainer, Tooltip, XAxis, YAxis,
} from 'recharts';

type Profile = {
  level: string;
  computed_level: number;
  level_title: string;
  focusAreas: string[];
  weakAreas: string[];
  learningStyle: string;
  preferredDifficulty: string;
  preferred_time_slot: string;
  dailyGoal: number;
  totalStudyTime: number;
  streak: number;
  sessionCount: number;
  problemSolvedCount: number;
  problemAccuracy: number;
  lastActiveAt: string;
};

type KnowledgeState = {
  skill_name: string;
  category: string;
  mastery: number;
  attempts: number;
  correct_count: number;
  last_practiced_at: string;
};

type Dashboard = {
  heatmap: { date: string; minutes: number; sessions: number }[];
  radar: { category: string; mastery: number }[];
  trend: { date: string; minutes: number }[];
  stats: { total_minutes: number; total_sessions: number; total_problems: number; avg_accuracy: number };
};

const timeSlotLabels: Record<string, string> = {
  morning: '上午', afternoon: '下午', evening: '晚间', night: '深夜',
};

const categoryLabels: Record<string, string> = {
  python: 'Python', cpp: 'C++', database: '数据库', algorithm: '算法', agent: 'AI Agent',
};

function masteryColor(m: number) {
  if (m >= 0.85) return 'bg-green-500';
  if (m >= 0.6) return 'bg-blue-500';
  if (m >= 0.3) return 'bg-orange-500';
  return 'bg-red-500';
}

function masteryText(m: number) {
  if (m >= 0.85) return '精通';
  if (m >= 0.6) return '熟练';
  if (m >= 0.3) return '学习中';
  return '入门';
}

export default function AgentProfilePage() {
  const [profile, setProfile] = useState<Profile | null>(null);
  const [knowledgeStates, setKnowledgeStates] = useState<KnowledgeState[]>([]);
  const [dashboard, setDashboard] = useState<Dashboard | null>(null);
  const [editing, setEditing] = useState(false);
  const [editForm, setEditForm] = useState({ learning_style: '', preferred_difficulty: '', preferred_time_slot: '', daily_goal: 30, focus_areas: '', weak_areas: '' });

  const [loading, setLoading] = useState(true);
  const [insights, setInsights] = useState<Insights | null>(null);

  const loadData = useCallback(() => {
    setLoading(true);
    const withTimeout = <T,>(p: Promise<T>, ms = 8000): Promise<T> =>
      Promise.race([p, new Promise<T>((_, rej) => setTimeout(() => rej(new Error('timeout')), ms))]);

    let done = 0;
    const onDone = () => { done++; if (done >= 4) setLoading(false); };

    withTimeout(apiFetch<Profile>('/api/v1/agent/profile'))
      .then(p => setProfile(p))
      .catch(() => undefined)
      .finally(onDone);

    withTimeout(apiFetch<{ items: KnowledgeState[] }>('/api/v1/agent/knowledge-states'))
      .then(ks => setKnowledgeStates(ks.items || []))
      .catch(() => undefined)
      .finally(onDone);

    withTimeout(apiFetch<Dashboard>('/api/v1/agent/dashboard'))
      .then(d => setDashboard(d))
      .catch(() => undefined)
      .finally(onDone);

    withTimeout(apiFetch<Insights>('/api/v1/agent/insights'))
      .then(i => setInsights(i))
      .catch(() => undefined)
      .finally(onDone);
  }, []);

  useEffect(() => { loadData(); }, [loadData]);

  const startEdit = () => {
    if (!profile) return;
    setEditForm({
      learning_style: profile.learningStyle || '',
      preferred_difficulty: profile.preferredDifficulty || '',
      preferred_time_slot: profile.preferred_time_slot || '',
      daily_goal: profile.dailyGoal || 30,
      focus_areas: profile.focusAreas?.join(', ') || '',
      weak_areas: profile.weakAreas?.join(', ') || '',
    });
    setEditing(true);
  };

  const saveEdit = async () => {
    await apiFetch('/api/v1/agent/profile', {
      method: 'PUT',
      body: JSON.stringify({
        learning_style: editForm.learning_style || undefined,
        preferred_difficulty: editForm.preferred_difficulty || undefined,
        preferred_time_slot: editForm.preferred_time_slot || undefined,
        daily_goal: editForm.daily_goal || undefined,
        focus_areas: editForm.focus_areas ? editForm.focus_areas.split(',').map(s => s.trim()).filter(Boolean) : undefined,
        weak_areas: editForm.weak_areas ? editForm.weak_areas.split(',').map(s => s.trim()).filter(Boolean) : undefined,
      }),
    });
    setEditing(false);
    loadData();
  };

  if (loading && !profile) {
    return (
      <div className="flex flex-col items-center justify-center h-96 gap-4">
        <div className="animate-spin w-8 h-8 border-4 border-primary border-t-transparent rounded-full" />
        <div className="text-sm text-muted-foreground">正在加载学习画像...</div>
      </div>
    );
  }

  if (!profile) {
    return (
      <div className="flex flex-col items-center justify-center h-96 gap-4">
        <div className="text-sm text-muted-foreground">无法加载画像数据</div>
        <button onClick={loadData} className="px-4 py-2 bg-primary text-primary-foreground rounded-xl text-sm font-medium hover:opacity-90">重新加载</button>
      </div>
    );
  }

  // 知识图谱节点：按 category 分组，取 mastery 最高的几个
  const topSkills = [...knowledgeStates].sort((a, b) => b.mastery - a.mastery).slice(0, 12);

  return (
    <div className="max-w-6xl mx-auto pb-8">
      {loading && profile && (
        <div className="mb-4 flex items-center gap-2 px-4 py-2 bg-primary/5 border border-primary/10 rounded-lg">
          <div className="animate-spin w-4 h-4 border-2 border-primary border-t-transparent rounded-full" />
          <span className="text-xs text-muted-foreground">正在加载更多数据...</span>
        </div>
      )}
      {/* 页面标题 */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold text-foreground">学习画像</h1>
          <p className="text-muted-foreground mt-2">基于学习行为动态生成的个性化画像，驱动智能推荐和路由优化</p>
        </div>
        <button onClick={startEdit} className="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-xl hover:opacity-90 transition-opacity text-sm font-medium">
          <Edit3 className="w-4 h-4" /> 编辑画像
        </button>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* 左侧：基本信息 */}
        <div className="lg:col-span-1 space-y-6">
          {/* 等级卡片 */}
          <div className="p-6 bg-surface border border-border rounded-xl">
            <div className="flex items-center gap-3 mb-4">
              <div className="w-12 h-12 rounded-full bg-primary/10 flex items-center justify-center">
                <Brain className="w-6 h-6 text-primary" />
              </div>
              <div className="flex-1">
                <h3 className="font-semibold text-foreground">{profile.level}</h3>
                <p className="text-sm text-muted-foreground">Lv.{profile.computed_level} {profile.level_title}</p>
              </div>
            </div>
            {/* 经验条 */}
            <div className="mb-4">
              <div className="flex justify-between text-xs text-muted-foreground mb-1">
                <span>等级进度</span>
                <span>Lv.{profile.computed_level}</span>
              </div>
              <div className="h-2 bg-surface-container rounded-full overflow-hidden">
                <div className="h-full bg-gradient-to-r from-primary to-primary/60 rounded-full transition-all" style={{ width: `${Math.min(100, (profile.computed_level / 10) * 100)}%` }} />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-4 text-center">
              <div>
                <div className="text-2xl font-bold text-foreground">{profile.totalStudyTime}h</div>
                <div className="text-xs text-muted-foreground">总学习时长</div>
              </div>
              <div>
                <div className="text-2xl font-bold text-foreground flex items-center justify-center gap-1">
                  <Flame className="w-5 h-5 text-orange-500" />{profile.streak}
                </div>
                <div className="text-xs text-muted-foreground">连续打卡</div>
              </div>
              <div>
                <div className="text-2xl font-bold text-foreground">{profile.sessionCount}</div>
                <div className="text-xs text-muted-foreground">AI 对话次数</div>
              </div>
              <div>
                <div className="text-2xl font-bold text-foreground">{profile.problemSolvedCount}</div>
                <div className="text-xs text-muted-foreground">解题总数</div>
              </div>
            </div>
          </div>

          {/* 学习偏好 */}
          <div className="p-6 bg-surface border border-border rounded-xl">
            <h3 className="font-semibold text-foreground mb-4 flex items-center gap-2">
              <Target className="w-4 h-4 text-primary" /> 学习偏好
            </h3>
            <div className="space-y-3">
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">学习方式</span>
                <span className="text-foreground font-medium">{profile.learningStyle || '未设置'}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">偏好难度</span>
                <span className="text-foreground font-medium">{profile.preferredDifficulty || '未设置'}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">偏好时段</span>
                <span className="text-foreground font-medium">{timeSlotLabels[profile.preferred_time_slot] || profile.preferred_time_slot || '未设置'}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">每日目标</span>
                <span className="text-foreground font-medium">{profile.dailyGoal} 分钟</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-muted-foreground">正确率</span>
                <span className="text-foreground font-medium">{(profile.problemAccuracy * 100).toFixed(0)}%</span>
              </div>
              {profile.lastActiveAt && (
                <div className="flex justify-between text-sm">
                  <span className="text-muted-foreground">最后活跃</span>
                  <span className="text-foreground font-medium">{new Date(profile.lastActiveAt).toLocaleDateString('zh-CN')}</span>
                </div>
              )}
            </div>
          </div>

          {/* 强弱项 */}
          <div className="grid grid-cols-2 gap-4">
            <div className="p-4 bg-green-500/5 border border-green-500/20 rounded-xl">
              <h4 className="text-sm font-medium text-green-700 dark:text-green-400 mb-2">擅长领域</h4>
              <div className="flex flex-wrap gap-2">
                {profile.focusAreas?.length ? profile.focusAreas.map((area, i) => (
                  <span key={i} className="px-2 py-1 text-xs bg-green-500/10 text-green-700 dark:text-green-400 rounded">{area}</span>
                )) : <span className="text-xs text-muted-foreground">暂无数据</span>}
              </div>
            </div>
            <div className="p-4 bg-orange-500/5 border border-orange-500/20 rounded-xl">
              <h4 className="text-sm font-medium text-orange-700 dark:text-orange-400 mb-2">待加强</h4>
              <div className="flex flex-wrap gap-2">
                {profile.weakAreas?.length ? profile.weakAreas.map((area, i) => (
                  <span key={i} className="px-2 py-1 text-xs bg-orange-500/10 text-orange-700 dark:text-orange-400 rounded">{area}</span>
                )) : <span className="text-xs text-muted-foreground">暂无数据</span>}
              </div>
            </div>
          </div>
        </div>

        {/* 右侧：图表和数据 */}
        <div className="lg:col-span-2 space-y-6">
          {/* 知识图谱（ECharts 关系图） */}
          <div className="p-6 bg-surface border border-border rounded-xl">
            <h3 className="font-semibold text-foreground mb-4 flex items-center gap-2">
              <Network className="w-4 h-4 text-primary" /> 知识掌握图谱
            </h3>
            <KnowledgeGraph nodes={topSkills.map(ks => ({ id: ks.skill_name, name: ks.skill_name, category: ks.category, mastery: ks.mastery, attempts: ks.attempts }))} />
          </div>

          {/* 雷达图：按方向掌握度 */}
          {dashboard && dashboard.radar.length > 0 && (
            <div className="p-6 bg-surface border border-border rounded-xl">
              <h3 className="font-semibold text-foreground mb-4 flex items-center gap-2">
                <TrendingUp className="w-4 h-4 text-primary" /> 方向掌握度雷达
              </h3>
              <div className="w-full h-64">
                <ResponsiveContainer width="100%" height="100%">
                  <RadarChart data={dashboard.radar}>
                    <PolarGrid />
                    <PolarAngleAxis dataKey="category" tick={{ fontSize: 12, fill: '#94a3b8' }} />
                    <PolarRadiusAxis angle={90} domain={[0, 1]} tick={{ fontSize: 10 }} />
                    <Radar dataKey="mastery" stroke="#6366f1" fill="#6366f1" fillOpacity={0.4} />
                    <Tooltip />
                  </RadarChart>
                </ResponsiveContainer>
              </div>
            </div>
          )}

          {/* 学习热力图（30 天） */}
          {dashboard && dashboard.heatmap.length >= 0 && (
            <div className="p-6 bg-surface border border-border rounded-xl">
              <h3 className="font-semibold text-foreground mb-4 flex items-center gap-2">
                <Clock className="w-4 h-4 text-primary" /> 学习热力图（最近 30 天）
              </h3>
              <div className="flex flex-wrap gap-1">
                {Array.from({ length: 30 }, (_, i) => {
                  const d = new Date();
                  d.setDate(d.getDate() - (29 - i));
                  const dateStr = d.toISOString().slice(0, 10);
                  const day = dashboard.heatmap.find(h => h.date === dateStr);
                  const mins = day?.minutes || 0;
                  const intensity = mins === 0 ? 0 : Math.min(4, Math.ceil(mins / 30));
                  const colors = ['bg-surface-container', 'bg-primary/20', 'bg-primary/40', 'bg-primary/60', 'bg-primary/80'];
                  return (
                    <div key={i} className={`w-6 h-6 rounded ${colors[intensity]} border border-border/50`} title={`${dateStr}: ${mins} 分钟`} />
                  );
                })}
              </div>
              <div className="flex items-center gap-2 mt-3 text-xs text-muted-foreground">
                <span>少</span>
                {['bg-surface-container', 'bg-primary/20', 'bg-primary/40', 'bg-primary/60', 'bg-primary/80'].map((c, i) => (
                  <div key={i} className={`w-4 h-4 rounded ${c} border border-border/50`} />
                ))}
                <span>多</span>
              </div>
            </div>
          )}

          {/* 7 天趋势 */}
          {dashboard && dashboard.trend.length > 0 && (
            <div className="p-6 bg-surface border border-border rounded-xl">
              <h3 className="font-semibold text-foreground mb-4 flex items-center gap-2">
                <TrendingUp className="w-4 h-4 text-primary" /> 最近 7 天学习趋势
              </h3>
              <div className="w-full h-40">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={dashboard.trend}>
                    <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                    <XAxis dataKey="date" tick={{ fontSize: 10 }} tickFormatter={(d: string) => d.slice(5)} />
                    <YAxis tick={{ fontSize: 10 }} />
                    <Tooltip />
                    <Bar dataKey="minutes" fill="#6366f1" radius={[4, 4, 0, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </div>
          )}

          {/* 统计汇总 */}
          {dashboard && (
            <div className="grid grid-cols-4 gap-4">
              <div className="p-4 bg-surface border border-border rounded-xl text-center">
                <div className="text-2xl font-bold text-foreground">{Math.floor(dashboard.stats.total_minutes / 60)}h</div>
                <div className="text-xs text-muted-foreground">总时长</div>
              </div>
              <div className="p-4 bg-surface border border-border rounded-xl text-center">
                <div className="text-2xl font-bold text-foreground">{dashboard.stats.total_sessions}</div>
                <div className="text-xs text-muted-foreground">对话数</div>
              </div>
              <div className="p-4 bg-surface border border-border rounded-xl text-center">
                <div className="text-2xl font-bold text-foreground">{dashboard.stats.total_problems}</div>
                <div className="text-xs text-muted-foreground">解题数</div>
              </div>
              <div className="p-4 bg-surface border border-border rounded-xl text-center">
                <div className="text-2xl font-bold text-foreground">{(dashboard.stats.avg_accuracy * 100).toFixed(0)}%</div>
                <div className="text-xs text-muted-foreground">正确率</div>
              </div>
            </div>
          )}

          {/* 学习洞察（埋点数据分析） */}
          <div className="p-6 bg-surface border border-border rounded-xl">
            <h3 className="font-semibold text-foreground mb-4 flex items-center gap-2">
              <Sparkles className="w-4 h-4 text-primary" /> 学习洞察
            </h3>
            <InsightsCards insights={insights} />
          </div>
        </div>
      </div>

      {/* 编辑对话框 */}
      {editing && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onClick={() => setEditing(false)}>
          <div className="bg-background border border-border rounded-2xl p-6 w-full max-w-md mx-4" onClick={e => e.stopPropagation()}>
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-foreground">编辑学习画像</h2>
              <button onClick={() => setEditing(false)} className="text-muted-foreground hover:text-foreground"><X className="w-5 h-5" /></button>
            </div>
            <div className="space-y-4">
              <div>
                <label className="text-sm font-medium text-foreground">学习方式</label>
                <input value={editForm.learning_style} onChange={e => setEditForm({ ...editForm, learning_style: e.target.value })} placeholder="如：实践型、理论型" className="w-full mt-1 px-3 py-2 bg-surface-container border-none rounded-lg text-sm text-foreground focus:ring-2 focus:ring-primary/30 focus:outline-none" />
              </div>
              <div>
                <label className="text-sm font-medium text-foreground">偏好难度</label>
                <input value={editForm.preferred_difficulty} onChange={e => setEditForm({ ...editForm, preferred_difficulty: e.target.value })} placeholder="如：简单、中等、困难" className="w-full mt-1 px-3 py-2 bg-surface-container border-none rounded-lg text-sm text-foreground focus:ring-2 focus:ring-primary/30 focus:outline-none" />
              </div>
              <div>
                <label className="text-sm font-medium text-foreground">偏好时段</label>
                <select value={editForm.preferred_time_slot} onChange={e => setEditForm({ ...editForm, preferred_time_slot: e.target.value })} className="w-full mt-1 px-3 py-2 bg-surface-container border-none rounded-lg text-sm text-foreground focus:ring-2 focus:ring-primary/30 focus:outline-none">
                  <option value="">未设置</option>
                  <option value="morning">上午</option>
                  <option value="afternoon">下午</option>
                  <option value="evening">晚间</option>
                  <option value="night">深夜</option>
                </select>
              </div>
              <div>
                <label className="text-sm font-medium text-foreground">每日目标（分钟）</label>
                <input type="number" value={editForm.daily_goal} onChange={e => setEditForm({ ...editForm, daily_goal: parseInt(e.target.value) || 30 })} className="w-full mt-1 px-3 py-2 bg-surface-container border-none rounded-lg text-sm text-foreground focus:ring-2 focus:ring-primary/30 focus:outline-none" />
              </div>
              <div>
                <label className="text-sm font-medium text-foreground">擅长领域（逗号分隔）</label>
                <input value={editForm.focus_areas} onChange={e => setEditForm({ ...editForm, focus_areas: e.target.value })} placeholder="如：Python, 数据结构" className="w-full mt-1 px-3 py-2 bg-surface-container border-none rounded-lg text-sm text-foreground focus:ring-2 focus:ring-primary/30 focus:outline-none" />
              </div>
              <div>
                <label className="text-sm font-medium text-foreground">待加强（逗号分隔）</label>
                <input value={editForm.weak_areas} onChange={e => setEditForm({ ...editForm, weak_areas: e.target.value })} placeholder="如：并发编程, 系统设计" className="w-full mt-1 px-3 py-2 bg-surface-container border-none rounded-lg text-sm text-foreground focus:ring-2 focus:ring-primary/30 focus:outline-none" />
              </div>
            </div>
            <div className="flex gap-3 mt-6">
              <button onClick={() => setEditing(false)} className="flex-1 px-4 py-2 bg-surface-container text-foreground rounded-xl hover:opacity-80 text-sm font-medium">取消</button>
              <button onClick={saveEdit} className="flex-1 px-4 py-2 bg-primary text-primary-foreground rounded-xl hover:opacity-90 text-sm font-medium">保存</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
