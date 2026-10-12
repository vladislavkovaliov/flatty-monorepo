# DEVELOPMENT — Development Contract

Главный development contract проекта. Агент обязан прочитать этот файл перед началом работы.

---

## 1. Глобальная цель проекта

Разработать Telegram-бота для управления торрент-загрузками через magnet-ссылки.

Итоговая система должна позволять пользователю:

1. Отправить magnet-ссылку в Telegram.
2. Распарсить magnet и получить: название, info hash, trackers, размер (если есть).
3. Выбрать тип контента: 🎬 Фильм / 📺 Сериал.
4. Выбрать качество: 2160p / 4K, 1080p, 720p, «Не определять».
5. Подтвердить загрузку.
6. Добавить torrent в очередь.
7. Смотреть состояние загрузок: статус, прогресс, скорость, ETA, размер.
8. Управлять загрузками: pause, resume, cancel, delete.
9. После завершения организовывать файлы.
10. Хранить фильмы в `Movies/`.
11. Хранить сериалы в `TV/`.
12. В дальнейшем использовать эту библиотеку как источник для Lampa.

---

## 2. Текущий этап

На первом этапе реализуется **только Telegram UI и обработка magnet-ссылок**.

### TEST_MODE=true (критически важно на MVP)

В этом режиме запрещено:

- подключение к qBittorrent;
- реальные torrent downloads;
- запись torrent-файлов на NAS;
- доступ к NAS;
- перемещение медиа;
- реальные filesystem downloads.

Можно только **симулировать** загрузку.

### Current Scope (что реализуем сейчас)

```text
/start
  → Receive magnet
  → Parse magnet
  → Display metadata
  → Choose Movie / Series
  → Choose Quality
  → Confirmation
  → FakeTorrentProvider
  → Test Queue
  → /downloads
```

Не реализовывать сейчас (без явного перевода проекта на соответствующую фазу):

- qBittorrent
- NAS, SMB, NFS
- real downloads
- media organizer
- Lampa

---

## 3. Правила агента

Перед началом работы агент обязан прочитать:

- `docs/DEVELOPMENT.md` (этот файл)
- перед реализацией конкретной feature: `docs/features/XXX-name.md`
- если меняется архитектура: `docs/ARCHITECTURE.md`
- если принимается существенное архитектурное решение: `docs/decisions/`

Агент обязан:

1. Не писать Telegram business logic в handlers.
2. Не вызывать qBittorrent напрямую из handlers.
3. Использовать service layer.
4. Использовать provider abstraction.
5. Добавлять тесты.
6. Обновлять документацию.
7. Обновлять feature file.
8. Не удалять существующие тесты без причины.
9. Не добавлять зависимости без необходимости.
10. Не коммитить secrets.
11. Не включать реальные downloads на MVP.
12. Проверять lint.
13. Проверять type checking.
14. Запускать tests.
15. Создавать ADR при существенном архитектурном изменении.

---

## 4. Python Tooling

- Python 3.12+
- aiogram 3.x
- pytest, pytest-asyncio, pytest-cov
- Ruff, mypy
- Конфигурация инструментов — предпочтительно в `pyproject.toml`

Команды:

```text
ruff check .
ruff format .
ruff format --check .
mypy app
pytest
pytest -v
pytest --cov=app --cov-report=term-missing
```

Перед завершением feature все команды должны завершаться успешно:

```text
ruff check .
ruff format --check .
mypy app
pytest
```

---

## 5. Dependencies

Runtime:

- aiogram
- python-dotenv

Development:

- pytest
- pytest-asyncio
- pytest-cov
- ruff
- mypy

Не добавлять dependency без необходимости. В будущем могут появиться `redis`, `sqlalchemy`, `alembic`, `httpx` — но только при реализации соответствующей feature.

---

## 6. Logging

- Использовать Python logging, не `print(...)` для application logging.
- Можно логировать: event, user_id, torrent_id, info_hash, status, error type.
- Нельзя логировать: `BOT_TOKEN`, passwords, NAS credentials, private keys.

---

## 7. Git

Использовать Conventional Commits:

```text
feat: add magnet parser
feat: add torrent queue UI
fix: handle invalid magnet links
test: add parser tests
refactor: extract torrent provider
docs: add feature specification
chore: configure ruff
```

Никогда не commit:

- `.env`
- tokens, passwords, private keys, credentials

---

## 8. Feature Workflow

Каждая существенная feature проходит:

```text
Feature idea
  → Создать docs/features/XXX-name.md
  → Описать Acceptance Criteria
  → Определить dependencies
  → Implement
  → Write tests
  → Update documentation
  → Lint
  → Type check
  → Tests
  → Feature = Implemented
```

Статусы feature (использовать только эти):

- `Planned`
- `In Progress`
- `Implemented`
- `Blocked`
- `Deprecated`

Правила features:

- Каждая feature — отдельный Markdown-файл в `docs/features/`.
- Нельзя создавать один огромный `docs/features.md`.
- ID feature после создания менять нельзя.

---

## 9. Definition of Done

Feature считается завершённой только если:

- [ ] Code implemented
- [ ] Unit tests added
- [ ] Integration tests added where necessary
- [ ] Error handling implemented
- [ ] Logging implemented
- [ ] Security considered
- [ ] Documentation updated
- [ ] Feature documentation updated
- [ ] Ruff check passes
- [ ] Ruff format check passes
- [ ] Mypy passes
- [ ] Existing tests pass

---

## 10. Security (общие правила)

Никогда нельзя доверять данным от пользователя: название torrent, имя файла, callback data, пути, текст, metadata.

- При работе с filesystem защищаться от `../../` и абсолютных путей.
- Никогда не логировать: `BOT_TOKEN`, `QBITTORRENT_PASSWORD`, `NAS_PASSWORD`, `REDIS_PASSWORD`, `DATABASE_PASSWORD`.

---

## 11. Структура документации

См. `docs/ROADMAP.md` для этапов, `docs/ARCHITECTURE.md` для архитектуры.

- `docs/features/` — каждая feature отдельным файлом по шаблону.
- `docs/decisions/` — Architecture Decision Records (ADR).
