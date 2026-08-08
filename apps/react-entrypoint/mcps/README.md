# Invoice Parser

Go CLI + MCP инструмент для парсинга email-счетов (.eml) и PDF-квитанций,
извлечения структурированных данных через RAG pipeline (Ollama), генерации
SQL INSERT с валидацией и создания HTML/PDF отчётов.

---

## Возможности

- Парсинг `.eml` (MIME, multipart, text/plain, text/html, PDF вложения)
- Парсинг `.pdf` (штрафы, квитанции, счета)
- RAG извлечение данных: поставщик, сумма, дата, описание, категория
- LLM-категоризация расходов
- Генерация SQL INSERT
- Валидация SQL: синтаксис (sqlparser) + EXPLAIN + INSERT...RETURNING в реальную БД
- HTML + PDF отчёт
- MCP сервер для интеграции с opencode

---

## Команды

```bash
# Собрать
go build -o invoice-parser ./cmd/invoice-parser

# Обработать один файл (.eml или .pdf)
./invoice-parser -file invoice.eml

# Обработать все файлы в папке
./invoice-parser -dir ./invoices/

# Обработать + проверить SQL через реальную БД
./invoice-parser -file invoice.pdf -db-dsn "postgres://postgres:postgres@localhost:5432/mydb"

# Режим MCP сервера (для opencode)
./invoice-parser --mcp

# Указать путь для HTML отчёта
./invoice-parser -file invoice.eml -output report.html

# Справка
./invoice-parser -h
```

## Переменные окружения

| Переменная | По умолчанию | Описание |
|---|---|---|
| `OLLAMA_BASE_URL` | `http://192.168.1.85:11434` | Адрес Ollama сервера |
| `OLLAMA_MODEL` | `mistral` | Модель для генерации |
| `OLLAMA_EMBEDDING_MODEL` | `nomic-embed-text` | Модель для эмбеддингов |

---

## Структура проекта

```
invoice-parser/
├── cmd/invoice-parser/
│   └── main.go                  — точка входа, CLI флаги, пайплайн
├── internal/
│   ├── email/
│   │   ├── parser.go            — парсинг .eml (MIME, multipart, quoted-printable)
│   │   └── reader.go            — чтение invoice-файлов (.eml / .pdf)
│   ├── llm/
│   │   └── ollama.go            — HTTP клиент для Ollama (/api/embeddings + /api/generate)
│   ├── rag/
│   │   ├── splitter.go          — рекурсивный сплиттер текста (1000/200)
│   │   ├── vectorstore.go       — in-memory векторное хранилище + cosine similarity
│   │   └── pipeline.go          — RAG пайплайн: сплит → эмбеддинг → поиск
│   ├── extract/
│   │   └── extract.go           — RAG → структурированный JSON (InvoiceRecord)
│   ├── sql/
│   │   ├── generator.go         — InvoiceRecord → SQL INSERT
│   │   └── validator.go         — валидация: sqlparser + DB EXPLAIN + INSERT...RETURNING
│   ├── category/
│   │   └── classifier.go        — LLM-классификация категорий расходов
│   ├── mcp/
│   │   └── server.go            — MCP stdio сервер для opencode
│   └── report/
│       └── report.go            — генерация HTML + PDF отчёта
├── templates/
│   └── report.html              — шаблон HTML отчёта
├── opencode.json                 — MCP конфигурация для opencode
├── AGENTS.md                     — инструкция для opencode-агента
├── README.md                     — этот файл
├── go.mod / go.sum
└── invoice-parser                — скомпилированный бинарник
```

## MCP инструменты (opencode)

```json
{
  "mcp": {
    "invoice-parser": {
      "type": "local",
      "enabled": true,
      "command": "./invoice-parser",
      "args": ["--mcp"]
    }
  }
}
```

| Инструмент | Параметры | Описание |
|---|---|---|
| `parse_invoice` | `file: string` | Парсит .eml/.pdf → JSON + SQL |
| `categorize` | `vendor: string`, `description?: string` | Предлагает категорию для поставщика |

---

## Что делать дальше

1. **Взять настоящий счёт** — экспортировать письмо из почты в `.eml` (Gmail: ⋮ → Скачать сообщение / Outlook: Файл → Сохранить как → EML) или скачать PDF квитанцию
2. **Положить в корень проекта** — `invoice.eml` или `invoice.pdf`
3. **Запустить** — `./invoice-parser -file invoice.eml`
4. **Проверить результат** — `output.html` (откроется в браузере)
5. **Попробовать opencode** — запустить `--mcp` и через opencode вызвать `parse_invoice`
6. **Добавить свои категории** — в `cmd/invoice-parser/main.go`, маппинг `CategoryMapping`
7. **Подключить БД** — `--db-dsn "postgres://..."` для полной валидации SQL

### Запуск с opencode

```bash
# Терминал 1: запустить MCP сервер
./invoice-parser --mcp

# opencode.json уже настроен, opencode сам запустит через stdio
```

---

## Архитектура (RAG pipeline)

```
.eml / .pdf
  └→ email/reader.go (извлечение текста)
       └→ rag/splitter.go (разбивка на чанки 1000 символов)
            └→ llm/ollama.go (эмбеддинг каждого чанка)
                 └→ rag/vectorstore.go (сохранение в in-memory store)
                      └→ llm/ollama.go (эмбеддинг запроса)
                           └→ rag/vectorstore.go (cosine similarity, top-5)
                                └→ llm/ollama.go (генерация JSON через LLM)
                                     └→ extract/extract.go (парсинг JSON → InvoiceRecord)
                                          └→ sql/generator.go (INSERT)
                                               └→ sql/validator.go (проверка)
                                                    └→ report/report.go (HTML + PDF)
```
