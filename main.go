package main

import "log"

func main() {
	a := App{}
	err := a.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	a.Run(cfg.Port)
}
