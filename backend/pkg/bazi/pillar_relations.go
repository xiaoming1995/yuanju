package bazi

// GanFiveHe reports whether two stems form one of the traditional five combinations.
// It does not imply that the combination transforms into another element.
func GanFiveHe(a, b string) bool {
	_, ok := ganWuhe[[2]string{a, b}]
	return ok
}

// ZhiSixHe reports a traditional six-combination between two branches.
// It does not imply transformation or a favorable event.
func ZhiSixHe(a, b string) bool {
	return a != "" && b != "" && sixHe[a] == b
}

// ZhiSixChong reports a traditional six-clash between two branches.
func ZhiSixChong(a, b string) bool {
	return a != "" && b != "" && sixChong[a] == b
}
