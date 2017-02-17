package main

import "log"

func main() {
	a := App{}

	err := a.Initialize()
	if err != nil {
		log.Fatal(err)
	}
	defer a.DB.Close()

	a.Run(cfg.Port)
}
