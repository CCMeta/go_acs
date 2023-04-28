package main

import "github.com/kataras/iris/v12"

func shit() {
	println("main start")
	app := iris.New()

	// Routers
	app.Handle("GET", "/", func(ctx iris.Context) {
		ctx.HTML("<h1>fuck12ssss22</h1>")
	})
	app.Handle("GET", "/ping", func(ctx iris.Context) {
		ctx.WriteString("pong")
	})

	// Config done and go run
	app.Run(iris.Addr(":8080"))
}
