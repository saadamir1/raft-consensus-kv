package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
)

func main() {
	// Define command line flags
	serverAddr := flag.String("server", "localhost:8080", "Server address (host:port)")
	operation := flag.String("op", "get", "Operation to perform: put, append, get")
	key := flag.String("key", "", "Key to operate on")
	value := flag.String("value", "", "Value for put/append operations")
	flag.Parse()

	// Verify required flags
	if *key == "" {
		fmt.Println("Error: Key is required")
		flag.Usage()
		return
	}

	// Construct the appropriate URL
	var url string
	switch *operation {
	case "put":
		if *value == "" {
			fmt.Println("Error: Value is required for put operation")
			flag.Usage()
			return
		}
		url = fmt.Sprintf("http://%s/put", *serverAddr)
		data := map[string]string{
			"key":   *key,
			"value": *value,
		}
		sendRequest("POST", url, data)

	case "append":
		if *value == "" {
			fmt.Println("Error: Value is required for append operation")
			flag.Usage()
			return
		}
		url = fmt.Sprintf("http://%s/append", *serverAddr)
		data := map[string]string{
			"key":   *key,
			"value": *value,
		}
		sendRequest("POST", url, data)

	case "get":
		url = fmt.Sprintf("http://%s/get?key=%s", *serverAddr, *key)
		sendRequest("GET", url, nil)

	default:
		fmt.Printf("Error: Unknown operation '%s'. Use put, append, or get.\n", *operation)
	}
}

// sendRequest sends HTTP requests to the server
func sendRequest(method, url string, data map[string]string) {
	var req *http.Request
	var err error

	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Printf("Error encoding data: %v\n", err)
			return
		}
		req, err = http.NewRequest(method, url, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			return
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	fmt.Printf("Status: %s\n", resp.Status)
	if len(body) > 0 {
		fmt.Printf("Response: %s\n", string(body))
	}
}