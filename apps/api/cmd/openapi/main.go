package main

import (
	"encoding/json"
	"fmt"
	"os"

	contract "github.com/KARSIFT/vocanova-platform/apps/api/app/api"
)

func main() {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(contract.NewContractAPI().OpenAPI()); err != nil {
		fmt.Fprintf(os.Stderr, "generate OpenAPI: %v\n", err)
		os.Exit(1)
	}
}
