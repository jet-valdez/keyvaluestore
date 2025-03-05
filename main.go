package main

import (
	"fmt"
	"keyvaluestore/storage"
)

func main() {

	db, err := storage.NewInMemoryDB()

	fmt.Println("Hello, World!")
}
