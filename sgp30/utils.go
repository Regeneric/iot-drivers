package sgp30

import (
	"strings"
)

func WithLogger(log Logger) Option {
	return func(d *Device) {
		if log != nil {
			d.log = log
		}
	}
}

func Hex8(b uint8) string {
	const hexChars = "0123456789ABCDEF"
	return string([]byte{'0', 'x', hexChars[b>>4], hexChars[b&0x0F]})
}

func Hex16(w uint16) string {
	const hexChars = "0123456789ABCDEF"
	return string([]byte{'0', 'x', hexChars[w>>12], hexChars[(w>>8)&0x0F], hexChars[(w>>4)&0x0F], hexChars[w&0x0F]})
}

func stringSanitize(s string) string {
	s = strings.ToLower(s)
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "_")

	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || (r == '-') || (r == '_') {
			b.WriteRune(r)
		}
	}

	out := b.String()
	if out == "" {
		return "unknown"
	}
	return out
}
