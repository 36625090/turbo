//go:generate go run  main.go --app example --config config.hcl --log.level trace  --log.console --log.path logs  --http.path=/example --ui  --http.port 8081
package main
import (
	"github.com/36625090/turbo"
	_ "github.com/36625090/turbo"
	"github.com/36625090/turbo/example/services/account/controller"
	"github.com/36625090/turbo/logical"
	"github.com/36625090/turbo/option"
	"github.com/36625090/turbo/server"
	"github.com/36625090/turbo/transport"
	"github.com/hashicorp/go-hclog"
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
		//do something
		inv.InitializeSigner(&signer{ctx.Logger})
	})

	if err := inv.Start(); err != nil {
		log.Fatal(err)
		return
	}

}

type signer struct {
	hclog.Logger
}

func (s signer) Sign(keyId string, resp transport.Codec) (string, error) {
	s.Info("signing", "id", keyId, "resp", resp)
	return "11111111111111111111", nil
}

func (s signer) Verify(keyId, sign string, req transport.Codec) error {
	s.Info("signing", "id", keyId, "sign",sign,"req", req)
	return nil
}
