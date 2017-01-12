/*
Copyright (c) 2017 Kristopher K. Watts

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package main

import (
	"bufio"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os"
)

var (
	fVerbose = flag.Bool("v", false, "Verbose output")
	fOut     = flag.String("o", "/tmp/output.bin", "Output file")
	verbose  bool
	out      string
	files    []string
	nw       nilWriter
)

type nilWriter bool

func init() {
	flag.Parse()
	verbose = *fVerbose
	out = *fOut
	files = flag.Args()
	if len(out) == 0 {
		log.Fatal("output file required")
	}
	if len(files) <= 2 {
		log.Fatal("I need at least 3 input files for this to work")
	}
	if !verbose {
		log.SetOutput(nw)
	}
}

func main() {
	var fins [][]byte
	fout, err := os.Create(out)
	if err != nil {
		log.Fatal("Failed to open", out, err)
	}
	defer fout.Close()
	var lastLen int
	for i := range files {
		bts, err := ioutil.ReadFile(files[i])
		if err != nil {
			log.Println("Failed to read file", files[i], err)
			return
		}
		if lastLen == 0 {
			lastLen = len(bts)
		}
		if lastLen != len(bts) {
			log.Println("File", files[i], "Does not match the last length of", lastLen)
			return
		}
		fins = append(fins, bts)
	}
	log.Println("Processing", len(fins), "files")
	n, err := vote(fout, fins)
	if err != nil {
		log.Println("Failed to vote", err)
		return
	}
	log.Println(n, "bits corrected out of", 8*lastLen)
}

func vote(fout *os.File, fins [][]byte) (int, error) {
	var corrected int
	if len(fins) == 0 {
		return -1, errors.New("Invalid vote set")
	}
	out := bufio.NewWriter(fout)
	outsize := len(fins[0])
	byteSet := make([]byte, len(fins))
	for i := 0; i < outsize; i++ {
		for j := range fins {
			byteSet[j] = fins[j][i]
		}
		b, n := voteByte(byteSet)
		corrected += n
		if err := out.WriteByte(b); err != nil {
			return -1, err
		}
	}
	return corrected, out.Flush()
}

func voteByte(bt []byte) (byte, int) {
	//do byte check to see if they are all identical, this SHOULD be the most common case
	if fastCheck(bt) {
		return bt[0], 0
	}
	var out byte
	var corrected int
	//do a bitwise vote
	for i := uint(0); i < 8; i++ {
		var one int
		for _, b := range bt {
			if (b & (1 << i)) != 0 {
				one++
			}
		}
		//majority wins
		if one > (len(bt) / 2) {
			out |= (1 << i)
		}
		//if not unanimous, it was corrected
		if one != len(bt) && one != 0 {
			corrected++
		}
	}
	return out, corrected
}

func fastCheck(b []byte) bool {
	x := b[0]
	for i := range b {
		if b[i] != x {
			return false
		}
	}
	return true
}

func (nw nilWriter) Write(p []byte) (int, error) {
	return len(p), nil
}
