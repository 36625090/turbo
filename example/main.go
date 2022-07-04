package main

import (
	"github.com/36625090/turbo"
	_ "github.com/36625090/turbo"
	"github.com/36625090/turbo/example/services/account/controller"
	"github.com/36625090/turbo/logical"
	"github.com/36625090/turbo/option"
	"github.com/36625090/turbo/server"
	"log"
	"os"
)

func init() {

}

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	opts, err := option.NewOptions()
	if err != nil {
		os.Exit(1)
	}

	factories := map[string]logical.Factory{
		"account": controller.Factory,
	}

	inv, err := turbo.Default(opts, factories)

	if err != nil {
		log.Fatal(err)
		return
	}

	inv.Initialize(func(ctx *server.TurboContext) {

	})

	if err := inv.Start(); err != nil {
		log.Fatal(err)
		return
	}

}
