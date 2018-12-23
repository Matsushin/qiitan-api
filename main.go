package main

import (
	"github.com/Matsushin/qiitan-api/handler"
	"github.com/Matsushin/qiitan-api/logger"
)

func main() {
	handler := handler.NewHandler()
	err := handler.Run()
	if err != nil {
		logger.WithoutContext().Fatalf("Starting Server FAILED!!: %+v", err)
	}
}
