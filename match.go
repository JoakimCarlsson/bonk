package bonk

func globMatch(pattern, s string) bool {
	px, sx := 0, 0
	nextPx, nextSx := 0, -1

	for sx < len(s) {
		if px < len(pattern) {
			switch pattern[px] {
			case '*':
				if px+1 < len(pattern) && pattern[px+1] == '*' {
					px += 2
					if px < len(pattern) && pattern[px] == '/' {
						px++
					}
					nextPx = px
					nextSx = sx
					continue
				}
				nextPx = px
				nextSx = sx + 1
				px++
				continue
			case '?':
				px++
				sx++
				continue
			default:
				if pattern[px] == s[sx] {
					px++
					sx++
					continue
				}
			}
		}

		if nextSx >= 0 && nextSx <= len(s) {
			px = nextPx
			sx = nextSx
			nextSx++
			continue
		}

		return false
	}

	for px < len(pattern) && pattern[px] == '*' {
		px++
	}

	return px == len(pattern)
}
