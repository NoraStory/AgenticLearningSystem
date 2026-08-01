package api

import (
	"math"
	"testing"
)

func TestBktUpdateKnownSequences(t *testing.T) {
	p := defaultBkt
	// 做对一次：P(L|correct)=0.2857（含 P(S)=0.1,P(G)=0.25），+ 学习转移 0.3 → 0.5
	m := bktUpdate(p, p.Prior, true)
	expected := 0.5
	if math.Abs(m-expected) > 0.01 {
		t.Errorf("做对一次 = %v, want ≈%v", m, expected)
	}
	// 连续做对单调上升（未达 clamp 上界前）
	prev := m
	for i := 0; i < 3; i++ {
		m = bktUpdate(p, m, true)
		if m <= prev {
			t.Fatalf("第 %d 次做对后 mastery 未上升: %v → %v", i+2, prev, m)
		}
		prev = m
	}
	// 做错后掌握度低于做对后（同先验对比，学习转移对两者都有）
	wrong := bktUpdate(p, p.Prior, false)
	right := bktUpdate(p, p.Prior, true)
	if wrong >= right {
		t.Errorf("做错后掌握度应低于做对后: wrong=%v right=%v", wrong, right)
	}
	// clamp 上界：连对多次后不超过 0.99
	m = 0.5
	for i := 0; i < 100; i++ {
		m = bktUpdate(p, m, true)
	}
	if m > 0.99 {
		t.Errorf("mastery 超过 clamp 上界: %v", m)
	}
	// clamp 下界
	if m < 0.01 {
		t.Errorf("mastery 低于 clamp 下界: %v", m)
	}
}

func TestBktUpdateParamsSensitivity(t *testing.T) {
	// 同序列下，slip 高 → 做对后上升更少（更可能靠猜）
	low := bktParams{Prior: 0.3, Transition: 0.3, Guess: 0.25, Slip: 0.05}
	high := bktParams{Prior: 0.3, Transition: 0.3, Guess: 0.25, Slip: 0.4}
	lo := bktUpdate(low, low.Prior, true)
	hi := bktUpdate(high, high.Prior, true)
	if lo <= hi {
		t.Errorf("slip 低应上升更多: low=%v high=%v", lo, hi)
	}
	// 参数真正生效：相同先验、不同参数得到不同结果
	if math.Abs(lo-hi) < 0.01 {
		t.Errorf("参数对结果无影响: %v vs %v", lo, hi)
	}
}
