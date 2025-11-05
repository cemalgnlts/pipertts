package cmd

import (
	"fmt"
	"os"
	pipertts "pipertts/pkg"
)

func Start() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage <model.onnx> 'hello world'\n")
		os.Exit(1)
	}

	var model string = os.Args[1]
	var text string = os.Args[2]

	pipertts.Generate(model, text)
}
