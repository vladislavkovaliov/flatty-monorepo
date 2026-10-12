# API

Internal service interfaces, provider interfaces, contracts.

---

## 1. Magnet Parser

Файл: `app/services/magnet_parser.py`

### Модель

```python
@dataclass
class MagnetInfo:
    raw: str
    info_hash: str
    display_name: str | None
    trackers: list[str]
    size_bytes: int | None
```

### Что извлекает

- `xt`
- `btih`
- `dn`
- `tr`
- `xl`
- остальные query parameters без потери информации.

### Правила

- проверять схему `magnet:?`;
- проверять наличие info hash;
- декодировать URL encoding;
- поддерживать несколько trackers;
- корректно работать без `dn`, без `tr`, без `xl`;
- сохранять исходный magnet;
- не падать из-за неизвестных параметров.

---

## 2. Torrent Provider Interface

Telegram handlers никогда не должны напрямую обращаться к qBittorrent.

```python
class TorrentProvider(Protocol):
    async def add_magnet(self, magnet: str, save_path: str) -> str: ...
    async def get_status(self, torrent_id: str) -> TorrentStatus: ...
    async def pause(self, torrent_id: str) -> None: ...
    async def resume(self, torrent_id: str) -> None: ...
    async def delete(self, torrent_id: str, delete_files: bool = False) -> None: ...
```

Реализации:

```text
TorrentProvider
   +--> FakeTorrentProvider   (app/providers/fake.py, MVP)
   +--> QBittorrentProvider   (app/providers/qbittorrent.py, будущее)
```

---

## 3. Torrent Service

Файл: `app/services/torrent_service.py`

Ответственность:

- создание torrent task;
- валидация;
- duplicate detection;
- вызов provider;
- преобразование provider status;
- application-level operations.

Torrent service **не должен** содержать Telegram-код.

---

## 4. Callback Data Contract

Формат: короткие строки, разделённые `:`.

| Callback | Действие |
|---|---|
| `torrent:category:movie` | выбрать «Фильм» |
| `torrent:category:series` | выбрать «Сериал» |
| `torrent:quality:2160` | качество 2160p / 4K |
| `torrent:quality:1080` | качество 1080p |
| `torrent:quality:720` | качество 720p |
| `torrent:confirm:add` | добавить в очередь |
| `torrent:confirm:cancel` | отмена |
| `torrent:confirm:back` | назад |
| `torrent:quality:unknown` | «Не определять» (Quality.UNKNOWN) |
| `torrent:open:<id>` | открыть детальный вид задачи |
| `torrent:list:back` | вернуться к списку загрузок |
| `torrent:pause:<id>` / `torrent:resume:<id>` | пауза / продолжить |
| `torrent:cancel:<id>` / `torrent:delete:<id>` | отменить / удалить из списка |
| `menu:downloads` / `menu:settings` / `menu:help` | пункты главного меню (FEAT-001) |

Запрещено класть в callback: полный magnet, большие JSON, секреты, большие metadata objects. Данные — в FSM или database.

---

## 5. User Isolation Contract

- `/downloads` показывает только загрузки текущего Telegram user.
- Перед каждой операцией (pause / resume / cancel / delete / open) обязательно проверять ownership.

---

## 6. Storage Configuration Contract

Application обращается к storage только через configuration:

```text
DOWNLOAD_PATH=/mnt/dns320/Downloads
MOVIES_PATH=/mnt/dns320/Movies
TV_PATH=/mnt/dns320/TV
```

Никакой vendor-specific logic в коде.

---

## 7. HTTP API

FastAPI-приложение `app/api/` поверх того же `TorrentService` (FEAT-020, ADR-0004).

| Метод | Путь | Успех | Ошибки |
|---|---|---|---|
| GET | `/api/tasks` | 200, список задач | — |
| GET | `/api/tasks/{id}` | 200 | 404 |
| POST | `/api/tasks` | 201 | 400 magnet/пользователь, 409 дубль, 422 валидация |
| POST | `/api/tasks/{id}/pause` | 200 | 404, 409 неверный статус |
| POST | `/api/tasks/{id}/resume` | 200 | 404, 409 |
| POST | `/api/tasks/{id}/cancel` | 200 | 404, 409 (уже отменена) |
| DELETE | `/api/tasks/{id}` | 204 | 404 |

- Актор — `API_USER_ID`; ownership = ровно поведение Telegram (§5).
- Swagger UI: `/docs`, схема: `/openapi.json`. Только localhost, без аутентификации.
- Неожиданная ошибка → 500 `{"detail": "internal error"}` + traceback в лог.
- WebSocket (`/api/ws`) — будущая задача, REST-контракт не изменится.
