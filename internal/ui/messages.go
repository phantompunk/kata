package ui

import "fmt"

func ShowError(err error) {
	switch e := err.(type) {
	default:
		fmt.Println("✘", e.Error())
	}
}
