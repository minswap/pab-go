package cli

import "strings"

func isOutputArg(arg string) bool {
	return strings.HasPrefix(arg, "addr_test")
}

func formatOutputArg(arg string) string {
	parts := strings.Split(arg, " + ")
	var sb strings.Builder
	for i, part := range parts {
		if i > 0 {
			sb.WriteString("\n    + ")
		}
		sb.WriteString(part)
	}
	return sb.String()
}

func FormatCLIArgs(args ...string) string {
	var sb strings.Builder
	for _, arg := range args {
		if sb.Len() > 0 {
			if strings.HasPrefix(arg, "--") {
				sb.WriteString("\n  ")
			} else {
				sb.WriteString(" ")
			}
		}
		if isOutputArg(arg) {
			arg = formatOutputArg(arg)
		}
		sb.WriteString(arg)
	}
	return sb.String()
}
