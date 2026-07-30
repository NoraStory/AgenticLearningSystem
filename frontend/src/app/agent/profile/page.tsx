'use client';

import { Brain, Target, TrendingUp, Award, BookOpen, Code, Database, Sparkles, Network } from 'lucide-react';
import { useEffect, useState } from 'react';
import { apiFetch } from '@/lib/api';

// 用户画像数据
const fallbackProfile = {
  level: '中级开发者',
  focusAreas: ['Python', '数据结构', 'AI Agent'],
  weakAreas: ['并发编程', '系统设计'],
  learningStyle: '实践型',
  preferredDifficulty: '中等',
  dailyGoal: 60, // 分钟
  totalStudyTime: 256, // 小时
  streak: 15, // 天
};

// 知识掌握度
const fallbackKnowledgeAreas = [
  { name: 'Python 基础', level: 85, icon: Code, color: 'bg-blue-500' },
  { name: 'Python 进阶', level: 62, icon: Code, color: 'bg-blue-600' },
  { name: 'C++ 基础', level: 45, icon: Code, color: 'bg-purple-500' },
  { name: '数据结构', level: 58, icon: TrendingUp, color: 'bg-green-500' },
  { name: '算法', level: 35, icon: Target, color: 'bg-orange-500' },
  { name: '数据库', level: 40, icon: Database, color: 'bg-cyan-500' },
  { name: 'AI Agent', level: 28, icon: Sparkles, color: 'bg-rose-500' },
  { name: '系统设计', level: 15, icon: Brain, color: 'bg-indigo-500' },
];

// 学习偏好
const preferences = [
  { label: '学习方式', value: '实践为主，理论为辅' },
  { label: '偏好难度', value: '中等挑战' },
  { label: '学习时段', value: '晚间 20:00-23:00' },
  { label: '单次时长', value: '45-60 分钟' },
  { label: '擅长领域', value: 'Python、Web 开发' },
  { label: '待加强', value: '算法、并发编程' },
];

// 最近的知识图谱节点
const fallbackRecentTopics = [
  { name: 'Python 装饰器', connections: 5, mastery: 78 },
  { name: '二叉树遍历', connections: 3, mastery: 65 },
  { name: 'SQL 连接', connections: 4, mastery: 72 },
  { name: 'LangChain Chain', connections: 2, mastery: 45 },
  { name: '动态规划', connections: 6, mastery: 38 },
];

export default function AgentProfilePage() {
  const [profile, setProfile] = useState(fallbackProfile);
  const [knowledgeAreas, setKnowledgeAreas] = useState(fallbackKnowledgeAreas);
  const [recentTopics, setRecentTopics] = useState(fallbackRecentTopics);

  useEffect(() => {
    Promise.all([
      apiFetch<{ level: string; focus_areas: string[]; weak_areas: string[]; learning_style: string; preferred_difficulty: string; daily_goal: number; total_study_time: number; streak: number }>('/api/v1/agent/profile'),
      apiFetch<{ areas: Array<{ name: string; level: number; color: string }>; recent_topics: typeof fallbackRecentTopics }>('/api/v1/agent/knowledge'),
    ]).then(([profileData, knowledge]) => {
      setProfile({ level: profileData.level, focusAreas: profileData.focus_areas, weakAreas: profileData.weak_areas, learningStyle: profileData.learning_style, preferredDifficulty: profileData.preferred_difficulty, dailyGoal: profileData.daily_goal, totalStudyTime: profileData.total_study_time, streak: profileData.streak });
      setKnowledgeAreas(knowledge.areas.map((area, index) => ({ ...fallbackKnowledgeAreas[index % fallbackKnowledgeAreas.length], ...area })));
      setRecentTopics(knowledge.recent_topics);
    }).catch(() => undefined);
  }, []);
  return (
    <div className="max-w-6xl mx-auto">
      {/* 页面标题 */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-foreground">学习画像</h1>
        <p className="text-muted-foreground mt-2">
          AI 基于你的学习行为生成的个性化画像，用于智能推荐和路由优化
        </p>
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
              <div>
                <h3 className="font-semibold text-foreground">{profile.level}</h3>
                <p className="text-sm text-muted-foreground">Lv.5 进阶学习者</p>
              </div>
            </div>
            <div className="grid grid-cols-3 gap-4 text-center">
              <div>
                <div className="text-2xl font-bold text-foreground">{profile.totalStudyTime}h</div>
                <div className="text-xs text-muted-foreground">总学习时长</div>
              </div>
              <div>
                <div className="text-2xl font-bold text-foreground">{profile.streak}</div>
                <div className="text-xs text-muted-foreground">连续打卡</div>
              </div>
              <div>
                <div className="text-2xl font-bold text-foreground">{profile.dailyGoal}min</div>
                <div className="text-xs text-muted-foreground">每日目标</div>
              </div>
            </div>
          </div>

          {/* 学习偏好 */}
          <div className="p-6 bg-surface border border-border rounded-xl">
            <h3 className="font-semibold text-foreground mb-4 flex items-center gap-2">
              <Target className="w-4 h-4 text-primary" />
              学习偏好
            </h3>
            <div className="space-y-3">
              {preferences.map((pref, i) => (
                <div key={i} className="flex justify-between items-center">
                  <span className="text-sm text-muted-foreground">{pref.label}</span>
                  <span className="text-sm font-medium text-foreground">{pref.value}</span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* 右侧：知识掌握度 */}
        <div className="lg:col-span-2 space-y-6">
          {/* 知识领域掌握度 */}
          <div className="p-6 bg-surface border border-border rounded-xl">
            <h3 className="font-semibold text-foreground mb-4 flex items-center gap-2">
              <TrendingUp className="w-4 h-4 text-primary" />
              知识领域掌握度
            </h3>
            <div className="space-y-4">
              {knowledgeAreas.map((area, i) => {
                const Icon = area.icon;
                return (
                  <div key={i}>
                    <div className="flex items-center justify-between mb-1.5">
                      <div className="flex items-center gap-2">
                        <Icon className="w-4 h-4 text-muted-foreground" />
                        <span className="text-sm font-medium text-foreground">{area.name}</span>
                      </div>
                      <span className="text-sm text-muted-foreground">{area.level}%</span>
                    </div>
                    <div className="h-2 bg-surface-container rounded-full overflow-hidden">
                      <div 
                        className={`h-full ${area.color} rounded-full transition-all`}
                        style={{ width: `${area.level}%` }}
                      />
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* 知识图谱可视化 */}
          <div className="p-6 bg-surface border border-border rounded-xl">
            <h3 className="font-semibold text-foreground mb-4 flex items-center gap-2">
              <Network className="w-4 h-4 text-primary" />
              知识图谱
            </h3>
            <div className="relative h-64 bg-surface-container rounded-xl overflow-hidden">
              {/* 知识节点 */}
              <svg className="absolute inset-0 w-full h-full">
                {/* 连接线 */}
                <line x1="50%" y1="30%" x2="30%" y2="50%" stroke="currentColor" strokeWidth="1" className="text-border" />
                <line x1="50%" y1="30%" x2="70%" y2="50%" stroke="currentColor" strokeWidth="1" className="text-border" />
                <line x1="30%" y1="50%" x2="20%" y2="75%" stroke="currentColor" strokeWidth="1" className="text-border" />
                <line x1="30%" y1="50%" x2="40%" y2="75%" stroke="currentColor" strokeWidth="1" className="text-border" />
                <line x1="70%" y1="50%" x2="60%" y2="75%" stroke="currentColor" strokeWidth="1" className="text-border" />
                <line x1="70%" y1="50%" x2="80%" y2="75%" stroke="currentColor" strokeWidth="1" className="text-border" />
              </svg>
              {/* 节点 */}
              <div className="absolute top-[25%] left-[45%] w-12 h-12 bg-primary rounded-full flex items-center justify-center text-primary-foreground text-xs font-bold shadow-lg">
                基础
              </div>
              <div className="absolute top-[45%] left-[25%] w-10 h-10 bg-green-500 rounded-full flex items-center justify-center text-white text-xs font-medium shadow">
                语法
              </div>
              <div className="absolute top-[45%] left-[65%] w-10 h-10 bg-blue-500 rounded-full flex items-center justify-center text-white text-xs font-medium shadow">
                算法
              </div>
              <div className="absolute top-[70%] left-[15%] w-8 h-8 bg-green-400 rounded-full flex items-center justify-center text-white text-xs shadow">
                变量
              </div>
              <div className="absolute top-[70%] left-[35%] w-8 h-8 bg-green-400 rounded-full flex items-center justify-center text-white text-xs shadow">
                函数
              </div>
              <div className="absolute top-[70%] left-[55%] w-8 h-8 bg-blue-400 rounded-full flex items-center justify-center text-white text-xs shadow">
                排序
              </div>
              <div className="absolute top-[70%] left-[75%] w-8 h-8 bg-blue-400 rounded-full flex items-center justify-center text-white text-xs shadow">
                查找
              </div>
            </div>
            <div className="mt-3 flex items-center justify-center gap-4 text-xs text-muted-foreground">
              <span className="flex items-center gap-1"><span className="w-3 h-3 bg-green-500 rounded-full"></span>已掌握</span>
              <span className="flex items-center gap-1"><span className="w-3 h-3 bg-blue-500 rounded-full"></span>学习中</span>
              <span className="flex items-center gap-1"><span className="w-3 h-3 bg-gray-300 rounded-full"></span>未开始</span>
            </div>
          </div>

          {/* 知识图谱节点 */}
          <div className="p-6 bg-surface border border-border rounded-xl">
            <h3 className="font-semibold text-foreground mb-4 flex items-center gap-2">
              <Sparkles className="w-4 h-4 text-primary" />
              最近学习的知识点
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              {recentTopics.map((topic, i) => (
                <div key={i} className="flex items-center gap-3 p-3 bg-surface-container rounded-lg">
                  <div className="flex-1">
                    <div className="text-sm font-medium text-foreground">{topic.name}</div>
                    <div className="text-xs text-muted-foreground">
                      {topic.connections} 个关联 · 掌握度 {topic.mastery}%
                    </div>
                  </div>
                  <div className="w-10 h-10 relative">
                    <svg className="w-10 h-10 -rotate-90">
                      <circle cx="20" cy="20" r="16" fill="none" stroke="currentColor" strokeWidth="3" className="text-surface-container" />
                      <circle 
                        cx="20" cy="20" r="16" fill="none" stroke="currentColor" strokeWidth="3" 
                        className="text-primary"
                        strokeDasharray={`${topic.mastery} ${100 - topic.mastery}`}
                        strokeLinecap="round"
                      />
                    </svg>
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* 强弱项分析 */}
          <div className="grid grid-cols-2 gap-4">
            <div className="p-4 bg-green-50 border border-green-200 rounded-xl">
              <h4 className="text-sm font-medium text-green-800 mb-2">擅长领域</h4>
              <div className="flex flex-wrap gap-2">
                {profile.focusAreas.map((area, i) => (
                  <span key={i} className="px-2 py-1 text-xs bg-green-100 text-green-700 rounded">
                    {area}
                  </span>
                ))}
              </div>
            </div>
            <div className="p-4 bg-orange-50 border border-orange-200 rounded-xl">
              <h4 className="text-sm font-medium text-orange-800 mb-2">待加强</h4>
              <div className="flex flex-wrap gap-2">
                {profile.weakAreas.map((area, i) => (
                  <span key={i} className="px-2 py-1 text-xs bg-orange-100 text-orange-700 rounded">
                    {area}
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
