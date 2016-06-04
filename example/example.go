package main

import (
	"github.com/softking/fly"
)

// HelloPre pre
func HelloPre(c *fly.Context) bool {
	c.WriteString("pre")
	c.WriteString(c.Param["name"] + "\n")
	return true
}

// Hello hello
func Hello(c *fly.Context) bool {

	c.WriteString("hello")
	c.WriteString(c.Param["name"] + "\n")
	return true
}

// HelloAfter after
func HelloAfter(c *fly.Context) bool {
	c.WriteString("after")
	c.WriteString(c.Param["name"] + "\n")
	return true
}

// Mid midware
func Mid(c *fly.Context) bool {
	c.WriteString("midpre\n")
	c.Next()
	c.WriteString("midafter\n")

	return true
}

func main() {
	router := fly.IWillFly()

	router.MidWare(Mid)

	router.GET("/hello/:name", HelloPre, Hello, HelloAfter)

	fly.ReloadRun(router, ":2222")

}
