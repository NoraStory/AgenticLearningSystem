'use client';

import { useRef, useState } from 'react';
import { apiFetch } from '@/lib/api';
import { Loader2, Upload, X, AlertCircle, CheckCircle } from 'lucide-react';

interface TemplateSection {
  title: string;
  items: string[];
}

interface TemplateRegisterModalProps {
  onClose: () => void;
  onRegistered: () => void;
}

export default function TemplateRegisterModal({ onClose, onRegistered }: TemplateRegisterModalProps) {
  const [file, setFile] = useState<File | null>(null);
  const [templateId, setTemplateId] = useState('');
  const [sections, setSections] = useState<TemplateSection[]>([]);
  const [name, setName] = useState('');
  const [step, setStep] = useState<'upload' | 'confirm' | 'done'>('upload');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const f = e.target.files?.[0];
    if (f) {
      setFile(f);
      setName(f.name.replace(/\.docx$/i, ''));
      setError('');
    }
  };

  const handleUpload = async () => {
    if (!file) return;
    setIsLoading(true);
    setError('');
    try {
      const form = new FormData();
      form.append('file', file);
      const result = await apiFetch<{ template_id: string; name: string; sections: TemplateSection[] }>(
        '/api/v1/resume/templates/upload',
        { method: 'POST', body: form },
      );
      setTemplateId(result.template_id);
      setSections(result.sections || []);
      setName(result.name || name);
      setStep('confirm');
    } catch (err) {
      setError(err instanceof Error ? err.message : '上传失败');
    } finally {
      setIsLoading(false);
    }
  };

  const updateSectionTitle = (idx: number, value: string) => {
    setSections((prev) => prev.map((s, i) => (i === idx ? { ...s, title: value } : s)));
  };

  const updateItem = (si: number, ii: number, value: string) => {
    setSections((prev) => prev.map((s, i) =>
      i === si ? { ...s, items: s.items.map((it, j) => (j === ii ? value : it)) } : s,
    ));
  };

  const removeItem = (si: number, ii: number) => {
    setSections((prev) => prev.map((s, i) =>
      i === si ? { ...s, items: s.items.filter((_, j) => j !== ii) } : s,
    ));
  };

  const addItem = (si: number) => {
    setSections((prev) => prev.map((s, i) => (i === si ? { ...s, items: [...s.items, ''] } : s)));
  };

  const handleConfirm = async () => {
    if (!templateId) return;
    setIsLoading(true);
    setError('');
    try {
      await apiFetch(`/api/v1/resume/templates/${templateId}/confirm`, {
        method: 'POST',
        body: JSON.stringify({ name, sections }),
      });
      setStep('done');
    } catch (err) {
      setError(err instanceof Error ? err.message : '注册失败');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="bg-card rounded-2xl border border-border shadow-xl max-w-2xl w-full max-h-[85vh] flex flex-col">
        <div className="flex items-center justify-between px-6 py-4 border-b border-border">
          <h2 className="text-lg font-semibold text-foreground">上传简历模板</h2>
          <button onClick={onClose} className="text-muted-foreground hover:text-foreground">
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="flex-1 overflow-y-auto px-6 py-4 space-y-4">
          {error && (
            <div className="flex items-center gap-2 rounded-lg bg-destructive/10 px-3 py-2 text-sm text-destructive">
              <AlertCircle className="w-4 h-4 flex-shrink-0" />
              {error}
            </div>
          )}

          {step === 'upload' && (
            <div>
              <div
                onClick={() => fileInputRef.current?.click()}
                className="border-2 border-dashed border-border rounded-xl p-8 text-center cursor-pointer hover:border-primary/50 transition-colors"
              >
                <Upload className="w-10 h-10 text-muted-foreground mx-auto mb-3" />
                {file ? (
                  <div>
                    <p className="text-foreground font-medium mb-1">{file.name}</p>
                    <p className="text-sm text-muted-foreground">点击更换文件</p>
                  </div>
                ) : (
                  <div>
                    <p className="text-foreground mb-2">选择你的 DOCX 简历模板</p>
                    <p className="text-sm text-muted-foreground">系统将自动解析章节结构,AI 识别后需你确认</p>
                  </div>
                )}
              </div>
              <input
                ref={fileInputRef}
                type="file"
                accept=".docx"
                onChange={handleFileSelect}
                className="hidden"
              />
            </div>
          )}

          {step === 'confirm' && (
            <div>
              <div className="mb-4">
                <label className="text-sm text-muted-foreground mb-1 block">模板名称</label>
                <input
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  className="w-full px-3 py-2 bg-surface-container rounded-lg border border-border text-sm"
                />
              </div>
              <p className="text-sm text-muted-foreground mb-3">
                AI 已识别以下章节,请确认或修正后注册。章节内的条目将作为可替换内容。
              </p>
              {sections.map((section, si) => (
                <div key={si} className="mb-4 rounded-xl border border-border p-3">
                  <input
                    value={section.title}
                    onChange={(e) => updateSectionTitle(si, e.target.value)}
                    className="w-full px-2 py-1 bg-transparent font-medium text-foreground border-b border-border mb-2 text-sm"
                    placeholder="章节标题"
                  />
                  <div className="space-y-1.5">
                    {section.items.map((item, ii) => (
                      <div key={ii} className="flex items-center gap-2">
                        <input
                          value={item}
                          onChange={(e) => updateItem(si, ii, e.target.value)}
                          className="flex-1 px-2 py-1 bg-surface-container rounded text-sm"
                          placeholder="条目内容"
                        />
                        <button
                          onClick={() => removeItem(si, ii)}
                          className="text-muted-foreground hover:text-destructive"
                          title="删除条目"
                        >
                          <X className="w-4 h-4" />
                        </button>
                      </div>
                    ))}
                  </div>
                  <button
                    onClick={() => addItem(si)}
                    className="mt-2 text-xs text-muted-foreground hover:text-primary"
                  >
                    + 添加条目
                  </button>
                </div>
              ))}
            </div>
          )}

          {step === 'done' && (
            <div className="text-center py-8">
              <CheckCircle className="w-12 h-12 text-green-500 mx-auto mb-3" />
              <p className="text-foreground font-medium mb-1">模板注册成功</p>
              <p className="text-sm text-muted-foreground">现在可以在简历优化中选择该模板导出。</p>
            </div>
          )}
        </div>

        <div className="flex justify-end gap-3 px-6 py-4 border-t border-border">
          {step === 'upload' && (
            <>
              <button
                onClick={onClose}
                className="px-4 py-2 rounded-lg text-sm text-muted-foreground hover:bg-muted"
              >
                取消
              </button>
              <button
                onClick={handleUpload}
                disabled={!file || isLoading}
                className="px-4 py-2 rounded-lg text-sm bg-primary text-primary-foreground font-medium hover:bg-primary/90 disabled:opacity-50 flex items-center gap-2"
              >
                {isLoading && <Loader2 className="w-4 h-4 animate-spin" />}
                解析模板
              </button>
            </>
          )}
          {step === 'confirm' && (
            <>
              <button
                onClick={() => setStep('upload')}
                className="px-4 py-2 rounded-lg text-sm text-muted-foreground hover:bg-muted"
              >
                返回
              </button>
              <button
                onClick={handleConfirm}
                disabled={isLoading || sections.length === 0}
                className="px-4 py-2 rounded-lg text-sm bg-primary text-primary-foreground font-medium hover:bg-primary/90 disabled:opacity-50 flex items-center gap-2"
              >
                {isLoading && <Loader2 className="w-4 h-4 animate-spin" />}
                确认注册
              </button>
            </>
          )}
          {step === 'done' && (
            <button
              onClick={() => {
                onRegistered();
                onClose();
              }}
              className="px-4 py-2 rounded-lg text-sm bg-primary text-primary-foreground font-medium"
            >
              完成
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
