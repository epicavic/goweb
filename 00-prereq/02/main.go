package main

import "fmt"

var list []string

func init() {
	list = initList()
}

func main() {
	fmt.Println(list)
	list = addList(list)
	fmt.Println(list)
}

func initList() []string {
	l := []string{"1", "2", "3"}
	return l
}

func addList(l []string) []string {
	l = append(l, "4", "5", "6")
	return l
}
