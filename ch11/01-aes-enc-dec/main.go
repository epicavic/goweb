package main

import (
	"aes-enc-dec/utils"
	"fmt"
)

// AES keys should be of length 16, 24, 32
func main() {
	key := "111023043350789514532147"
	message := `
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
	`
	fmt.Printf("Original message: %s\n", message)
	encryptedString := utils.EncryptString(key, message)
	fmt.Printf("Encrypted message: %s\n", encryptedString)
	decryptedString := utils.DecryptString(key, encryptedString)
	fmt.Printf("Decrypted message: %s\n", decryptedString)
}
