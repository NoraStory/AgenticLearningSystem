'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { Braces, Loader2 } from 'lucide-react';
import { login, register, saveAuth } from '@/lib/api';

export default function AuthPage() {
  const router = useRouter();
  const [mode, setMode] = useState<'login' | 'register'>('login');
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const switchMode = (next: 'login' | 'register') => {
    setMode(next);
    setError(null);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    // 前端校验
    if (!email.includes('@') || email.length < 5) {
      setError('请输入有效的邮箱地址');
      return;
    }
    if (password.length < 8) {
      setError('密码至少 8 位');
      return;
    }
    if (mode === 'register' && username.trim().length < 2) {
      setError('用户名至少 2 个字符');
      return;
    }

    setLoading(true);
    try {
      const result =
        mode === 'login'
          ? await login(email.trim(), password)
          : await register(username.trim(), email.trim(), password);
      saveAuth(result);
      router.push('/');
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : '操作失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-[calc(100vh-3.5rem)] flex items-center justify-center px-4 py-10 bg-background">
      <div className="w-full max-w-md">
        {/* Logo */}
        <div className="flex items-center justify-center gap-2.5 mb-6">
          <Braces className="text-primary w-7 h-7" />
          <span className="font-bold text-2xl text-foreground">CodeForge</span>
          <span className="text-[11px] text-primary-foreground bg-primary px-2 py-0.5 rounded-sm font-medium">
            Academy
          </span>
        </div>

        <div className="bg-surface-container rounded-2xl border border-border shadow-sm overflow-hidden">
          {/* Tab 切换 */}
          <div className="flex border-b border-border">
            {(['login', 'register'] as const).map((m) => (
              <button
                key={m}
                type="button"
                onClick={() => switchMode(m)}
                className={`flex-1 py-3 text-sm font-medium transition-colors ${
                  mode === m
                    ? 'text-primary border-b-2 border-primary bg-primary/5'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
              >
                {m === 'login' ? '登录' : '注册'}
              </button>
            ))}
          </div>

          <form onSubmit={handleSubmit} className="p-6 space-y-4">
            {mode === 'register' && (
              <div>
                <label className="block text-sm text-muted-foreground mb-1.5">用户名</label>
                <input
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  placeholder="2 个字符以上"
                  className="w-full px-3.5 py-2.5 bg-surface rounded-xl border border-border text-sm text-foreground placeholder:text-muted-foreground/40 focus:outline-none focus:ring-2 focus:ring-primary/30"
                />
              </div>
            )}

            <div>
              <label className="block text-sm text-muted-foreground mb-1.5">邮箱</label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="you@example.com"
                autoComplete="email"
                className="w-full px-3.5 py-2.5 bg-surface rounded-xl border border-border text-sm text-foreground placeholder:text-muted-foreground/40 focus:outline-none focus:ring-2 focus:ring-primary/30"
              />
            </div>

            <div>
              <label className="block text-sm text-muted-foreground mb-1.5">密码</label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="至少 8 位"
                autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
                className="w-full px-3.5 py-2.5 bg-surface rounded-xl border border-border text-sm text-foreground placeholder:text-muted-foreground/40 focus:outline-none focus:ring-2 focus:ring-primary/30"
              />
            </div>

            {error && (
              <div className="rounded-lg bg-destructive/10 px-3 py-2.5 text-sm text-destructive">
                {error}
              </div>
            )}

            <button
              type="submit"
              disabled={loading}
              className="w-full flex items-center justify-center gap-2 py-2.5 bg-primary text-primary-foreground rounded-xl text-sm font-medium hover:opacity-90 disabled:opacity-50 transition-opacity"
            >
              {loading && <Loader2 className="w-4 h-4 animate-spin" />}
              {mode === 'login' ? '登录' : '创建账号'}
            </button>

            <p className="text-xs text-muted-foreground/70 text-center">
              密码使用 Argon2id 加密存储，不会以明文保存
            </p>
          </form>
        </div>

        <p className="text-center text-sm text-muted-foreground mt-5">
          <Link href="/" className="hover:text-primary transition-colors">
            ← 返回首页
          </Link>
        </p>
      </div>
    </div>
  );
}
