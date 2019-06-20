package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	outFile := flag.String("o", "y.tab.go", "parser output")
	statFile := flag.String("v", "y.output", "create parsing tables")
	prefix := flag.String("p", "yy", "name prefix")
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Println("Need exactly one argument")
		os.Exit(1)
	}

	var g LRGen
	var err error
	g.Path = flag.Arg(0)
	inFile, err := os.Open(g.Path)
	if err != nil {
		log.Fatal(err)
	}
	g.In = bufio.NewReader(inFile)
	g.Out, err = os.Create(*outFile)
	if err != nil {
		log.Fatal(err)
	}
	g.Stat, err = os.Create(*statFile)
	if err != nil {
		log.Fatal(err)
	}
	g.Prefix = *prefix
	g.Run()
}
