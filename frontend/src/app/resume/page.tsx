'use client';

import { useState, useRef, useEffect } from 'react';
import { apiDownload, apiFetch } from '@/lib/api';
import {
  Upload,
  FileText,
  Sparkles,
  CheckCircle,
  AlertCircle,
  Download,
  ChevronRight,
  Loader2,
  Eye,
  RefreshCw,
  X,
} from 'lucide-react';

// 模板类型定义（对应数据库结构）
interface ResumeTemplate {
  id: string;
  name: string;
  description: string;
  category: string;
  preview: string;
  structure: {
    sections: string[];
    style: string;
  };
}

// 分析结果类型
interface AnalysisResult {
  score: number;
  atsScore: number;
  keywordMatch: { keyword: string; found: boolean }[];
  strengths: string[];
  weaknesses: string[];
  suggestions: string[];
}

export default function ResumePage() {
  const [activeTab, setActiveTab] = useState<'analyze' | 'optimize'>('analyze');
  const [uploadedFile, setUploadedFile] = useState<File | null>(null);
  const [uploadedFileId, setUploadedFileId] = useState<string>('');
  const [isAnalyzing, setIsAnalyzing] = useState(false);
  const [analysisResult, setAnalysisResult] = useState<AnalysisResult | null>(null);

  // 模板相关状态
  const [templates, setTemplates] = useState<ResumeTemplate[]>([]);
  const [isLoadingTemplates, setIsLoadingTemplates] = useState(true);
  const [selectedTemplate, setSelectedTemplate] = useState<string>('');
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [templateDetail, setTemplateDetail] = useState<ResumeTemplate | null>(null);
  const [isOptimizing, setIsOptimizing] = useState(false);
  const [optimizedResume, setOptimizedResume] = useState<string>('');
  const [exportFormat, setExportFormat] = useState<'pdf' | 'docx' | 'html'>('pdf');
  const fileInputRef = useRef<HTMLInputElement>(null);

  // 从后端数据库读取模板
  useEffect(() => {
    apiFetch<{ templates: ResumeTemplate[] }>('/api/v1/resume/templates')
      .then((data) => setTemplates(data.templates))
      .catch(() => setTemplates(getMockTemplates()))
      .finally(() => setIsLoadingTemplates(false));
  }, []);
  // Mock 模板数据（实际应从数据库获取）
  const getMockTemplates = (): ResumeTemplate[] => [
    {
      id: 'tech-standard',
      name: '技术岗位标准版',
      description: '适合程序员、工程师等技术岗位',
      category: 'tech',
      preview: '💻',
      structure: {
        sections: ['基本信息', '技能清单', '工作经历', '项目经验', '教育背景', '开源与社区'],
        style: '简洁专业，突出技术栈和项目成果',
      },
    },
    {
      id: 'tech-modern',
      name: '技术岗位现代版',
      description: '现代化设计，适合互联网行业',
      category: 'tech',
      preview: '🚀',
      structure: {
        sections: ['基本信息', '技术栈', '工作经历', '项目亮点', '开源贡献', '教育背景'],
        style: '现代简约，强调技术深度和影响力',
      },
    },
    {
      id: 'product',
      name: '产品经理版',
      description: '适合产品经理、运营等岗位',
      category: 'product',
      preview: '📊',
      structure: {
        sections: ['基本信息', '个人简介', '工作经历', '项目经验', '数据成果', '教育背景'],
        style: '数据驱动，突出产品思维和业务成果',
      },
    },
    {
      id: 'design',
      name: '设计师版',
      description: '适合 UI/UX 设计师',
      category: 'design',
      preview: '🎨',
      structure: {
        sections: ['基本信息', '设计技能', '作品集', '工作经历', '教育背景'],
        style: '视觉优先，展示设计作品和审美',
      },
    },
    {
      id: 'marketing',
      name: '市场营销版',
      description: '适合市场、运营、销售等岗位',
      category: 'marketing',
      preview: '📈',
      structure: {
        sections: ['基本信息', '核心能力', '工作经历', '业绩成果', '教育背景'],
        style: '结果导向，突出业绩数据和营销能力',
      },
    },
    {
      id: 'general',
      name: '通用版',
      description: '适合大多数岗位',
      category: 'general',
      preview: '📋',
      structure: {
        sections: ['基本信息', '个人简介', '工作经历', '教育背景', '技能证书'],
        style: '经典布局，平衡各方面内容',
      },
    },
    {
      id: 'academic',
      name: '学术版',
      description: '适合科研、教育等学术岗位',
      category: 'academic',
      preview: '🎓',
      structure: {
        sections: ['基本信息', '教育背景', '研究方向', '发表论文', '科研项目', '学术荣誉'],
        style: '学术规范，突出研究成果和学术背景',
      },
    },
    {
      id: 'fresh-graduate',
      name: '应届生版',
      description: '适合应届毕业生',
      category: 'fresh',
      preview: '🌟',
      structure: {
        sections: ['基本信息', '教育背景', '实习经历', '项目经验', '校园经历', '技能证书'],
        style: '突出学习能力和潜力，弱化工作经验',
      },
    },
  ];

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setUploadedFile(file);
      setAnalysisResult(null);
      setOptimizedResume('');
    }
  };

  const handleAnalyze = async () => {
    if (!uploadedFile) return;
    setIsAnalyzing(true);
    try {
      const form = new FormData();
      form.append('file', uploadedFile);
      const uploaded = await apiFetch<{ file_id: string }>('/api/v1/resume/upload', { method: 'POST', body: form });
      setUploadedFileId(uploaded.file_id);
      const result = await apiFetch<{ analysis: AnalysisResult }>('/api/v1/resume/analyze', {
        method: 'POST', body: JSON.stringify({ file_id: uploaded.file_id }),
      });
      setAnalysisResult(result.analysis);
    } catch (error) {
      console.error(error);
    } finally {
      setIsAnalyzing(false);
    }
  };
  const handleSelectTemplate = (templateId: string) => {
    setSelectedTemplate(templateId);
    const template = templates.find((t) => t.id === templateId);
    setTemplateDetail(template || null);
  };

  const handleOptimize = async () => {
    if (!analysisResult || !selectedTemplate || !uploadedFileId) return;
    setIsOptimizing(true);
    try {
      const result = await apiFetch<{ optimized_content: { text: string } }>('/api/v1/resume/optimize', {
        method: 'POST',
        body: JSON.stringify({ file_id: uploadedFileId, template_id: selectedTemplate, optimization_directions: analysisResult.suggestions }),
      });
      setOptimizedResume(result.optimized_content.text);
    } catch (error) {
      console.error(error);
    } finally {
      setIsOptimizing(false);
    }
  };
  const handleExport = () => {
    // 模拟导出
    alert(`正在导出为 ${exportFormat.toUpperCase()} 格式...`);
  };

  const handleRemoveFile = () => {
    setUploadedFile(null);
    setUploadedFileId('');
    setAnalysisResult(null);
    setOptimizedResume('');
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  // 获取分类列表
  const categories = [
    { id: 'all', name: '全部' },
    { id: 'tech', name: '技术岗位' },
    { id: 'product', name: '产品经理' },
    { id: 'design', name: '设计师' },
    { id: 'marketing', name: '市场营销' },
    { id: 'general', name: '通用' },
    { id: 'academic', name: '学术' },
    { id: 'fresh', name: '应届生' },
  ];

  // 根据分类筛选模板
  const filteredTemplates =
    selectedCategory === 'all'
      ? templates
      : templates.filter((t) => t.category === selectedCategory);

  return (
    <div className="flex-1 overflow-y-auto">
      <div className="max-w-6xl mx-auto p-8">
        {/* 页面标题 */}
        <div className="mb-8">
          <h1 className="text-2xl font-bold text-foreground mb-2">简历分析与优化</h1>
          <p className="text-muted-foreground">AI 驱动的智能简历服务，助你打造完美简历</p>
        </div>

        {/* Tab 切换 */}
        <div className="flex gap-2 mb-6 border-b border-border pb-4">
          <button
            onClick={() => setActiveTab('analyze')}
            className={`px-4 py-2 rounded-lg font-medium transition-colors ${
              activeTab === 'analyze'
                ? 'bg-primary text-primary-foreground'
                : 'bg-muted text-muted-foreground hover:bg-muted/80'
            }`}
          >
            简历分析
          </button>
          <button
            onClick={() => setActiveTab('optimize')}
            className={`px-4 py-2 rounded-lg font-medium transition-colors ${
              activeTab === 'optimize'
                ? 'bg-primary text-primary-foreground'
                : 'bg-muted text-muted-foreground hover:bg-muted/80'
            }`}
          >
            简历优化
          </button>
        </div>

        {/* 简历分析 Tab */}
        {activeTab === 'analyze' && (
          <div className="space-y-6">
            {/* 上传区域 */}
            <div className="bg-card rounded-xl border border-border p-6">
              <h2 className="text-lg font-semibold text-foreground mb-4">上传简历</h2>
              <div
                onClick={() => fileInputRef.current?.click()}
                className="border-2 border-dashed border-border rounded-xl p-8 text-center cursor-pointer hover:border-primary/50 transition-colors"
              >
                <Upload className="w-12 h-12 text-muted-foreground mx-auto mb-4" />
                {uploadedFile ? (
                  <div>
                    <p className="text-foreground font-medium mb-2">{uploadedFile.name}</p>
                    <p className="text-sm text-muted-foreground">
                      {(uploadedFile.size / 1024 / 1024).toFixed(2)} MB
                    </p>
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        handleRemoveFile();
                      }}
                      className="mt-4 px-4 py-2 bg-destructive/10 text-destructive rounded-lg hover:bg-destructive/20 transition-colors"
                    >
                      移除文件
                    </button>
                  </div>
                ) : (
                  <div>
                    <p className="text-foreground mb-2">点击或拖拽文件到此处上传</p>
                    <p className="text-sm text-muted-foreground">支持 PDF、DOC、DOCX 格式，最大 10MB</p>
                  </div>
                )}
                <input
                  ref={fileInputRef}
                  type="file"
                  accept=".pdf,.doc,.docx"
                  onChange={handleFileUpload}
                  className="hidden"
                />
              </div>

              {uploadedFile && !analysisResult && (
                <button
                  onClick={handleAnalyze}
                  disabled={isAnalyzing}
                  className="mt-4 w-full py-3 bg-primary text-primary-foreground rounded-lg font-medium hover:bg-primary/90 transition-colors disabled:opacity-50 flex items-center justify-center gap-2"
                >
                  {isAnalyzing ? (
                    <>
                      <Loader2 className="w-5 h-5 animate-spin" />
                      AI 正在分析...
                    </>
                  ) : (
                    <>
                      <Sparkles className="w-5 h-5" />
                      开始 AI 分析
                    </>
                  )}
                </button>
              )}
            </div>

            {/* 分析结果 */}
            {analysisResult && (
              <div className="bg-card rounded-xl border border-border p-6">
                <h2 className="text-lg font-semibold text-foreground mb-4">分析结果</h2>

                {/* 评分 */}
                <div className="grid grid-cols-2 gap-4 mb-6">
                  <div className="bg-muted/50 rounded-xl p-4 text-center">
                    <p className="text-sm text-muted-foreground mb-1">综合评分</p>
                    <p
                      className={`text-4xl font-bold ${
                        analysisResult.score >= 80
                          ? 'text-green-500'
                          : analysisResult.score >= 60
                          ? 'text-amber-500'
                          : 'text-red-500'
                      }`}
                    >
                      {analysisResult.score}
                    </p>
                  </div>
                  <div className="bg-muted/50 rounded-xl p-4 text-center">
                    <p className="text-sm text-muted-foreground mb-1">ATS 友好度</p>
                    <p
                      className={`text-4xl font-bold ${
                        analysisResult.atsScore >= 80
                          ? 'text-green-500'
                          : analysisResult.atsScore >= 60
                          ? 'text-amber-500'
                          : 'text-red-500'
                      }`}
                    >
                      {analysisResult.atsScore}
                    </p>
                  </div>
                </div>

                {/* 关键词匹配 */}
                <div className="mb-6">
                  <h3 className="font-medium text-foreground mb-3">关键词匹配</h3>
                  <div className="flex flex-wrap gap-2">
                    {analysisResult.keywordMatch.map((item, index) => (
                      <span
                        key={index}
                        className={`px-3 py-1 rounded-full text-sm ${
                          item.found
                            ? 'bg-green-500/10 text-green-600'
                            : 'bg-red-500/10 text-red-600'
                        }`}
                      >
                        {item.keyword} {item.found ? '✓' : '✗'}
                      </span>
                    ))}
                  </div>
                </div>

                {/* 亮点 */}
                <div className="mb-6">
                  <h3 className="font-medium text-foreground mb-3 flex items-center gap-2">
                    <CheckCircle className="w-5 h-5 text-green-500" />
                    亮点
                  </h3>
                  <ul className="space-y-2">
                    {analysisResult.strengths.map((item, index) => (
                      <li key={index} className="flex items-start gap-2 text-sm text-foreground">
                        <ChevronRight className="w-4 h-4 text-green-500 mt-0.5 flex-shrink-0" />
                        {item}
                      </li>
                    ))}
                  </ul>
                </div>

                {/* 改进建议 */}
                <div>
                  <h3 className="font-medium text-foreground mb-3 flex items-center gap-2">
                    <AlertCircle className="w-5 h-5 text-amber-500" />
                    改进建议
                  </h3>
                  <ul className="space-y-2">
                    {analysisResult.suggestions.map((item, index) => (
                      <li key={index} className="flex items-start gap-2 text-sm text-foreground">
                        <ChevronRight className="w-4 h-4 text-amber-500 mt-0.5 flex-shrink-0" />
                        {item}
                      </li>
                    ))}
                  </ul>
                </div>

                <button
                  onClick={() => setActiveTab('optimize')}
                  className="mt-6 w-full py-3 bg-primary text-primary-foreground rounded-lg font-medium hover:bg-primary/90 transition-colors flex items-center justify-center gap-2"
                >
                  <Sparkles className="w-5 h-5" />
                  前往简历优化
                </button>
              </div>
            )}
          </div>
        )}

        {/* 简历优化 Tab */}
        {activeTab === 'optimize' && (
          <div className="space-y-6">
            {/* 选择模板 */}
            <div className="bg-card rounded-xl border border-border p-6">
              <h2 className="text-lg font-semibold text-foreground mb-4">选择简历模板</h2>
              <p className="text-sm text-muted-foreground mb-4">
                模板从数据库加载，AI 将读取模板结构并写入优化内容
              </p>

              {/* 分类筛选 */}
              <div className="flex flex-wrap gap-2 mb-4">
                {categories.map((cat) => (
                  <button
                    key={cat.id}
                    onClick={() => setSelectedCategory(cat.id)}
                    className={`px-3 py-1.5 rounded-lg text-sm transition-colors ${
                      selectedCategory === cat.id
                        ? 'bg-primary text-primary-foreground'
                        : 'bg-muted text-muted-foreground hover:bg-muted/80'
                    }`}
                  >
                    {cat.name}
                  </button>
                ))}
              </div>

              {/* 模板列表 */}
              {isLoadingTemplates ? (
                <div className="flex items-center justify-center py-8">
                  <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
                  <span className="ml-2 text-muted-foreground">正在加载模板...</span>
                </div>
              ) : (
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                  {filteredTemplates.map((template) => (
                    <button
                      key={template.id}
                      onClick={() => handleSelectTemplate(template.id)}
                      className={`p-4 rounded-xl border-2 transition-all text-left ${
                        selectedTemplate === template.id
                          ? 'border-primary bg-primary/5'
                          : 'border-border hover:border-primary/50'
                      }`}
                    >
                      <div className="text-3xl mb-2">{template.preview}</div>
                      <p className="font-medium text-foreground text-sm">{template.name}</p>
                      <p className="text-xs text-muted-foreground mt-1">{template.description}</p>
                    </button>
                  ))}
                </div>
              )}

              {/* 模板详情 */}
              {templateDetail && (
                <div className="mt-4 p-4 bg-muted/50 rounded-xl">
                  <div className="flex items-center justify-between mb-2">
                    <h3 className="font-medium text-foreground">
                      已选择：{templateDetail.name}
                    </h3>
                    <button
                      onClick={() => {
                        setSelectedTemplate('');
                        setTemplateDetail(null);
                      }}
                      className="text-muted-foreground hover:text-foreground"
                    >
                      <X className="w-4 h-4" />
                    </button>
                  </div>
                  <p className="text-sm text-muted-foreground mb-2">
                    风格：{templateDetail.structure.style}
                  </p>
                  <p className="text-sm text-muted-foreground">
                    包含章节：{templateDetail.structure.sections.join('、')}
                  </p>
                </div>
              )}
            </div>

            {/* 优化操作 */}
            {selectedTemplate && analysisResult && (
              <div className="bg-card rounded-xl border border-border p-6">
                <h2 className="text-lg font-semibold text-foreground mb-4">AI 优化</h2>
                <p className="text-sm text-muted-foreground mb-4">
                  AI 将根据分析结果和所选模板结构，生成优化后的简历内容
                </p>

                {!optimizedResume ? (
                  <button
                    onClick={handleOptimize}
                    disabled={isOptimizing}
                    className="w-full py-3 bg-primary text-primary-foreground rounded-lg font-medium hover:bg-primary/90 transition-colors disabled:opacity-50 flex items-center justify-center gap-2"
                  >
                    {isOptimizing ? (
                      <>
                        <Loader2 className="w-5 h-5 animate-spin" />
                        AI 正在读取模板并优化...
                      </>
                    ) : (
                      <>
                        <Sparkles className="w-5 h-5" />
                        开始 AI 优化
                      </>
                    )}
                  </button>
                ) : (
                  <div>
                    <div className="flex items-center justify-between mb-4">
                      <h3 className="font-medium text-foreground flex items-center gap-2">
                        <Eye className="w-5 h-5 text-primary" />
                        优化预览
                      </h3>
                      <button
                        onClick={handleOptimize}
                        disabled={isOptimizing}
                        className="text-sm text-muted-foreground hover:text-foreground flex items-center gap-1"
                      >
                        <RefreshCw className="w-4 h-4" />
                        重新生成
                      </button>
                    </div>

                    <div className="bg-muted/50 rounded-xl p-6 mb-4 max-h-96 overflow-y-auto">
                      <pre className="whitespace-pre-wrap text-sm text-foreground font-mono">
                        {optimizedResume}
                      </pre>
                    </div>

                    {/* 导出选项 */}
                    <div className="flex items-center gap-4">
                      <div className="flex-1">
                        <label className="text-sm text-muted-foreground mb-2 block">
                          导出格式
                        </label>
                        <div className="flex gap-2">
                          {(['pdf', 'docx', 'html'] as const).map((format) => (
                            <button
                              key={format}
                              onClick={() => setExportFormat(format)}
                              className={`px-4 py-2 rounded-lg text-sm transition-colors ${
                                exportFormat === format
                                  ? 'bg-primary text-primary-foreground'
                                  : 'bg-muted text-muted-foreground hover:bg-muted/80'
                              }`}
                            >
                              {format.toUpperCase()}
                            </button>
                          ))}
                        </div>
                      </div>
                      <button
                        onClick={handleExport}
                        className="px-6 py-3 bg-green-500 text-white rounded-lg font-medium hover:bg-green-600 transition-colors flex items-center gap-2 self-end"
                      >
                        <Download className="w-5 h-5" />
                        导出简历
                      </button>
                    </div>
                  </div>
                )}
              </div>
            )}

            {!analysisResult && (
              <div className="bg-amber-500/10 border border-amber-500/20 rounded-xl p-4 flex items-start gap-3">
                <AlertCircle className="w-5 h-5 text-amber-500 flex-shrink-0 mt-0.5" />
                <div>
                  <p className="font-medium text-amber-600 mb-1">请先进行简历分析</p>
                  <p className="text-sm text-amber-600/80">
                    AI 需要基于分析结果来优化简历内容。
                    <button
                      onClick={() => setActiveTab('analyze')}
                      className="underline ml-1"
                    >
                      前往分析
                    </button>
                  </p>
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
