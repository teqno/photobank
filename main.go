package main

import "example.com/photobank/router"

func main() {
	r := router.SetupRouter() 
	r.Run(":8080")
}
