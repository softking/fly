package main

import (
	"github.com/softking/fly"
	"github.com/softking/fly/midware"
)


type name struct{
	Name string
	ID int
}

// HelloPre pre
func HelloPre(c *fly.Context) {
	c.WriteString(200, "pre \n")

}

// Hello hello
func Hello(c *fly.Context) {
	a := c.Query("cao")
	c.WriteString(200, "hello   ")
	c.WriteString(200, a+"   ")
	c.WriteString(200, c.Param("name")+"\n")
}

// HelloAfter after
func HelloAfter(c *fly.Context) {
	c.WriteString(200, "after \n")
}

// Mid midware
func Mid(c *fly.Context) {
	c.WriteString(200, "midpre\n")
	c.Next()
	c.WriteString(200, "midafter\n")
}


func Json(c *fly.Context){
	c.WriteJSON(200, name{Name:"lei", ID:123})
}

func main() {
	router := fly.IWillFly()

	router.MidWare(midware.Logger, midware.Recovery)
	router.AddMidware(Mid)

	router.GET("/hello/:name", HelloPre, Hello, HelloAfter)

	router.GET("/json", Json)

	fly.ReloadRun(router, ":8888")

}
