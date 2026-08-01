#!/usr/bin/env python3
"""BKT 参数离线拟合脚本（纯标准库，零第三方依赖）。

用法：python bkt_fit.py < data.json
输入：{"sequences":[{"category":"algorithm","skills":[{"skill":"两数之和","obs":[1,0,1,1]}]}]}
输出：{"fitted":[{"category":"algorithm","prior":0.13,"transition":0.22,"guess":0.3,"slip":0.08,"n_skills":12,"n_obs":87,"loglik":-31.2}]}

用 EM 算法（前向-后向）对 2 状态 BKT 模型做极大似然拟合：
  状态 L=0（未掌握）/ L=1（已掌握）
  P(L0=1)=prior, P(learn)=transition, P(guess)=guess, P(slip)=slip
观测序列 obs∈{0,1}：做题结果（1=正确），由 Go 端从 submissions 构建。
全部数值计算在 log 域（log-sum-exp），长序列不溢出。
"""

import json
import math
import sys


def clamp(v, lo=0.01, hi=0.99):
    return max(lo, min(hi, v))


def lse(a, b):
    """log(exp(a) + exp(b))，防下溢。"""
    m = max(a, b)
    return m + math.log(math.exp(a - m) + math.exp(b - m))


def fit_skill(obs, prior, transition, guess, slip, max_iter=200, tol=1e-6):
    """对单个技能序列做 EM 拟合，返回 (prior, transition, guess, slip, loglik)。"""
    T = len(obs)
    NEG_INF = float("-inf")

    def obs_log_prob(l, o):
        # P(o=1|l=0)=guess, P(o=1|l=1)=1-slip
        if o == 1:
            return math.log(guess) if l == 0 else math.log(1 - slip)
        return math.log(1 - guess) if l == 0 else math.log(slip)

    def trans_log_prob(l, l_next):
        # BKT 转移矩阵 [[1-T, T], [0, 1]]：
        # 未掌握 → 学会概率 T；已掌握保持掌握（不会忘记）
        if l == 1:
            return 0.0 if l_next == 1 else NEG_INF
        return math.log(transition) if l_next == 1 else math.log(1 - transition)

    for _ in range(max_iter):
        # ---- E 步：前向-后向（log 域）----
        alpha_log = [[0.0, 0.0] for _ in range(T)]
        alpha_log[0][0] = math.log(1 - prior) + obs_log_prob(0, obs[0])
        alpha_log[0][1] = math.log(prior) + obs_log_prob(1, obs[0])
        for t in range(1, T):
            for l in range(2):
                acc = alpha_log[t - 1][0] + trans_log_prob(0, l) + obs_log_prob(l, obs[t])
                acc2 = alpha_log[t - 1][1] + trans_log_prob(1, l) + obs_log_prob(l, obs[t])
                alpha_log[t][l] = lse(acc, acc2)
        # 后向：β_T = 1
        beta_log = [[0.0, 0.0] for _ in range(T)]
        for t in range(T - 2, -1, -1):
            for l in range(2):
                acc = trans_log_prob(l, 0) + obs_log_prob(0, obs[t + 1]) + beta_log[t + 1][0]
                acc2 = trans_log_prob(l, 1) + obs_log_prob(1, obs[t + 1]) + beta_log[t + 1][1]
                beta_log[t][l] = lse(acc, acc2)

        # γ_t(l) = P(L_t=l | obs)（归一化）
        gamma = [[0.0, 0.0] for _ in range(T)]
        for t in range(T):
            a = alpha_log[t][0] + beta_log[t][0]
            b = alpha_log[t][1] + beta_log[t][1]
            z = lse(a, b)
            gamma[t][0] = math.exp(a - z)
            gamma[t][1] = math.exp(b - z)

        # ξ_t(l,l') = P(L_t=l, L_{t+1}=l' | obs)，按 Σ_{l,l'} 归一化与 γ 同尺度
        xi = [[[0.0, 0.0] for _ in range(2)] for _ in range(T - 1)]
        for t in range(T - 1):
            raw = [[0.0, 0.0] for _ in range(2)]
            for l in range(2):
                for l_next in range(2):
                    tp = trans_log_prob(l, l_next)
                    if tp == NEG_INF:
                        raw[l][l_next] = 0.0
                    else:
                        raw[l][l_next] = math.exp(
                            alpha_log[t][l] + tp
                            + obs_log_prob(l_next, obs[t + 1]) + beta_log[t + 1][l_next]
                        )
            z = raw[0][0] + raw[0][1] + raw[1][0] + raw[1][1]
            if z > 0:
                for l in range(2):
                    for l_next in range(2):
                        xi[t][l][l_next] = raw[l][l_next] / z

        # ---- M 步：参数更新 ----
        new_prior = gamma[0][1]
        # transition = P(0→1) 的加权估计：Σξ_t(0,1) / Σγ_t(0)（t=0..T-2）
        num_tr = sum(xi[t][0][1] for t in range(T - 1))
        den_tr = sum(gamma[t][0] for t in range(T - 1))
        new_transition = num_tr / den_tr if den_tr > 0 else transition
        num_g = sum(gamma[t][0] * (1 if obs[t] == 1 else 0) for t in range(T))
        den_g = sum(gamma[t][0] for t in range(T))
        new_guess = num_g / den_g if den_g > 0 else guess
        num_s = sum(gamma[t][1] * (1 if obs[t] == 0 else 0) for t in range(T))
        den_s = sum(gamma[t][1] for t in range(T))
        new_slip = num_s / den_s if den_s > 0 else slip

        new_prior = clamp(new_prior)
        new_transition = clamp(new_transition)
        new_guess = clamp(new_guess)
        new_slip = clamp(new_slip)
        if new_guess + new_slip > 0.9:
            # 退化约束：guess+slip 过大说明模型接近随机猜
            k = 0.9 / (new_guess + new_slip)
            new_guess *= k
            new_slip *= k

        delta = max(abs(new_prior - prior), abs(new_transition - transition),
                    abs(new_guess - guess), abs(new_slip - slip))
        prior, transition, guess, slip = new_prior, new_transition, new_guess, new_slip
        if delta < tol:
            break

    # 数据似然：log P(obs) = log Σ_l α_T(l)（前向算法最后一步）
    loglik = lse(alpha_log[T - 1][0], alpha_log[T - 1][1])
    return prior, transition, guess, slip, loglik


def fit_category(seqs, prior, transition, guess, slip, max_iter=200, tol=1e-6):
    """对同分类下的多个技能序列联合做 EM 拟合（共享一组参数）。

    E 步对每个序列分别跑前向-后向，把各序列的充分统计量（gamma/xi）求和；
    M 步用汇总统计量更新参数。比逐序列独立拟合更稳定、更符合
    「该分类共享 BKT 参数」的建模语义。
    """
    for _ in range(max_iter):
        sum_gamma0 = sum_gamma1 = 0.0
        sum_xi01 = 0.0
        sum_g_obs = sum_g_miss = 0.0
        sum_s_obs = sum_s_miss = 0.0
        total_loglik = 0.0
        for obs in seqs:
            T = len(obs)
            NEG_INF = float("-inf")

            def obs_log_prob(l, o):
                if o == 1:
                    return math.log(guess) if l == 0 else math.log(1 - slip)
                return math.log(1 - guess) if l == 0 else math.log(slip)

            def trans_log_prob(l, l_next):
                if l == 1:
                    return 0.0 if l_next == 1 else NEG_INF
                return math.log(transition) if l_next == 1 else math.log(1 - transition)

            alpha_log = [[0.0, 0.0] for _ in range(T)]
            alpha_log[0][0] = math.log(1 - prior) + obs_log_prob(0, obs[0])
            alpha_log[0][1] = math.log(prior) + obs_log_prob(1, obs[0])
            for t in range(1, T):
                for l in range(2):
                    acc = alpha_log[t - 1][0] + trans_log_prob(0, l) + obs_log_prob(l, obs[t])
                    acc2 = alpha_log[t - 1][1] + trans_log_prob(1, l) + obs_log_prob(l, obs[t])
                    alpha_log[t][l] = lse(acc, acc2)
            beta_log = [[0.0, 0.0] for _ in range(T)]
            for t in range(T - 2, -1, -1):
                for l in range(2):
                    acc = trans_log_prob(l, 0) + obs_log_prob(0, obs[t + 1]) + beta_log[t + 1][0]
                    acc2 = trans_log_prob(l, 1) + obs_log_prob(1, obs[t + 1]) + beta_log[t + 1][1]
                    beta_log[t][l] = lse(acc, acc2)

            gamma = [[0.0, 0.0] for _ in range(T)]
            for t in range(T):
                a = alpha_log[t][0] + beta_log[t][0]
                b = alpha_log[t][1] + beta_log[t][1]
                z = lse(a, b)
                gamma[t][0] = math.exp(a - z)
                gamma[t][1] = math.exp(b - z)

            # 充分统计量
            sum_gamma0 += sum(gamma[t][0] for t in range(T))
            sum_gamma1 += sum(gamma[t][1] for t in range(T))
            for t in range(T - 1):
                # ξ_t(0,1)：未掌握 → 学会
                x01 = math.exp(
                    alpha_log[t][0] + math.log(transition)
                    + obs_log_prob(1, obs[t + 1]) + beta_log[t + 1][1]
                )
                z = (math.exp(alpha_log[t][0] + math.log(1 - transition) + obs_log_prob(0, obs[t + 1]) + beta_log[t + 1][0])
                     + x01 + math.exp(alpha_log[t][1] + obs_log_prob(1, obs[t + 1]) + beta_log[t + 1][1]))
                if z > 0:
                    sum_xi01 += x01 / z
            for t in range(T):
                if obs[t] == 1:
                    sum_g_obs += gamma[t][0]
                else:
                    sum_g_miss += gamma[t][0]
                if obs[t] == 1:
                    sum_s_miss += gamma[t][1]
                else:
                    sum_s_obs += gamma[t][1]
            total_loglik += lse(alpha_log[T - 1][0], alpha_log[T - 1][1])

        # ---- M 步 ----
        new_prior = sum_gamma1 / (sum_gamma0 + sum_gamma1) if (sum_gamma0 + sum_gamma1) > 0 else prior
        new_transition = sum_xi01 / sum_gamma0 if sum_gamma0 > 0 else transition
        new_guess = sum_g_obs / (sum_g_obs + sum_g_miss) if (sum_g_obs + sum_g_miss) > 0 else guess
        new_slip = sum_s_obs / (sum_s_obs + sum_s_miss) if (sum_s_obs + sum_s_miss) > 0 else slip

        new_prior = clamp(new_prior)
        new_transition = clamp(new_transition)
        new_guess = clamp(new_guess)
        new_slip = clamp(new_slip)
        if new_guess + new_slip > 0.9:
            k = 0.9 / (new_guess + new_slip)
            new_guess *= k
            new_slip *= k

        delta = max(abs(new_prior - prior), abs(new_transition - transition),
                    abs(new_guess - guess), abs(new_slip - slip))
        prior, transition, guess, slip = new_prior, new_transition, new_guess, new_slip
        if delta < tol:
            break
    return prior, transition, guess, slip, total_loglik


def main():
    sys.stdout.reconfigure(encoding="utf-8")
    data = json.load(sys.stdin)
    seqs = data.get("sequences", [])
    fitted = []
    for group in seqs:
        category = group.get("category", "global")
        skills = group.get("skills", [])
        obs_list = []
        for sk in skills:
            obs = [1 if o else 0 for o in sk.get("obs", [])]
            if len(obs) < 3:
                continue  # 序列太短，拟合不稳定
            obs_list.append(obs)
        if not obs_list:
            continue
        prior, transition, guess, slip, loglik = fit_category(obs_list, 0.1, 0.3, 0.25, 0.1)
        fitted.append({
            "category": category,
            "prior": round(prior, 4),
            "transition": round(transition, 4),
            "guess": round(guess, 4),
            "slip": round(slip, 4),
            "n_skills": len(obs_list),
            "n_obs": sum(len(o) for o in obs_list),
            "loglik": round(loglik, 2),
        })
    print(json.dumps({"fitted": fitted}, ensure_ascii=False))


if __name__ == "__main__":
    main()
