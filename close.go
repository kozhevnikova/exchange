package main

import (
	"fmt"
	"io"
	"os"
)

func Close(c io.Closer) {
	err := c.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR > defer close file >", err)
	}
}
