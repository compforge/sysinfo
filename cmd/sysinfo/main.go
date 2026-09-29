// Copyright © 2016 Zlatko Čalušić
//
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file.

// sysinfo is a very simple utility demonstrating sysinfo library capabilities. Start it to get pretty formatted JSON
// output of all the info that sysinfo library provides. Due to its simplicity, the source code of the utility also
// doubles down as an example of how to use the library.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/compforge/sysinfo"
)

func main() {
	if len(os.Args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: sysinfo (prints system information as JSON)")
		os.Exit(2)
	}

	var si sysinfo.SysInfo

	si.GetSysInfo()

	data, err := json.MarshalIndent(&si, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	if _, err := fmt.Println(string(data)); err != nil {
		log.Fatal(err)
	}
}
