package main

import "strings"

// Channel kanal bilgilerini tutar
type Channel struct {
	Name  string
	URL   string
	Logo  string
	Group string
}

// ParseM3U bir M3U/M3U8 içeriğini parse eder, kanalları ve kategorileri döner
func ParseM3U(lines []string) ([]Channel, []string) {
	var channels []Channel
	catSet := make(map[string]bool)

	var name, logo, group string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#EXTINF") {
			// Kanal adı: son virgülden sonra
			if idx := strings.LastIndex(line, ","); idx >= 0 {
				name = strings.TrimSpace(line[idx+1:])
			}
			logo = getAttr(line, "tvg-logo")
			group = getAttr(line, "group-title")
		} else if strings.HasPrefix(line, "http") || strings.HasPrefix(line, "rtsp") || strings.HasPrefix(line, "rtmp") {
			grp := group
			if grp == "" {
				grp = "Diğer"
			}
			channels = append(channels, Channel{
				Name:  name,
				URL:   line,
				Logo:  logo,
				Group: grp,
			})
			catSet[grp] = true
			// Sıfırla
			name, logo, group = "", "", ""
		}
	}

	// Kategorileri sıralı listeye çevir
	cats := make([]string, 0, len(catSet))
	for c := range catSet {
		cats = append(cats, c)
	}
	sortStrings(cats)

	return channels, cats
}

// getAttr bir EXTINF satırından attr="değer" formatındaki değeri çeker
func getAttr(line, attr string) string {
	key := attr + `="`
	idx := strings.Index(line, key)
	if idx < 0 {
		return ""
	}
	start := idx + len(key)
	end := strings.Index(line[start:], `"`)
	if end < 0 {
		return ""
	}
	return line[start : start+end]
}

// sortStrings basit sıralama (sort paketi kullanmadan, küçük listeler için)
func sortStrings(s []string) {
	for i := 0; i < len(s); i++ {
		for j := i + 1; j < len(s); j++ {
			if s[i] > s[j] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}
