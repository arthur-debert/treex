package parser

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

// Data represents the structure of our JSON data
type Data struct {
	Name    string            `json:"name"`
	Version string            `json:"version"`
	Config  map[string]string `json:"config"`
}

// ParseFile reads and parses a JSON file
func ParseFile(filename string) {
	fmt.Printf("Parsing file: %s\n", filename)
	
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}
	
	var parsed Data
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}
	
	fmt.Printf("Parsed data: %+v\n", parsed)
}

// ParseString parses JSON from a string
func ParseString(jsonStr string) (*Data, error) {
	var data Data
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
} 