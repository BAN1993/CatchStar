package flog

import (
	"runtime"
	"strings"
)

func GetCaller(callDepth int, suffixesToIgnore ...string) (file string, line int) {
	callDepth++
outer:
	for {
		var ok bool
		_, file, line, ok = runtime.Caller(callDepth)
		if !ok {
			file = "[Unknown]"
			line = 0
			break
		}

		for _, s := range suffixesToIgnore {
			if strings.HasSuffix(file, s) {
				callDepth++
				continue outer
			}
		}
		break
	}
	return
}

func GetCallerIgnoringLogMulti(callDepth int) (string, int) {
	return GetCaller(callDepth + 1,
		"logrus/hooks.go",
		"logrus/entry.go",
		"logrus/logger.go",
		"logrus/exported.go",
		"asm_amd64.s")
}

