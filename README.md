# Traloc

Локальный перевод Markdown-документов через Ollama

<p align="center">
  <img src="https://img.shields.io/github/go-mod/go-version/artixzenevich/traloc">
  <img src="https://img.shields.io/github/v/release/artixzenevich/traloc">
  <img src="https://img.shields.io/github/license/artixzenevich/traloc">
  <img src="https://img.shields.io/github/actions/workflow/status/artixzenevich/traloc/ci.yml?branch=main">
</p>

---

## Описание

Traloc — CLI-утилита для перевода Markdown-файлов через локальную LLM, запущенную в [Ollama](https://ollama.ai). Данные не покидают ваш компьютер.

### Возможности

- Пакетная обработка нескольких файлов (glob-паттерны)
- Кэширование результатов перевода для повторного использования
- Маскировка блоков кода и инлайн-кода — LLM переводит только текст, код остаётся нетронутым
- Автоматическое разбиение больших файлов на чанки
- Стриминг ответа со спиннером и прогресс-баром
- Автозапуск и остановка Ollama, выгрузка модели из VRAM
- Graceful shutdown по Ctrl+C
- Конфигурация через YAML-файл

---

## Требования

- Go 1.26 или новее
- Установленный [Ollama](https://ollama.ai) (скачивается отдельно)
- Модель для перевода (по умолчанию `translategemma:4b`)

Установка модели:

```bash
ollama pull translategemma:4b
```

---

## Установка

**Собрать из исходников:**

```bash
git clone https://github.com/artixzenevich/traloc.git
cd traloc
make build
```

Бинарник появится в `bin/traloc`.

**Через go install:**

```bash
go install github.com/artixzenevich/traloc/cmd/traloc@latest
```

---

## Использование

```bash
# перевести все .md в текущей директории на русский
traloc -in "*.md"

# перевести конкретный файл на немецкий
traloc -in README.md -to German

# перевести файлы из папки docs в папку translated
traloc -in "docs/*.md" -outdir translated

# перевести на французский с нестандартной моделью
traloc -in article.md -to French -model llama3.2:3b

# не выгружать модель и не останавливать Ollama после работы
traloc -in "*.md" -keep-ollama

# несколько файлов
traloc -in file1.md -to French file2.md file3.md
```

### Флаги

| Флаг | По умолчанию | Описание |
|---|---|---|
| `-in` | — | Паттерн для поиска файлов (`*.md`) или список файлов через пробел |
| `-outdir` | (рядом с исходным) | Директория для сохранения переведённых файлов |
| `-from` | `English` | Исходный язык |
| `-to` | `Russian` | Целевой язык |
| `-model` | `translategemma:4b` | Модель Ollama |
| `-chunk` | `3500` | Максимальное количество токенов на один чанк |
| `-config` | `.traloc.yaml` | Путь к конфигурационному файлу |
| `-cache` | `.translate-cache` | Путь к директории кэша переводов |
| `-keep-ollama` | `false` | Не останавливать Ollama и не выгружать модель после завершения |

### Выходные файлы

Итоговый файл сохраняется рядом с исходным с добавлением кода языка:

```
документ.md  →  документ.ru.md
статья.md    →  статья.de.md      (при -to German)
```

Если указан `-outdir`:

```
-outdir translated  →  translated/документ.ru.md
```

---

## Конфигурация

Traloc ищет `.traloc.yaml` в текущей директории. Можно указать другой путь через флаг `-config`.

```yaml
languages:
  english: en
  russian: ru
  german: de
  french: fr
  spanish: es
  italian: it
  portuguese: pt
  chinese: zh
  japanese: ja
  korean: ko
  ukrainian: uk
  polish: pl
  dutch: nl
  turkish: tr
  arabic: ar

default_model: translategemma:4b
chunk_size: 3500
cache_dir: .translate-cache
```

Значения из конфига переопределяются флагами командной строки. Например, `-model llama3.2:3b` перепишет `default_model` из конфига.

---

## Как это работает

1. **Поиск файлов** — паттерн расширяется через `filepath.Glob`, в результат добавляются позиционные аргументы
2. **Запуск Ollama** — если сервер не отвечает на `:11434`, Traloc запускает `ollama serve` и ждёт до 30 секунд
3. **Разбиение на чанки** — текст делится по параграфам, длинные параграфы дробятся по предложениям. Каждый чанк не превышает `-chunk` токенов (грубая оценка: латиница ÷4, кириллица ÷2)
4. **Маскировка кода** — блоки ` ``` ` и инлайн `` ` `` заменяются на плейсхолдеры `@@KEEP_BLOCK_N@@` и `@@KEEP_INLINE_N@@`
5. **Перевод** — каждый чанк отправляется в Ollama через streaming API. Промпт содержит инструкцию сохранить Markdown-разметку и не трогать плейсхолдеры
6. **Кэширование** — SHA256 хеш от `модель|язык_источник|язык_цель|текст` сохраняется в `.translate-cache/`. Повторный перевод того же текста с теми же параметрами занимает миллисекунды
7. **Восстановление** — плейсхолдеры заменяются обратно на исходный код
8. **Сборка** — все чанки склеиваются через двойной перенос строки и сохраняются

---

## Сборка из исходников

```bash
make build       # собрать бинарник
make test        # запустить тесты
make lint        # запустить линтер (golangci-lint)
make run ARGS="-in *.md"   # собрать и запустить
make clean       # удалить bin/
```

---

## Структура проекта

```
traloc/
├── cmd/traloc/main.go          — точка входа, флаги
├── internal/
│   ├── app/app.go              — основная логика
│   ├── config/config.go        — конфигурация (YAML + флаги)
│   ├── ollama/lifecycle.go     — управление процессом Ollama
│   ├── styles/styles.go        — lipgloss-стили
│   └── translator/
│       ├── cache.go            — кэширование SHA256
│       ├── code_store.go       — маскировка кода
│       ├── tokenizer.go        — токенизация и разбиение
│       └── translate.go        — вызов Ollama API
├── .github/workflows/ci.yml    — CI (линтер, тесты, сборка)
├── .golangci.yml               — конфиг линтера
├── .traloc.yaml                — дефолтный конфиг
├── Makefile
├── go.mod
└── go.sum
```

---

## Лицензия

MIT
