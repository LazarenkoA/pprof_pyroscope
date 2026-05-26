# Pyroscope Local Setup

Локальный стек: Pyroscope + Grafana в одном docker-compose.


## Запуск

```bash
docker compose up -d
```

## Адреса

- Pyroscope UI: http://localhost:4040
- Grafana:       http://localhost:3000
  - Explore Profiles: http://localhost:3000/a/grafana-pyroscope-app/profiles-explorer

## Подключение Go-приложения

#### 1. Установить SDK

```bash
go get github.com/grafana/pyroscope-go
```

#### 2. Инициализировать агент

```go
import "github.com/grafana/pyroscope-go"

profiler, err := pyroscope.Start(pyroscope.Config{
    ApplicationName: "my-go-service",
    ServerAddress:   "http://localhost:4040",  // или http://pyroscope:4040 внутри compose
    Logger:          pyroscope.StandardLogger,
    Tags:            map[string]string{"env": "local"},
    ProfileTypes: []pyroscope.ProfileType{
        pyroscope.ProfileCPU,
        pyroscope.ProfileAllocObjects,
        pyroscope.ProfileAllocSpace,
        pyroscope.ProfileInuseObjects,
        pyroscope.ProfileInuseSpace,
    },
})
if err != nil {
    log.Fatal(err)
}
defer profiler.Stop()
```

#### 3. Разметка кода тегами (опционально)

```go
pyroscope.TagWrapper(ctx, pyroscope.Labels("handler", "my_handler"), func(ctx context.Context) {
    // профилируемый код
})
```

#### 4. Если приложение тоже в compose

Добавьте в docker-compose.yml:

```yaml
  your-app:
    build: .
    environment:
      - PYROSCOPE_SERVER_ADDRESS=http://pyroscope:4040
    networks:
      - profiling
    depends_on:
      - pyroscope
```

### Полезные команды

```bash
# Проверить готовность Pyroscope
curl http://localhost:4040/ready

# Остановить стек
docker compose down

# Удалить данные (сбросить хранилище)
docker compose down -v
```

## Подключить ванильный pprof

Единственное, что нужно поменять
В файле `alloy/config.alloy` замени порт на тот, где у тебя висит приложение:
```
targets = [
  {
      "__address__"  = "host.docker.internal:8080",  // <-- твой порт
      "service_name" = "my-go-service",
  }
]
```
Приложение на хосте → host.docker.internal:<PORT> (уже прописан extra_hosts в compose)

Приложение тоже в compose → просто имя контейнера: my-app-container:8080

Запуск `docker compose up -d`

### Проверить, что Alloy видит таргеты и не ругается
`docker logs alloy -f`

Адреса:
- Alloy UI (статус скрейпа, дебаг) → http://localhost:12345
- Pyroscope UI → http://localhost:4040
- Grafana Explore Profiles → http://localhost:3000/a/grafana-pyroscope-app/profiles-explorer


