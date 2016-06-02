package test

import (
	"time"
	"testing"
	"github.com/softking/fly"
)

func pre(c *fly.Context) bool {
	c.WriteString("pre\n")
	return true
}

func after(c *fly.Context) bool {
	c.WriteString("after\n")
	return true
}

func hello(c *fly.Context) bool {
	time.Sleep(20 * time.Second)
	data, ok := c.GetParam("msg")
	if ok {
		c.WriteString(data + "\n")
	} else {
		c.WriteString("error\n")
	}
	return true
}

func mid1(c *fly.Context) bool {
	c.WriteString("111\n")
	c.Next()
	c.WriteString("222\n")
	return true
}

func mid2(c *fly.Context) bool {
	c.WriteString("333\n")
	c.WriteString("444\n")
	return true
}

func Test_Fly(t *testing.T) {
	f := fly.IWillFly()
	f.Midware(mid1, mid2)
	f.Get("/hello", pre, hello, after)
	f.POST("/hello", pre, hello, after)
	f.RunReload()

}

