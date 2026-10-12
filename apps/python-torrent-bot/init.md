# Torrent Telegram Bot — Development Specification

## 1. Глобальная цель проекта

Разработать Telegram-бота для управления торрент-загрузками через magnet-ссылки.

Итоговая система должна позволять пользователю:

1. Отправить magnet-ссылку в Telegram.
2. Распарсить magnet и получить:
   - название;
   - info hash;
   - trackers;
   - размер, если он присутствует.
3. Выбрать тип контента:
   - 🎬 Фильм;
   - 📺 Сериал.
4. Выбрать качество:
   - 2160p / 4K;
   - 1080p;
   - 720p;
   - Не определять.
5. Подтвердить загрузку.
6. Добавить torrent в очередь.
7. Смотреть состояние загрузок:
   - статус;
   - прогресс;
   - скорость;
   - ETA;
   - размер.
8. Управлять загрузками:
   - pause;
   - resume;
   - cancel;
   - delete.
9. После завершения организовывать файлы.
10. Хранить фильмы в `Movies/`.
11. Хранить сериалы в `TV/`.
12. В дальнейшем использовать эту библиотеку как источник для Lampa.

---

# 2. Главный архитектурный принцип

Проект нельзя реализовывать как один большой Telegram-скрипт.

Целевая архитектура:

```text
                    Telegram
                       |
                       v
              Telegram Bot / UI
                       |
                       v
              Application Services
                       |
          +------------+------------+
          |                         |
          v                         v
       Database              Torrent Provider
                                    |
                       +------------+------------+
                       |                         |
                       v                         v
                Fake Provider             qBittorrent
                                                |
                                                v
                                           NAS Storage
                                                |
                              +-----------------+----------------+
                              |                                  |
                              v                                  v
                           Movies                               TV
                              |                                  |
                              +-----------------+----------------+
                                                |
                                                v
                                              Lampa

Ответственность компонентов
Telegram
Отвечает только за:
- команды;
- сообщения;
- inline keyboard;
- FSM;
- отображение состояния;
- получение пользовательского ввода.
Telegram handlers не должны напрямую:
- обращаться к qBittorrent;
- работать с NAS;
- перемещать файлы;
- содержать бизнес-логику торрент-загрузок.
Application Services
Отвечают за бизнес-логику:
- создание torrent task;
- проверку данных;
- duplicate detection;
- управление состоянием;
- вызов provider;
- orchestration.
Domain
Содержит:
- модели;
- enum;
- бизнес-сущности;
- бизнес-правила.
Providers
Изолируют внешние системы.
Минимум:
FakeTorrentProvider
QBittorrentProvider

Telegram и application layer не должны знать внутренние детали qBittorrent.
Storage
Отвечает за хранение файлов.
В production:
Proxmox
   |
   v
NAS mount
   |
   +--> Downloads
   +--> Movies
   +--> TV

3. Текущий этап проекта
На первом этапе необходимо реализовать только Telegram UI и обработку magnet-ссылок.
Критически важно
На MVP:
TEST_MODE=true

В этом режиме запрещено:
- подключение к qBittorrent;
- реальные torrent downloads;
- запись torrent-файлов на NAS;
- доступ к NAS;
- перемещение медиа;
- реальные filesystem downloads.
Можно только симулировать загрузку.
4. MVP Flow
Основной пользовательский сценарий:
/start
    |
    v
Пользователь отправляет magnet
    |
    v
Magnet Parser
    |
    v
Показать информацию
    |
    v
Выбор:
  Фильм / Сериал
    |
    v
Выбор качества
    |
    v
Подтверждение
    |
    v
Fake Torrent Service
    |
    v
Test Queue
    |
    v
/downloads

5. Telegram UI
/start
Сообщение:
🤖 Torrent Bot

Отправьте magnet-ссылку, чтобы создать новую загрузку.

Кнопки:
📥 Мои загрузки
⚙️ Настройки
ℹ️ Помощь

Magnet
После отправки magnet:
🔗 Torrent найден

🎬 Название:
The.Matrix.1999.1080p.BluRay

🔑 Hash:
A1B2C3D4E5...

📡 Trackers:
3

💾 Размер:
не определён

Куда сохранить?

Кнопки:
🎬 Фильм
📺 Сериал
❌ Отмена

Выбор качества
Выберите качество:

Кнопки:
2160p / 4K
1080p
720p
Не определять

Confirmation
✅ Готово к добавлению

🎬 Название:
The Matrix

📁 Категория:
Фильм

🎞 Качество:
1080p

💾 Путь:
Movies/

🧪 TEST MODE

Файл реально скачиваться НЕ будет.

Кнопки:
⬇️ Добавить в очередь
✏️ Изменить
❌ Отмена

6. Magnet Parser
Файл:
app/services/magnet_parser.py

Parser должен извлекать:
- xt;
- btih;
- dn;
- tr;
- xl;
- остальные query parameters без потери информации.
Модель:
@dataclassclass MagnetInfo:    raw: str    info_hash: str    display_name: str | None    trackers: list[str]    size_bytes: int | None


Parser должен:
- проверять magnet:?;
- проверять наличие info hash;
- декодировать URL encoding;
- поддерживать несколько trackers;
- корректно работать без dn;
- корректно работать без tr;
- корректно работать без xl;
- сохранять исходный magnet;
- не падать из-за неизвестных параметров.
7. Domain Models
Использовать enum вместо произвольных строк.
class MediaType(str, Enum):    MOVIE = "movie"    SERIES = "series"


class Quality(str, Enum):    UHD_2160 = "2160p"    FHD_1080 = "1080p"    HD_720 = "720p"    UNKNOWN = "unknown"


class TorrentStatus(str, Enum):    QUEUED = "queued"    DOWNLOADING = "downloading"    PAUSED = "paused"    COMPLETED = "completed"    FAILED = "failed"    CANCELLED = "cancelled"


Torrent task должен содержать минимум:
class TorrentTask:    id: int    user_id: int    magnet: str    info_hash: str    name: str    media_type: MediaType    quality: Quality    status: TorrentStatus    progress: float    size_bytes: int | None    download_speed: int    upload_speed: int    eta_seconds: int | None    created_at: datetime    completed_at: datetime | None


8. Fake Torrent Provider
На MVP создать:
app/providers/fake.py

Fake provider:
- не делает HTTP-запросов;
- не подключается к qBittorrent;
- не обращается к NAS;
- не создаёт реальные torrent downloads;
- симулирует очередь;
- может симулировать прогресс.
Например:
⏳ Downloading

The Matrix

██████░░░░░░░░░░ 35%

35%

Speed: TEST
ETA: TEST

🧪 TEST MODE

9. Provider Interface
Telegram handlers никогда не должны напрямую обращаться к qBittorrent.
Создать интерфейс:
class TorrentProvider(Protocol):    async def add_magnet(        self,        magnet: str,        save_path: str,    ) -> str:        ...    async def get_status(        self,        torrent_id: str,    ) -> TorrentStatus:        ...    async def pause(        self,        torrent_id: str,    ) -> None:        ...    async def resume(        self,        torrent_id: str,    ) -> None:        ...    async def delete(        self,        torrent_id: str,        delete_files: bool = False,    ) -> None:        ...


В будущем:
TorrentProvider
      |
      +--> FakeTorrentProvider
      |
      +--> QBittorrentProvider

10. Torrent Service
Файл:
app/services/torrent_service.py

Ответственность:
- создание torrent task;
- валидация;
- duplicate detection;
- вызов provider;
- преобразование provider status;
- application-level operations.
Torrent service не должен содержать Telegram-код.
11. FSM
Использовать aiogram FSM.
class TorrentWizard(StatesGroup):    choosing_category = State()    choosing_quality = State()    confirming = State()


Нельзя использовать глобальные Python variables для хранения состояния пользователя.
На MVP можно использовать:
MemoryStorage

Позже можно перейти на Redis или другой persistent storage.
12. /downloads
Команда должна показывать только загрузки текущего Telegram user.
Пример:
📥 Мои загрузки

#001
🎬 The Matrix
1080p
⏳ 35%

#002
📺 Breaking Bad
1080p
✅ Завершено

Для активных:
📊 Открыть
⏸ Пауза
❌ Отменить

Для завершённых:
📊 Открыть
🗑 Удалить из списка

Перед каждой операцией обязательно проверять ownership.
13. Duplicate Detection
Главным идентификатором torrent является:
info_hash

Не использовать название как primary identity.
Алгоритм:
info_hash
    |
    v
Существует task?
    |
    +---- YES ---> показать существующий task
    |
    +---- NO ----> создать новый task

14. Callback Data
Callback data должна быть короткой.
Разрешено:
torrent:category:movie
torrent:category:series

torrent:quality:2160
torrent:quality:1080
torrent:quality:720

torrent:confirm:add
torrent:confirm:cancel
torrent:confirm:back

Не помещать в callback:
- полный magnet;
- большие JSON;
- секреты;
- большие metadata objects.
Данные должны храниться в FSM или database.
15. Security
Никогда нельзя доверять данным от пользователя:
- названию torrent;
- имени файла;
- callback data;
- путям;
- тексту;
- metadata.
При будущей работе с filesystem обязательно защищаться от:
../../

и абсолютных путей.
Никогда не логировать:
BOT_TOKEN
QBITTORRENT_PASSWORD
NAS_PASSWORD
REDIS_PASSWORD
DATABASE_PASSWORD

16. Configuration
Использовать .env.
Минимум:
BOT_TOKEN=change_me

APP_ENV=development
LOG_LEVEL=INFO

TEST_MODE=true

Будущая конфигурация:
QBITTORRENT_URL=
QBITTORRENT_USERNAME=
QBITTORRENT_PASSWORD=

DOWNLOAD_PATH=/mnt/dns320/Downloads
MOVIES_PATH=/mnt/dns320/Movies
TV_PATH=/mnt/dns320/TV

REDIS_URL=redis://localhost:6379/0
DATABASE_URL=

.env никогда не должен попадать в git.
Обязательно создать:
.env.example

17. Project Structure
torrent-telegram-bot/
│
├── app/
│   ├── __init__.py
│   ├── main.py
│   ├── bot.py
│   ├── config.py
│   │
│   ├── handlers/
│   │   ├── __init__.py
│   │   ├── start.py
│   │   ├── magnet.py
│   │   ├── downloads.py
│   │   └── settings.py
│   │
│   ├── keyboards/
│   │   ├── __init__.py
│   │   ├── main.py
│   │   ├── torrent.py
│   │   └── downloads.py
│   │
│   ├── services/
│   │   ├── __init__.py
│   │   ├── magnet_parser.py
│   │   ├── torrent_service.py
│   │   └── media_service.py
│   │
│   ├── providers/
│   │   ├── __init__.py
│   │   ├── base.py
│   │   └── fake.py
│   │
│   └── models/
│       ├── __init__.py
│       ├── torrent.py
│       └── media.py
│
├── tests/
│   ├── unit/
│   ├── integration/
│   └── fixtures/
│
├── docs/
│   ├── DEVELOPMENT.md
│   ├── ARCHITECTURE.md
│   ├── API.md
│   ├── CONFIGURATION.md
│   ├── DEPLOYMENT.md
│   ├── TESTING.md
│   ├── TROUBLESHOOTING.md
│   ├── ROADMAP.md
│   ├── features/
│   └── decisions/
│
├── .env.example
├── .gitignore
├── pyproject.toml
├── Dockerfile
├── docker-compose.yml
├── README.md
└── requirements.txt

18. Документация features
Это обязательное правило проекта.
В:
docs/features/

каждая feature должна иметь отдельный Markdown-файл.
Нельзя создавать один огромный:
docs/features.md

Вместо этого:
docs/features/
├── 001-start-menu.md
├── 002-magnet-parser.md
├── 003-media-type-selection.md
├── 004-quality-selection.md
├── 005-confirmation-flow.md
├── 006-test-download-queue.md
├── 007-downloads-list.md
├── 008-download-controls.md
├── 009-user-isolation.md
├── 010-persistent-state.md
├── 011-qbittorrent-provider.md
├── 012-nas-storage.md
├── 013-media-organizer.md
├── 014-movie-library.md
├── 015-series-library.md
├── 016-background-worker.md
├── 017-notifications.md
├── 018-duplicate-detection.md
└── 019-lampa-library.md

ID feature после создания менять нельзя.
19. Feature Template
Каждый файл feature должен содержать:
# FEAT-XXX — Название

## Status

Planned

## Goal

Какую проблему решает feature.

## User Story

As a user, I want...

## Scope

Что входит в feature.

## Out of Scope

Что не входит.

## User Flow

```text
step
  |
  v
step

UI
Описание Telegram UI.
Domain Changes
Какие модели и enum нужны.
API / Interfaces
Какие сервисы/provider interfaces нужны.
Storage Changes
Изменения database/filesystem.
Security
Security considerations.
Error Handling
Ошибки и поведение UI.
Tests
Какие тесты обязательны.
Dependencies
От каких feature зависит.
Configuration
Какие env variables нужны.
Implementation Tasks
- [ ] Task 1
- [ ] Task 2
- [ ] Task 3
Acceptance Criteria
- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Criterion 3
Related Documentation
Ссылки на связанные документы.

---

# 20. Feature Workflow

Каждая существенная feature проходит:

```text
Feature idea
    |
    v
Создать docs/features/XXX-name.md
    |
    v
Описать Acceptance Criteria
    |
    v
Определить dependencies
    |
    v
Implement
    |
    v
Write tests
    |
    v
Update documentation
    |
    v
Lint
    |
    v
Type check
    |
    v
Tests
    |
    v
Feature = Implemented

21. Feature Status
Использовать только:
Planned
In Progress
Implemented
Blocked
Deprecated

22. Первоначальный список Features
FEAT-001 Start Menu
FEAT-002 Magnet Parser
FEAT-003 Media Type Selection
FEAT-004 Quality Selection
FEAT-005 Confirmation Flow
FEAT-006 Test Download Queue
FEAT-007 Downloads List
FEAT-008 Download Controls
FEAT-009 User Isolation
FEAT-010 Persistent State
FEAT-011 qBittorrent Provider
FEAT-012 NAS Storage
FEAT-013 Media Organizer
FEAT-014 Movie Library
FEAT-015 Series Library
FEAT-016 Background Worker
FEAT-017 Notifications
FEAT-018 Duplicate Detection
FEAT-019 Lampa Library

FEAT-001 — FEAT-009 являются первоначальной зоной MVP.
FEAT-010 и выше не реализовывать, пока пользователь явно не переведёт проект на следующий этап.
23. Documentation Structure
docs/
│
├── DEVELOPMENT.md
├── ARCHITECTURE.md
├── API.md
├── CONFIGURATION.md
├── DEPLOYMENT.md
├── TESTING.md
├── TROUBLESHOOTING.md
├── ROADMAP.md
│
├── features/
│   ├── 001-start-menu.md
│   ├── 002-magnet-parser.md
│   ├── ...
│   └── 019-lampa-library.md
│
└── decisions/
    ├── 0001-architecture.md
    ├── 0002-torrent-provider.md
    └── 0003-storage.md

24. Назначение документации
DEVELOPMENT.md
Главный development contract проекта.
Содержит:
- глобальную цель;
- архитектурные правила;
- правила агента;
- Definition of Done.
ARCHITECTURE.md
Содержит:
- компоненты;
- зависимости;
- boundaries;
- data flow;
- provider architecture.
API.md
Содержит:
- internal service interfaces;
- provider interfaces;
- contracts.
CONFIGURATION.md
Содержит:
- env variables;
- defaults;
- production configuration.
DEPLOYMENT.md
Содержит:
- local запуск;
- Docker;
- Proxmox;
- NAS;
- production deployment.
TESTING.md
Содержит:
- testing strategy;
- unit tests;
- integration tests;
- external integration tests.
TROUBLESHOOTING.md
Содержит:
- известные ошибки;
- диагностику;
- типовые проблемы.
ROADMAP.md
Содержит этапы развития проекта.
features/
Каждая отдельная feature — отдельный файл.
decisions/
Architecture Decision Records.
25. ADR
Значимые архитектурные решения необходимо сохранять в:
docs/decisions/

Например:
0001-architecture.md
0002-torrent-provider.md
0003-storage.md

Шаблон:
# ADR-0001: Название

## Status

Accepted

## Context

Почему потребовалось решение.

## Decision

Что было решено.

## Consequences

Плюсы, минусы и последствия.

26. Testing
Структура:
tests/
├── unit/
│   ├── test_magnet_parser.py
│   └── test_torrent_service.py
│
├── integration/
│   └── test_bot_flow.py
│
└── fixtures/

Минимальные тесты Magnet Parser:
[ ] valid magnet
[ ] missing dn
[ ] multiple trackers
[ ] URL encoded dn
[ ] missing info hash
[ ] invalid scheme
[ ] xl parameter

Integration tests должны проверять:
/start
    |
    v
magnet
    |
    v
category
    |
    v
quality
    |
    v
confirmation
    |
    v
queue
    |
    v
/downloads

Тесты MVP не должны требовать:
- qBittorrent;
- NAS;
- реального torrent client.
27. Python Tooling
Использовать:
Python 3.12+
aiogram 3.x
pytest
pytest-asyncio
pytest-cov
Ruff
mypy

Предпочтительно хранить конфигурацию инструментов в:
pyproject.toml

Команды:
ruff check .

ruff format .

ruff format --check .

mypy app

pytest

pytest -v

pytest --cov=app --cov-report=term-missing

Перед завершением feature:
ruff check .
ruff format --check .
mypy app
pytest

Все команды должны завершаться успешно.
28. Dependencies
Runtime dependencies:
aiogram
python-dotenv

Development:
pytest
pytest-asyncio
pytest-cov
ruff
mypy

Не добавлять dependency без необходимости.
В будущем могут появиться:
redis
sqlalchemy
alembic
httpx

Но только при реализации соответствующей feature.
29. Logging
Использовать Python logging.
Не использовать:
print(...)


для application logging.
Можно логировать:
event
user_id
torrent_id
info_hash
status
error type

Нельзя логировать:
BOT_TOKEN
passwords
NAS credentials
private keys

30. Database
На MVP database необязательна.
Но архитектура должна позволять добавить database без переписывания Telegram UI.
В production ожидаются сущности:
users
torrents
torrent_events
settings
media_items

История бизнес-состояния должна храниться в database, а не в Redis.
31. Background Worker
Когда появятся реальные загрузки, нельзя постоянно опрашивать qBittorrent из Telegram handlers.
Целевая архитектура:
Telegram Bot
     |
     v
Database
     |
     v
Background Worker
     |
     v
qBittorrent

Worker:
- получает torrent state;
- обновляет database;
- отправляет Telegram notifications;
- обрабатывает завершённые torrents.
Не редактировать Telegram message каждую секунду.
32. qBittorrent
Будущая реализация:
app/providers/qbittorrent.py

Архитектура:
TorrentService
      |
      v
TorrentProvider
      |
      v
QBittorrentProvider
      |
      v
qBittorrent Web API

Telegram layer не должен зависеть от конкретной реализации qBittorrent.
33. NAS
NAS:
D-Link DNS-320

Torrent client работает на Proxmox.
NAS монтируется на Proxmox через SMB или NFS.
Пример:
/mnt/dns320/
├── Downloads/
├── Movies/
└── TV/

Application должна работать через configuration:
DOWNLOAD_PATH=/mnt/dns320/Downloads
MOVIES_PATH=/mnt/dns320/Movies
TV_PATH=/mnt/dns320/TV

Application не должна содержать D-Link-specific logic.
34. Download Flow
После появления реального qBittorrent:
Magnet
   |
   v
Parse
   |
   v
Select category
   |
   v
Select quality
   |
   v
Create task
   |
   v
qBittorrent
   |
   v
Downloads/
   |
   v
100% complete
   |
   v
Media Organizer
   |
   +--> Movies/
   |
   +--> TV/

Не перемещать файл в media library до завершения torrent.
35. Media Organizer
Будущий service:
MediaOrganizer

Ответственность:
- определить media type;
- определить movie/series;
- определить season;
- нормализовать название;
- создать директории;
- переместить файлы.
Movie:
Movies/
└── The Matrix (1999)/
    └── The Matrix (1999).mkv

Series:
TV/
└── Breaking Bad/
    ├── Season 01/
    │   ├── Breaking Bad - S01E01.mkv
    │   └── Breaking Bad - S01E02.mkv
    │
    └── Season 02/

36. Lampa
Lampa является downstream consumer.
Не делать прямую связь:
Telegram -> Lampa

Правильная архитектура:
Torrent
   |
   v
Media Organizer
   |
   v
Media Library
   |
   v
Lampa

Сначала должна стабилизироваться media library.
Только после этого реализовывать Lampa integration.
37. CI
CI должен выполнять:
checkout
    |
    v
install dependencies
    |
    +--> ruff check
    |
    +--> ruff format --check
    |
    +--> mypy
    |
    +--> pytest
    |
    v
success

Unit tests не должны требовать внешней инфраструктуры.
38. Git
Использовать Conventional Commits.
Примеры:
feat: add magnet parser
feat: add torrent queue UI
fix: handle invalid magnet links
test: add parser tests
refactor: extract torrent provider
docs: add feature specification
chore: configure ruff

Никогда не commit:
.env
tokens
passwords
private keys
credentials

39. Rules для Coding Agent
Перед началом работы агент обязан прочитать:
docs/DEVELOPMENT.md

Перед реализацией конкретной feature:
docs/features/XXX-name.md

Если меняется архитектура:
docs/ARCHITECTURE.md

Если принимается существенное архитектурное решение:
docs/decisions/

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
40. Current Scope
На текущем этапе реализовать только:
/start
    |
    v
Receive magnet
    |
    v
Parse magnet
    |
    v
Display metadata
    |
    v
Choose Movie / Series
    |
    v
Choose Quality
    |
    v
Confirmation
    |
    v
FakeTorrentProvider
    |
    v
Test Queue
    |
    v
/downloads

Не реализовывать сейчас:
qBittorrent
NAS
SMB
NFS
real downloads
media organizer
Lampa

если пользователь отдельно не перевёл проект на соответствующую фазу.
41. Definition of Done
Feature считается завершённой только если:
[ ] Code implemented
[ ] Unit tests added
[ ] Integration tests added where necessary
[ ] Error handling implemented
[ ] Logging implemented
[ ] Security considered
[ ] Documentation updated
[ ] Feature documentation updated
[ ] Ruff check passes
[ ] Ruff format check passes
[ ] Mypy passes
[ ] Existing tests pass

42. Roadmap
Phase 0 — Telegram UI
[ ] Start menu
[ ] Magnet parser
[ ] Movie / Series
[ ] Quality
[ ] Confirmation
[ ] Fake queue
[ ] Downloads list
[ ] Download controls
[ ] Tests
[ ] Documentation

Phase 1 — Persistent State
[ ] Database
[ ] Users
[ ] Torrents
[ ] Settings
[ ] Persistent FSM

Phase 2 — qBittorrent
[ ] Provider
[ ] Add magnet
[ ] Status
[ ] Pause
[ ] Resume
[ ] Cancel
[ ] Delete
[ ] Progress
[ ] Speed
[ ] ETA

Phase 3 — NAS
[ ] NAS mount
[ ] Downloads
[ ] Movies
[ ] TV
[ ] Storage abstraction
[ ] Disk space checks

Phase 4 — Media Organizer
[ ] Movie detection
[ ] Series detection
[ ] Season detection
[ ] Renaming
[ ] Folder creation
[ ] Completion processing

Phase 5 — Lampa
[ ] Stable media library
[ ] Lampa integration

Phase 6 — Production
[ ] Docker
[ ] Health checks
[ ] Monitoring
[ ] Structured logging
[ ] Backups
[ ] Background worker
[ ] CI/CD

43. Final Development Principle
Главная задача MVP — проверить UX и архитектурный workflow.
MVP не должен быть одноразовым прототипом.
Развитие должно идти:
Telegram UI
      +
Magnet Parser
      +
Fake Queue
      |
      v
Persistent State
      |
      v
qBittorrent
      |
      v
NAS
      |
      v
Media Organizer
      |
      v
Movies / TV
      |
      v
Lampa

Любое следующее расширение должно добавляться так, чтобы уже написанный Telegram UI и бизнес-слой не пришлось переписывать с нуля.
```
