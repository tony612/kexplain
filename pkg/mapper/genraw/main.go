package main

// A tool to generate resource data for remote.
// kubectl api-resources | go run main.go

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

func main() {
	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	data := string(bytes)
	lines := strings.Split(data, "\n")
	result := make([][4]string, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := regexp.MustCompile(`\s+`).Split(line, -1)
		if parts[0] == "NAME" {
			continue
		}
		if len(parts) != 4 && len(parts) != 5 {
			panic(fmt.Sprintf("\"%s\" must have 4 or 5 parts\n", line))
		}
		item := [4]string{parts[0], "", parts[len(parts)-3], parts[len(parts)-1]}
		if len(parts) == 5 {
			item[1] = parts[1]
		}
		result = append(result, item)
	}

	for _, item := range result {
		fmt.Printf(`  {"%s", "%s", "%s", "%s"},`+"\n", item[0], item[1], item[2], item[3])
	}
}
