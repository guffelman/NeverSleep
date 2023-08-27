package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

// convert /img/icon.ico to byte array and write to a file named icon.go

// write as text

func convertIcon() {
	// read file
	// convert to byte array
	// write to file

	//read
	f, err := os.Open("./img/icon.ico")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// convert to byte array
	data, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}

	// write to file
	// write as text

	a, err := os.Create("./icon.go")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// write as text
	_, err = a.WriteString("package main\n\n")
	if err != nil {
		log.Fatal(err)
	}

	_, err = a.WriteString("var icon = []byte{")
	if err != nil {
		log.Fatal(err)
	}

	for _, b := range data {
		_, err = a.WriteString(fmt.Sprintf("%d,", b))
		if err != nil {
			log.Fatal(err)
		}
	}

	_, err = a.WriteString("}")
	if err != nil {
		log.Fatal(err)
	}

	// save
	err = a.Sync()
	if err != nil {
		log.Fatal(err)
	}

}
