package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

var (
	mappingsFile   = flag.String("mappings", "golden.txt", "the mappings file")
	outputFile     = flag.String("output", "golden.log", "the deobfuscated mappings file")
	obfuscatedFile = flag.String("input", "obfuscated.log", "the obfuscated log file")
)

func main() {

	flag.Parse()

	f, err := os.Open(*mappingsFile)

	if err != nil {
		panic(err)
	}

	start := time.Now()

	reader := NewGoldenMappingsReader()

	reader.Read(f)

	f.Close()

	f, err = os.Open(*obfuscatedFile)

	if err != nil {
		panic(err)
	}

	defer f.Close()

	var buf bytes.Buffer

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		text := scanner.Text()

		buf.WriteString(text)
		buf.WriteRune('\n')
	}

	str := buf.String()

	for _, class := range reader.GoldenClassMap {
		for _, member := range class.AllMembers {
			obfName := class.ObfName + "." + member.ObfName()

			if strings.Contains(str, obfName) {
				str = strings.Replace(str, obfName, class.GoldenName+"."+member.GoldenName(), -1)
			}
		}
	}

	ioutil.WriteFile(*outputFile, []byte(str), os.ModePerm)

	end := time.Now()

	fmt.Println("completed in", end.Sub(start))
}
