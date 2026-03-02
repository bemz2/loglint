# loglint

`loglint` — линтер для Go, который проверяет лог-сообщения в коде и ищет нарушения правил оформления. Проект реализован как анализатор на базе `golang.org/x/tools/go/analysis`, может запускаться как standalone-инструмент и как Go plugin для `golangci-lint`.

## Что проверяет линтер

`loglint` проверяет следующие правила:

1. Лог-сообщение должно начинаться со строчной буквы.
2. Лог-сообщение должно быть только на английском языке.
3. Лог-сообщение не должно содержать спецсимволы и эмодзи.
4. Лог-сообщение не должно содержать потенциально чувствительные данные.

Поддерживаемые логгеры:
- `log/slog`
- `go.uber.org/zap`

## Что реализовано

- анализатор на `go/analysis`;
- поддержка `slog` и `zap`;
- standalone-запуск через `singlechecker`;
- Go plugin для `golangci-lint` (`.so`);
- конфигурация через JSON;
- `SuggestedFix` для правила о первой строчной букве;
- unit-тесты и `analysistest`;
- базовый GitLab CI.

## Требования

- Go 1.22+

## Структура проекта

```text
cmd/loglint/         standalone-запуск линтера
pkg/loglint/         логика анализатора
plugin/              entrypoint для golangci-lint Go plugin
```

## Сборка

### Standalone-версия

```bash
go build -o bin/loglint ./cmd/loglint
```

После сборки будет доступен исполняемый файл `bin/loglint`.

### Плагин для golangci-lint

```bash
mkdir -p build
go build -buildmode=plugin -o build/loglint.so ./plugin
```

После сборки будет доступен Go plugin `build/loglint.so`.

## Запуск

### Локальный запуск

Без предварительной сборки:

```bash
go run ./cmd/loglint ./...
```

После сборки:

```bash
./bin/loglint ./...
```

### Запуск с конфигурационным файлом

```bash
./bin/loglint -config ./loglint.json ./...
```

## Конфигурация

Линтер поддерживает JSON-конфиг. Пример файла `loglint.json`:

```json
{
  "check_lowercase_start": true,
  "check_english_only": true,
  "check_special_chars": true,
  "check_sensitive_data": true,
  "sensitive_keywords": [
    "password",
    "passwd",
    "pwd",
    "apikey",
    "secret",
    "token"
  ]
}
```

Описание полей:
- `check_lowercase_start` — включает проверку начала сообщения со строчной буквы;
- `check_english_only` — включает проверку на только английский текст;
- `check_special_chars` — включает проверку на спецсимволы и эмодзи;
- `check_sensitive_data` — включает проверку на чувствительные данные;
- `sensitive_keywords` — список ключевых слов для поиска чувствительных данных.

Если параметр не указан, используется значение по умолчанию.

## Подключение к golangci-lint

`loglint` поддерживает подключение как Go plugin через `buildmode=plugin`.

Сначала соберите плагин:

```bash
mkdir -p build
go build -buildmode=plugin -o build/loglint.so ./plugin
```

Пример `.golangci.yml`:

```yaml
linters-settings:
  custom:
    loglint:
      path: ./build/loglint.so
      description: Checks slog and zap log messages
      original-url: https://github.com/<your-user>/<your-repo>
      settings:
        check_lowercase_start: true
        check_english_only: true
        check_special_chars: true
        check_sensitive_data: true
        sensitive_keywords:
          - password
          - token
          - secret

linters:
  enable:
    - loglint
```

Если нужно хранить настройки в отдельном JSON-файле:

```yaml
linters-settings:
  custom:
    loglint:
      path: ./build/loglint.so
      description: Checks slog and zap log messages
      original-url: https://github.com/<your-user>/<your-repo>
      settings:
        config: ./loglint.json
```

## Примеры

Некорректно:

```go
slog.Info("Starting server")
slog.Info("запуск сервера")
slog.Info("server started!")
slog.Info("user password: " + password)
```

Корректно:

```go
slog.Info("starting server")
slog.Info("server started")
slog.Info("user authenticated successfully")
```

Пример запуска:

```bash
./bin/loglint ./...
```

## Автоисправления

Сейчас `SuggestedFix` реализован только для правила 1: если строковый литерал начинается с заглавной латинской буквы, линтер может предложить заменить первую букву на строчную.

Для правил про английский язык и чувствительные данные автоисправления не применяются, так как там нет безопасной механической правки без риска исказить смысл сообщения.

## Ограничения текущей реализации

- Проверка "только английский" основана на анализе Unicode-символов, без семантического понимания текста.
- Проверка чувствительных данных построена на ключевых словах и эвристиках, а не на полном анализе потока данных.
- В качестве интеграции с `golangci-lint` реализован Go plugin (`.so`), а не module plugin system.

## Тесты

Запуск тестов:

```bash
go test ./...
```

Проверка сборки:

```bash
go build ./...
```
