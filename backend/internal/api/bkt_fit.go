package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	"codeforge/backend/internal/model"
	"github.com/gin-gonic/gin"
)

// bktFitScript 拟合脚本路径（backend/scripts/bkt_fit.py）。
// 运行时按工作目录依次探测，与 loadSystemPrompt 的多路径模式一致。
var bktFitScriptPaths = []string{
	"scripts/bkt_fit.py",
	"backend/scripts/bkt_fit.py",
	filepath.Join("backend", "scripts", "bkt_fit.py"),
}

// bktFitInput 发送给脚本的输入结构。
type bktFitInput struct {
	Sequences []bktFitGroup `json:"sequences"`
}

type bktFitGroup struct {
	Category string       `json:"category"`
	Skills   []bktFitSkill `json:"skills"`
}

type bktFitSkill struct {
	Skill string `json:"skill"`
	Obs   []int  `json:"obs"`
}

// bktFitResult 脚本输出的单组拟合结果。
type bktFitResult struct {
	Category   string  `json:"category"`
	Prior      float64 `json:"prior"`
	Transition float64 `json:"transition"`
	Guess      float64 `json:"guess"`
	Slip       float64 `json:"slip"`
	NSkills    int     `json:"n_skills"`
	NObs       int     `json:"n_obs"`
	Loglik     float64 `json:"loglik"`
}

// runBKTFit 从 submissions 构建做题序列 → 调用 Python 脚本拟合 → 写回 bkt_params。
// 手动触发（不常驻、不加 cron）；结果立即生效（refreshBktCache）。
func (s *Server) runBKTFit(c *gin.Context) {
	// 1. 按 (user_id, problem_id) 分组取提交序列，升序
	type sub struct {
		UserID     string
		ProblemID  uint
		Status     string
		CreatedAt  time.Time
	}
	var rows []sub
	s.services.DB.Model(&model.Submission{}).
		Select("user_id, problem_id, status, created_at").
		Order("created_at asc").Find(&rows)
	if len(rows) == 0 {
		fail(c, 400, 400, "暂无提交记录，无法拟合")
		return
	}
	// 2. 按 problem_id 聚合序列 + 关联 problem.category
	seqByProblem := map[uint][]int{}
	catByProblem := map[uint]string{}
	var problems []model.Problem
	s.services.DB.Find(&problems)
	for _, p := range problems {
		catByProblem[p.ID] = p.Category
	}
	for _, r := range rows {
		obs := 0
		if r.Status == "accepted" {
			obs = 1
		}
		seqByProblem[r.ProblemID] = append(seqByProblem[r.ProblemID], obs)
	}
	// 3. 按 category 归组
	groups := map[string]*bktFitGroup{}
	for pid, obs := range seqByProblem {
		cat := catByProblem[pid]
		if cat == "" {
			cat = "global"
		}
		g, ok := groups[cat]
		if !ok {
			g = &bktFitGroup{Category: cat}
			groups[cat] = g
		}
		g.Skills = append(g.Skills, bktFitSkill{Skill: fmt.Sprintf("problem-%d", pid), Obs: obs})
	}
	// 4. 调脚本（stdin 传 JSON，stdout 收结果）
	script := ""
	for _, p := range bktFitScriptPaths {
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			script = p
			break
		}
	}
	if script == "" {
		fail(c, 500, 500, "找不到拟合脚本 scripts/bkt_fit.py")
		return
	}
	payload, _ := json.Marshal(bktFitInput{Sequences: sortedGroups(groups)})
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "python", script)
	cmd.Stdin = bytes.NewReader(payload)
	out, err := cmd.Output()
	if err != nil {
		fail(c, 500, 500, "拟合脚本执行失败: "+err.Error())
		return
	}
	var res struct {
		Fitted []bktFitResult `json:"fitted"`
	}
	if json.Unmarshal(bytes.TrimSpace(out), &res) != nil {
		fail(c, 500, 500, "拟合脚本输出解析失败")
		return
	}
	// 5. 写回 bkt_params
	written := 0
	for _, f := range res.Fitted {
		row := model.BktParam{Category: f.Category}
		if err := s.services.DB.Where("category = ?", f.Category).FirstOrCreate(&row, model.BktParam{Category: f.Category}).Error; err != nil {
			continue
		}
		s.services.DB.Model(&row).Updates(map[string]any{
			"prior": f.Prior, "transition": f.Transition, "guess": f.Guess, "slip": f.Slip,
			"source": "fitted", "sample_size": f.NObs,
		})
		written++
	}
	s.refreshBktCache() // 立即生效
	success(c, gin.H{"written": written, "fitted": res.Fitted})
}

func sortedGroups(m map[string]*bktFitGroup) []bktFitGroup {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]bktFitGroup, 0, len(keys))
	for _, k := range keys {
		out = append(out, *m[k])
	}
	return out
}

// bktParams 返回当前生效的 BKT 参数（排查用）。
func (s *Server) bktParams(c *gin.Context) {
	s.refreshBktCache()
	bktCache.mu.RLock()
	items := bktCache.items
	bktCache.mu.RUnlock()
	if items == nil {
		items = map[string]bktParams{}
	}
	// 转成可读结构（含兜底 global）
	data := gin.H{}
	for k, v := range items {
		data[k] = gin.H{"prior": v.Prior, "transition": v.Transition, "guess": v.Guess, "slip": v.Slip}
	}
	success(c, data)
}
