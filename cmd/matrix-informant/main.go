package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/edgetriggered/matrix-informant/pkg/informant"
)

var (
	buildHash = "unknown"
	buildDate = "unknown"
)

func main() {
	var conf = flag.String("c", "./conf/informant.yaml", "path to configuration")
	var ver = flag.Bool("v", false, "display version information")
	flag.Parse()

	if *ver {
		fmt.Printf("%v built on %v\n", buildHash, buildDate)
		os.Exit(0)
	}

	informant.Inform(*conf)
}
