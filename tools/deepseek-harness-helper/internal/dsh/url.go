package dsh

func ParseWebURL(line string) (string, bool) {
	matches := webURLPattern.FindStringSubmatch(stringTrim(line))
	if matches == nil {
		return "", false
	}
	return matches[1], true
}
