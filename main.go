package main

import (
	"mutualTLS-proxy/program"
	"os"
)

func main() {
	program.Main(os.Args[1:])
}
