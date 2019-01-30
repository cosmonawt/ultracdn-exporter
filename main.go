package main

import "log"

func main() {
	c := client{}
	c.login("", "")
	log.Printf("%+v", c)
}
