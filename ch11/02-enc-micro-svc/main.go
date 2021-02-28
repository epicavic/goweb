package main

import (
	"fmt"

	"enc-micro-svc/handlers"
	proto "enc-micro-svc/proto"

	micro "github.com/asim/go-micro/v3"
)

func main() {
	// Create a new service. Optionally include some options here.
	service := micro.NewService(
		micro.Name("encrypter"),
	)

	// Init will parse the command line flags.
	service.Init()

	// Register handler
	proto.RegisterEncrypterHandler(service.Server(), new(handlers.Encrypter))

	// Run the server
	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}

/*
$ go run main.go
2021-02-28 08:50:54  file=v3@v3.5.0/service.go:192 level=info Starting [service] encrypter
2021-02-28 08:50:54  file=server/rpc_server.go:820 level=info service=server Transport [http] Listening on [::]:50497
2021-02-28 08:50:54  file=server/rpc_server.go:840 level=info service=server Broker [http] Connected to 127.0.0.1:50498
2021-02-28 08:50:54  file=server/rpc_server.go:654 level=info service=server Registry [mdns] Registering node: encrypter-31acb2f8-b40d-46fd-9af2-003be2625a77
*/
