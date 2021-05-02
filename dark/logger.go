package dark

import (
	"log"
	"time"
)

func Logger() HandleFunc {
	return func(c *Context) {
		t := time.Now()
		c.Next()
		log.Printf("Middlewarer-Logger:[%d] %s in %v", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}
