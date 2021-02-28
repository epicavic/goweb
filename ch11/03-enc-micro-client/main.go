package main

import (
	"context"
	"fmt"

	proto "enc-micro-client/proto"

	micro "github.com/asim/go-micro/v3"
)

func main() {
	// Create a new service
	service := micro.NewService(micro.Name("encrypter.client"))
	// Initialise the client and parse command line flags
	service.Init()

	// Create new encrypter client
	encrypter := proto.NewEncrypterService("encrypter", service.Client())

	// Call the encrypter
	rsp, err := encrypter.Encrypt(context.TODO(), &proto.Request{
		Message: `
		I have existed from the morning of the world
		And I shall exist until the last star falls from the night
		Although I have taken the form of Gaius Caligula
		I am all man as I am no man and therefore I am
		A god
		I shall wait for the unanimous decision of the senate, Claudius
		All those who say 'aye', say 'aye'
		Aye
		Aye
		Aye, aye, aye, aye...
		He's a god now
		`,
		Key: "111023043350789514532147",
	})
	if err != nil {
		fmt.Println(err)
	}
	// Print response
	fmt.Println(rsp.Result)

	rsp, err = encrypter.Decrypt(context.TODO(), &proto.Request{
		Message: rsp.Result,
		Key:     "111023043350789514532147",
	})
	if err != nil {
		fmt.Println(err)
	}
	// Print response
	fmt.Println(rsp.Result)
}

/*
$ go run main.go
sNbhLfTyux3I3CI0ljxnDYhuchL0BsfUVDYrZnbMOjUwM7WpIZnGb0zhXah8nizDNU3Iz5NwsYg+ay8i+kakGA9XrYYJ3/0OVc4BmNpyvxHjK0gBG87kScczzj7HT2orbAXLhF5TVLcDIDn6axR+IyMNAvzzHtA+GyIlfuG6hZGRnjr2npg6fnDeKoYeenz5MlU2n9NyjB/Z0BnwySZr8y/xxh2P7ivz6wAMNM0rvqUiXk1OtD2W8ZhY5ByvpR1lqLj0gMDR/5GmF+cQyyRLaTc3lR57pDjCCIWA7osFKvC0XyoGV39AMjf/8XenjCnNK7m1J/6Wa0I/yKFLDhHudsHXQw/+K/mcchsIu9SFBjkKDOy2ifwARhr3LpRNFqpRInR5YX8SdYNFzNcAMutWjEmOFtF43lhF12qiAJxPRr2opR+bCoWuBrUBj8Ji2TrqFApnLUsAcL6wYBpcR8rVv7hIt+LKHobgy9NzDyChE4aWZjQLRcs3

		I have existed from the morning of the world
		And I shall exist until the last star falls from the night
		Although I have taken the form of Gaius Caligula
		I am all man as I am no man and therefore I am
		A god
		I shall wait for the unanimous decision of the senate, Claudius
		All those who say 'aye', say 'aye'
		Aye
		Aye
		Aye, aye, aye, aye...
		He's a god now
*/
