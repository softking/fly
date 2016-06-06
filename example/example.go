package main

import (
	"github.com/softking/fly"
)

// HelloPre pre
func HelloPre(c *fly.Context) {
	c.WriteString("pre \n")

}

// Hello hello
func Hello(c *fly.Context)  {

	c.WriteString("hello")
	c.WriteString(c.Param["name"] + "\n")
	c.Abort()
}

// HelloAfter after
func HelloAfter(c *fly.Context)  {
	c.WriteString("after \n")
}

// Mid midware
func Mid(c *fly.Context)  {
	c.WriteString("midpre\n")
	c.Next()
	c.WriteString("midafter\n")

}

func main() {
	router := fly.IWillFly()

	router.MidWare(Mid)

	router.GET("/hello/:name", HelloPre, Hello, HelloAfter)

	fly.ReloadRun(router, ":2222")

}
