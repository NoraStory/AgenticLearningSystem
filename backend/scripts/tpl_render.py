#!/usr/bin/env python3
"""tpl_render.py — 用 docxtpl 渲染已注册模板,输出最终 docx 二进制。

用法(由 Go 后端以子进程调用):
    tpl_render.py <template_path>
渲染数据 JSON 从 stdin 读取:
    {"sections": [{"title": "基本信息", "items": ["张三"]}, ...]}
数据按 sections 数组顺序填充模板中的 {{ sections[i].items[j] }} 占位符;
模板中的 {%tr %}(表格行循环)依赖 docxtpl 的 Jinja2 支持。

渲染结果写到临时文件后,其二进制输出到 stdout(stdout 只含文件字节,无日志)。
"""
import json
import os
import subprocess
import sys
import tempfile

import docxtpl

# Windows 下 stdin 默认 GBK 编码,Go 子进程以 UTF-8 交换数据,必须统一
sys.stdin.reconfigure(encoding="utf-8")


def main() -> int:
    if len(sys.argv) != 2:
        print("usage: tpl_render.py <template_path>", file=sys.stderr)
        return 2
    template_path = sys.argv[1]
    if not os.path.isfile(template_path):
        print(f"template not found: {template_path}", file=sys.stderr)
        return 1
    try:
        payload = json.load(sys.stdin)
    except Exception as exc:
        print(f"invalid stdin json: {exc}", file=sys.stderr)
        return 2
    sections = payload.get("sections") or []
    if not sections:
        print("sections must not be empty", file=sys.stderr)
        return 2

    try:
        tpl = docxtpl.DocxTemplate(template_path)
        tpl.render({"sections": sections})
        fd, tmp = tempfile.mkstemp(suffix=".docx")
        os.close(fd)
        tpl.save(tmp)
        with open(tmp, "rb") as fh:
            sys.stdout.buffer.write(fh.read())
        sys.stdout.buffer.flush()
        os.unlink(tmp)
    except Exception as exc:
        print(f"render failed: {exc}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
