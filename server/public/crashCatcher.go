package public

import (
	"DrillServerGo/flog"
	"fmt"
	"os"
	"runtime/debug"
)

/**
 * 捕获崩溃并打印日志
 * 只能捕获当前goroutine,所以需要在每个goroutine加上
 */
func CrashCatcher () {
	errs := recover()
	if errs == nil {
		return
	}

	str := fmt.Sprintf("===Catch a crash===[pid:%d]\n", os.Getpid())
	str = fmt.Sprintf("%s%s\n", str, "=== Error ===")
	str = fmt.Sprintf("%s%s\n", str, errs)
	str = fmt.Sprintf("%s%s\n", str, "=== Stack ===")
	str = fmt.Sprintf("%s%s", str, string(debug.Stack()))

	flog.GetInstance().Fatalf(str)
}
