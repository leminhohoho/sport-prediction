package proxy

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestProxiesValidation(t *testing.T) {
	data, err := os.ReadFile("proxies.txt")
	if err != nil {
		t.Fatal(err)
	}

	proxies := strings.Split(string(data), "\n")

	validProxies, err := GetValidProxies(proxies[:20])
	if err != nil {
		t.Fatal(err)
	}

	for _, validProxy := range validProxies {
		fmt.Println(validProxy)
	}
}
