// Copyright 2020 New Relic Corporation. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"fmt"
	"os"
	"strings"
)

// This fixture integration submits as inventory all the key=value pairs passed as arguments
// It uses a remote entity.

var inventoryTestSample = `{"name":"testing_integration","protocol_version":"2","integration_version":"1.2.3","integration_status":"","data":[{"entity":{"name":"localtest","type":"test","id_attributes":[{"Key":"idkey","Value":"idval"}],"displayName":"","metadata":null},"metrics":[],"inventory":{"cliargs":%s},"events":[],"add_hostname":false,"cluster":"local-test","service":"test-service"}]}`

func main() {
	cliargs := "{"
	for i, pair := range os.Args[1:] {
		kv := strings.Split(pair, "=")
		if len(kv) < 2 {
			_, _ = fmt.Fprint(os.Stderr, "argument must be in form key=value. Got:", pair)
			os.Exit(-1)
		}
		cliargs += fmt.Sprintf("\"%s\":\"%s\"", kv[0], kv[1])
		if i != len(os.Args[1:])-1 {
			cliargs += ","
		}
	}
	cliargs += "}"

	fmt.Println(fmt.Sprintf(inventoryTestSample, cliargs))
}
