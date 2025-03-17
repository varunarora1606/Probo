package main

import "fmt"

func main() {
	var name = "Varun"
	var arr []string
	fmt.Printf("Hello World %s!!\n", name)
	arr = append( arr, name)
	for _, booking := range arr {
		fmt.Println((booking))
	}
}