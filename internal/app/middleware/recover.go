package middleware

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func Recover() fiber.Handler {
	return recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e any) {
			traceId := c.Request().Header.Peek("X-Trace-Id")
			fmt.Printf("\n========== PANIC RECOVERED ==========\n"+
				"time     : %s\n"+
				"traceId  : %s\n"+
				"method   : %s\n"+
				"path     : %s\n"+
				"ip       : %s\n"+
				"userAgent: %s\n"+
				"error    : %v\n"+
				"-------- stack trace --------\n%s\n"+
				"===============================================================================\n",
				time.Now().Format(time.RFC3339),
				traceId,
				c.Method(),
				c.Path(),
				c.IP(),
				c.Get(fiber.HeaderUserAgent),
				e,
				string(debug.Stack()),
			)
		},
	})
}
