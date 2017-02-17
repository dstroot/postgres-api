package main

func main() {
	initialize()
	a := App{}
	a.Initialize(cfg.SQL.User, cfg.SQL.Password, cfg.SQL.Database)
	a.Run(cfg.Port)
}
