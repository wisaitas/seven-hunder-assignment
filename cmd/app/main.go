package main

import "github.com/7-solutions/backend-challenge/internal/app/initial"

func main() {
	app := initial.New()

	app.Run()

	app.Close()
}
