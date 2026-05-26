package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof" // стандартный pprof эндпоинт (опционально)
	"os"
	"runtime"
	"time"

	"github.com/grafana/pyroscope-go"
)

func main() {
	// Включаем mutex и block профилирование (опционально)
	runtime.SetMutexProfileFraction(5)
	runtime.SetBlockProfileRate(5)

	// Инициализируем Pyroscope агент
	profiler, err := pyroscope.Start(pyroscope.Config{
		ApplicationName: "my-go-service",

		// Адрес локального Pyroscope сервера из docker-compose
		ServerAddress: getEnv("PYROSCOPE_SERVER_ADDRESS", "http://localhost:4040"),

		Logger: pyroscope.StandardLogger,

		// Статические теги: среда, инстанс и т.п.
		Tags: map[string]string{
			"env":      getEnv("ENV", "local"),
			"hostname": os.Getenv("HOSTNAME"),
		},

		ProfileTypes: []pyroscope.ProfileType{
			// Базовые (включены по умолчанию)
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
			// Опциональные
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	})
	if err != nil {
		log.Fatalf("failed to start Pyroscope: %v", err)
	}
	defer profiler.Stop()

	// Пример: разметка кода тегами для дробления профиля по handler-ам
	http.HandleFunc("/fast", func(w http.ResponseWriter, r *http.Request) {
		pyroscope.TagWrapper(r.Context(), pyroscope.Labels("handler", "fast"), func(_ context.Context) {
			fastWork()
		})
	})

	http.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		pyroscope.TagWrapper(r.Context(), pyroscope.Labels("handler", "slow"), func(_ context.Context) {
			slowWork()
		})
	})

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func fastWork() {
	time.Sleep(1 * time.Millisecond)
}

func slowWork() {
	// Имитируем тяжёлую работу
	sum := 0
	for i := 0; i < 10_000_000; i++ {
		sum += i
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
