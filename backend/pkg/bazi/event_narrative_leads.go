package bazi

import "strings"

// changeLead 按 primary 信号子类型/极性/Source 选择 change 主题前导句，避免跨年模板碰撞。
func changeLead(p EventSignal) string {
	switch p.Type {
	case "伏吟":
		return "这一年旧事容易卷土重来，过去搁置的处理可能重新冒头"
	case "反吟":
		return "这一年节奏变化会比较突兀，环境、计划或关系都可能临时倒挂"
	case "大运合化":
		return "这一年大运能量被牵动重组，方向感会有一次明显的调整"
	case TypeJuShiZhong:
		return "这一年整体力量容易被放大，一个选择容易牵动多条线"
	}
	if p.Polarity == PolarityXiong {
		if p.Source == SourceXing || strings.Contains(p.Evidence, "刑") {
			return "这一年容易在反复和细节里消耗，问题未必爆发却拖出余波"
		}
		return "这一年的变动感比较强，旧问题或突发调整容易被推到眼前"
	}
	if p.Polarity == PolarityJi {
		return "这一年变动中带着调整空间，主动顺势比被动应对更省力"
	}
	return "这一年节奏不算稳定，但调整中容易找到新方向"
}

// healthLead 按 Evidence/Polarity 选择 health 主题前导句。
func healthLead(p EventSignal) string {
	if strings.Contains(p.Evidence, "冲") || strings.Contains(p.Evidence, "白虎") {
		return "这一年身体和安全节奏需要被前置考虑，意外性消耗要避免"
	}
	if p.Polarity == PolarityXiong {
		return "健康、安全或日常节奏是这一年的主线，压力点会比较直接"
	}
	return "这一年身心提醒会更频繁，作息节律值得重新校准"
}

// relationshipLead 按 primary.Type 选择 relationship 主题前导句。
func relationshipLead(p EventSignal) string {
	switch p.Type {
	case "婚恋_合":
		return "这一年人际或感情的靠近感增强，关系节奏容易加快"
	case "婚恋_冲":
		return "这一年关系、距离和承诺容易被检验，节奏可能出现明显波动"
	case "婚恋_变":
		return "这一年情感或合作的方向容易调整，分寸感和边界都会被试探"
	case TypeXingGeQingYi:
		return "这一年情绪表达和人际反应会更外露，主动沟通比闷着推进有效"
	case TypeXingGePanNi:
		return "这一年个性主张容易和外部要求碰撞，关键节点上要稳住态度"
	}
	return "人际、感情或家庭沟通是这一年的主线，情绪触发会比较明显"
}

// defaultHardLead 按 primary.Source 选择硬事件兜底前导句。
func defaultHardLead(p EventSignal) string {
	switch p.Source {
	case SourceKongwang:
		return "这一年带着虚而不实的不稳定感，承诺和计划要多确认细节"
	case SourceXing:
		return "这一年有内耗反复的影子，事情未必爆发但容易消耗心力"
	case SourceFuyin:
		return "这一年旧主题反复回头，过去没处理完的事情会再被推上来"
	case SourceHehua:
		return "这一年大运能量被牵动，方向上的关键节点会比预想更明显"
	}
	return "这一年不是完全平稳的年份，关键事件会比平时更容易显形"
}
