
package main

import (
	"context"
	"fmt"
	"os"

	"generation_bot/internal/generation"
)

func main() {
	client := generation.NewClient("http://127.0.0.1:8000")

	imageData, err := client.Generate(
		context.Background(),
		"A futuristic city",
	)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("test.png", imageData, 0644)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Image received! Size: %d bytes\n", len(imageData))
}
