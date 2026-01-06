package gee

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

func Recovery() Handler {
	return func(c *Context) {
		defer func() {
			if r := recover(); r != nil {
				panicMsg := fmt.Sprintf("%s", r)
				log.Printf("%s\n\n", trace(panicMsg))
				c.Fail(http.StatusInternalServerError, "Internal server error")
			}
		}()

		c.Next()
	}
}

// print stack trace for debug
func trace(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:]) // skip first 3 caller

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}
