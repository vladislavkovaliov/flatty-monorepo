# Задача 13. Миграции БД: goose вместо ручного slice

> Тема: **goose v3** — заменить ручной список миграций на фреймворк
> с версионированием и накатом.
> Результат: миграции накатываются один раз, с отслеживанием применённых версий,
> в порядке, который задан номерами файлов, с возможностью откатить последний шаг.
>
> Это учебная задача. Все правки делаешь ты. Ниже — пошаговый разбор с командами,
> ожидаемым выводом и проверкой после каждого шага. Если шаг не сходится с тем, что
> видишь у себя, — не иди дальше, сначала разберись.

---

## Зачем

Для solution-роли «миграции руками» — красный флаг. Сейчас в проекте **два**
независимых механизма наката, и оба без версионирования:

1. `apps/go-api/cmd/migrate/main.go` — ручной `slice` путей к SQL-файлам.
   Порядок задан вручную и не совпадает с нумерацией файлов:
   `005 → 002 → 003 → 006 → 007 → 004 → 008`, а `001` в списке вообще нет.
   Раннер не помнит, что уже применено: он выполняет все файлы заново при каждом
   запуске. Это держится только на идемпотентности SQL (`IF NOT EXISTS`,
   `ON CONFLICT`).
2. `scripts/init-db.sh` и `scripts/init-db-auth.sh` — те же миграции, но через
   `psql` в docker-compose. Причём **другим набором**: `init-db` гоняет `006` и
   `008` после `db.sql`, `init-db-auth` ждёт таблицу `"user"` от nest-graphql и
   гоняет `004` и `007`.

Три симптома одной болезни:

- **нет таблицы версий** — никто не знает, что применено, поэтому всё пишется
  идемпотентно, а первая неидемпотентная миграция (`ALTER TABLE ... DROP COLUMN`,
  `UPDATE` без `WHERE`) сломает повторный запуск;
- **порядок живёт в голове** — комментарий «005 must run before 002/003» в коде
  раннера и в шапке файла;
- **два владельца схемы** — `db.sql` + `init-db` + `init-db-auth` + `cmd/migrate`.
  Рассинхрон неизбежен: скрипты и раннер уже разъехались.

Фреймворк убирает первые два пункта одной зависимостью: таблица
`goose_db_version` хранит применённые версии, порядок — числовой префикс файла,
`up`/`down` — готовые операции. Третий пункт (один владелец) — ручная работа,
но она сводится к удалению лишнего кода.

На собеседовании ответ «goose с embed-миграциями, базовая схема — отдельный
`db.sql`, дальше всё через goose» — это то, что хотят услышать.

## Контекст проекта

Проверь сам, прежде чем верить списку:

- `apps/go-api/cmd/migrate/main.go` — текущий раннер: `pgxpool.New` + `os.ReadFile`
  по путям `migrations/NNN_*.sql`.
- `apps/go-api/migrations/` — 8 файлов `001…008`, все идемпотентные, без
  goose-аннотаций.
- `apps/go-api/go.mod` — `go 1.26.0`, `github.com/jackc/pgx/v5 v5.10.0`.
  goose пока нет.
- `apps/go-api/project.json` — таргеты `build`, `serve`, `test`, `lint`
  (`nx:run-commands`, `cwd: apps/go-api`). Таргета для миграций нет.
- `apps/go-api/internal/config/config.go` — `LoadConfig()` читает `DATABASE_URL`
  (по умолчанию `postgres://postgres:password@localhost:55000/flatty?sslmode=disable`)
  и `PORT`.
- `db.sql` (корень) — 583 строки базовой схемы. Важно: он **уже** содержит
  финальную схему (`expenses.description`, `expense_monthly_totals` /
  `expense_monthly_averages` с `PRIMARY KEY (resident_location_id, month, year)`),
  то есть `db.sql` — это «версия 0», а не древняя схема.
- `scripts/init-db.sh` — гоняет `/db.sql`, потом `psql` по `006` и `008`,
  создаёт базу `torrent_bot`.
- `scripts/init-db-auth.sh` — ждёт таблицу `"user"`, гоняет `004` (если нет
  `user_settings`) и `007`.
- `docker-compose.yml` — сервисы `init-db` и `init-db-auth` монтируют
  `./apps/go-api/migrations` в `/migrations:ro`.

Окружение для всех команд ниже:

```bash
cd /Users/vladislavkovaliov/WebstormProjects/flatty-budget-monorepo
export DATABASE_URL='postgres://postgres:password@localhost:55000/flatty?sslmode=disable'
```

## Целевое состояние

```
apps/go-api/
├── migrations/
│   ├── embed.go                       # package migrations, //go:embed *.sql
│   ├── 001_add_description_to_expenses.sql
│   ├── 002_add_resident_location_to_expense_stats.sql   # было 005
│   ├── 003_backfill_monthly_totals.sql                  # было 002
│   ├── 004_backfill_monthly_averages.sql                # было 003
│   ├── 005_add_applications.sql                         # было 006
│   ├── 006_add_user_id_to_expenses.sql                  # было 007
│   ├── 007_add_user_settings.sql                        # было 004
│   └── 008_add_vue_torrent_bot_application.sql
├── cmd/migrate/main.go                # sql.DB + goose.Up, без slice
└── project.json                       # + таргет migrate-db

scripts/init-db.sh                     # только db.sql + torrent_bot
scripts/init-db-auth.sh                # удалён
docker-compose.yml                     # + сервис migrate, без init-db-auth
```

Единственный владелец миграций — goose. `db.sql` отвечает только за базовую схему
и сид-данные.

---

## Шаг 0. Осмотреть текущее состояние

Не начинай правки, пока не увидел своими глазами, что применено в БД. Иначе
первый же запуск goose сделает не то, что ты ожидаешь.

```bash
cd apps/go-api
cat cmd/migrate/main.go
ls migrations/
```

Ожидаемо: slice из 7 путей (без `001`) и 8 файлов в каталоге.

Теперь посмотри на базу:

```bash
psql "$DATABASE_URL" -c '\dt'
psql "$DATABASE_URL" -c "SELECT to_regclass('goose_db_version') AS goose_table;"
psql "$DATABASE_URL" -c '\d expense_monthly_totals'
psql "$DATABASE_URL" -c '\d user_settings'
```

Что ты должен увидеть и понять:

- `goose_table` = `NULL` — таблицы версий нет, значит goose считает базу пустой;
- у `expense_monthly_totals` первичный ключ `(resident_location_id, month, year)` —
  то есть `005` уже применена (её и накатывал `db.sql` / `init-db`);
- `user_settings` существует, только если ты когда-нибудь гонял `cmd/migrate`
  руками или поднимал сервис `migrate` в compose. Если нет — `007` ещё не
  применена.

**Почему это важно.** goose не умеет «угадывать» состояние схемы. Его состояние —
только `goose_db_version`. Пустая таблица = «накатывай всё с нуля». Поэтому шаг 7
(пересоздание БД или baseline) — не формальность, а обязательная часть перехода.

## Шаг 1. Выбрать инструмент

Берём **goose v3** (`github.com/pressly/goose/v3`), и вот почему:

- **Библиотека, а не CLI.** `goose.Up(db, dir)` вызывается из твоего кода,
  миграции едут внутри бинаря через `embed`. Не надо ставить goose на каждую
  машину и в CI — достаточно `go build`.
- **Работает поверх `database/sql`.** Для Postgres есть готовый диалект;
  с pgx дружит через `pgx/v5/stdlib`.
- **`embed` из коробки.** `goose.SetBaseFS(fs)` — единственное, что нужно.
- **Понятная модель:** один файл = одна версия, `-- +goose Up` / `-- +goose Down`,
  таблица `goose_db_version`.

`golang-migrate` — рабочая альтернатива, но он чаще живёт как CLI, а `embed`
и программный запуск там менее удобны. Для учебной задачи важнее «увидеть
внутренности», а goose их не прячет.

## Шаг 2. Добавить зависимость

```bash
cd apps/go-api
go get github.com/pressly/goose/v3@latest
go list -m github.com/pressly/goose/v3
```

Ожидаемо: строка `github.com/pressly/goose/v3 v3.x.y` в выводе `go list`.

Проверь `go.mod`: goose попал в `require`, а транзитивные зависимости — в `go.sum`.
Ничего вручную не правь: `go get` делает это корректно, а ручная правка `go.mod`
часто ломает `go.sum`.

**Почему не глобальный бинарь.** `go install github.com/pressly/goose/v3/cmd/goose@latest`
тоже вариант, и он удобен, когда хочется гонять `goose status` из терминала.
Но тогда версия инструмента не зафиксирована в репозитории, а миграции читаются
с диска — то есть раннер снова зависит от `cwd`. Для этого проекта правильный
путь — библиотека + embed.

## Шаг 3. Обернуть миграции в Go-пакет с embed

`//go:embed` не умеет читать родительские каталоги (`..` запрещён). Поэтому
`embed`-переменная должна лежать в пакете, внутри которого находятся `.sql`.
Самый дешёвый вариант — новый файл прямо в `apps/go-api/migrations/`.

Создай `apps/go-api/migrations/embed.go`. Три части, **строго в таком порядке**:

1. `package migrations`
2. импорт `embed`
3. пустая строка, затем `//go:embed *.sql` **вплотную** над объявлением
   `var FS embed.FS`.

Порядок здесь не косметика. `//go:embed` — это директива, привязанная к
**следующему объявлению уровня пакета**, а не к файлу. Поэтому:

- `package` обязан быть первой строкой;
- между `//go:embed` и `var` не должно быть ни пустой строки, ни другого
  комментария — иначе привязка рвётся.

Типичная ошибка — поставить директиву в самое начало файла, выше `package`:

```
//go:embed *.sql      ← так нельзя
package migrations
```

Компилятор ответит `misplaced go:embed directive`, потому что приклеить директиву
не к чему: под ней не объявление, а `package`.

Проверка:

```bash
cd apps/go-api
go build ./migrations
go vet ./migrations
```

Ожидаемо: пустой вывод (успех). Если видишь `pattern *.sql: no matching files` —
файл лежит не в том каталоге.

**Почему `*.sql`, а не `migrations/*.sql`.** От этого зависит аргумент `dir`
в вызовах goose:

| Что в `//go:embed` | Аргумент `dir` в `goose.Up` |
|---|---|
| `*.sql` (файл лежит в `migrations/`) | `"."` |
| `migrations/*.sql` (файл лежит в корне модуля) | `"migrations"` |

goose делает `fs.ReadDir(dir)` по встроенной ФС, и путь должен совпадать
с тем, как файлы в неё попали. Это самая частая ошибка при первом знакомстве:
«файлы есть, а goose их не видит».

## Шаг 4. Разметить SQL-файлы аннотациями goose

goose не парсит SQL целиком — он режет файл по аннотациям и по `;`. Правила:

1. В каждом файле **ровно один** `-- +goose Up`, и он идёт первым.
2. `-- +goose Down` опционален, но обязан идти после `Up`.
3. Всё, что до `Up`, goose игнорирует — там можно оставить шапку с комментариями.
4. PL/pgSQL-блоки (`$$ ... $$`), внутри которых есть `;`, нужно обернуть в
   `-- +goose StatementBegin` / `-- +goose StatementEnd`. Иначе goose разрежет
   функцию по внутренним точкам с запятой и получит синтаксическую ошибку.

Начни с простого файла — `001_add_description_to_expenses.sql`. Было:

```sql
ALTER TABLE expenses ADD COLUMN IF NOT EXISTS description TEXT DEFAULT '';
```

Стало (добавь аннотации и честный down):

```sql
-- +goose Up
ALTER TABLE expenses ADD COLUMN IF NOT EXISTS description TEXT DEFAULT '';

-- +goose Down
ALTER TABLE expenses DROP COLUMN IF EXISTS description;
```

Дальше пройди по всем восьми. Ниже — что делать с `Down` в каждом случае.
Это главное упражнение шага: **down — это не формальность, а утверждение
о том, что откат возможен**.

| Файл | Down | Комментарий |
|---|---|---|
| `001_add_description_to_expenses` | `DROP COLUMN IF EXISTS description` | Обратимо (потеряются описания — но это и есть откат). |
| `002` (старый `005`, rescope статистики) | **пустой + комментарий** | Исходный ключ `(month, year)` восстановить нельзя: строки уже разбиты по `resident_location_id`. Честный откат — пересоздание БД. |
| `003` (старый `002`, backfill totals) | `DROP COLUMN IF EXISTS created_at` | Данные — производные от `expenses`, их можно пересчитать. |
| `004` (старый `003`, backfill averages) | `DROP COLUMN IF EXISTS created_at` | То же. |
| `005` (старый `006`, applications) | `DROP TABLE IF EXISTS applications` | Триггер уйдёт вместе с таблицей. Функцию `update_updated_at_column` **не** удаляй — её использует `db.sql`. |
| `006` (старый `007`, user_id) | `DROP INDEX IF EXISTS expenses_user_id_idx` + `DROP COLUMN IF EXISTS user_id` | Откат теряет привязку трат к пользователю. |
| `007` (старый `004`, user_settings) | `DROP TABLE IF EXISTS user_settings` | Триггер уходит с таблицей. |
| `008` (vue-torrent-bot) | `DELETE FROM applications WHERE name = 'vue-torrent-bot'` | Строка возвращается, если нужна. |

Два файла требуют `StatementBegin/StatementEnd`. Это не опционально:

- **`005` (бывший `006`)**: `CREATE OR REPLACE FUNCTION update_updated_at_column()`
  — тело на PL/pgSQL, внутри `$$` есть `;`.
- **`006` (бывший `007`)**: блок `DO $$ BEGIN ... END $$;` с `RAISE EXCEPTION`.

Оберни каждый такой блок целиком:

```
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION ... $$ ... $$ LANGUAGE plpgsql;
-- +goose StatementEnd
```

Обрати внимание: `CREATE TRIGGER ... EXECUTE FUNCTION ...;` в `005`/`007`
оборачивать **не нужно** — там нет `$$` и внутренних `;`.

Файлы `003`, `004` (бывшие `002`, `003`) содержат только `ALTER TABLE` и `INSERT`
— `StatementBegin` им не нужен.

**Проверка без запуска на рабочей БД.** Подними одноразовую базу и прогони goose
по ней после шага 5. Если хочешь проверить раньше — установи CLI
(`go install github.com/pressly/goose/v3/cmd/goose@latest`) и запусти
`goose -dir apps/go-api/migrations postgres "$DATABASE_URL" validate`.
Валидатор ловит отсутствие `Up`, дубли аннотаций и незакрытые `StatementBegin`.

## Шаг 5. Переписать `cmd/migrate/main.go`

Скелет (собери файл сам, не копируй целиком — смысл в том, чтобы руками пройти
по каждой части):

**Импорты.** `context`, `database/sql`, `flag`, `log`, `time`; пакет миграций
`flatty-budget/go-api/migrations`; `github.com/pressly/goose/v3`; и **обязательный
blank-import** `_ "github.com/jackc/pgx/v5/stdlib"` — он регистрирует драйвер
`pgx` для `database/sql`.

**Подключение.** Вместо `pgxpool.New` — `sql.Open("pgx", cfg.DatabaseUrl)`.
`pgxpool.Pool` — это высокоуровневый пул pgx, а не `*sql.DB`; goose работает
только с `*sql.DB`, потому что ему нужны собственные транзакции на миграцию.
Драйвер `stdlib` — официальный адаптер, отдельная зависимость не нужна:
`pgx/v5` уже в `go.mod`.

Не забудь `defer db.Close()` и `db.PingContext(ctx)` с коротким таймаутом:
`sql.Open` не открывает соединение, ошибка всплывёт на первом запросе.

**Настройка goose** (глобальное состояние, вызывается один раз до любых операций):

- `goose.SetBaseFS(migrations.FS)`
- `goose.SetDialect("postgres")` — вернёт ошибку, если диалект неизвестен,
  проверь её.

**Команды.** Минимально достаточно `goose.UpContext(ctx, db, ".")`. Полезно
добавить два режима через `flag`:

- `-status` → `goose.StatusContext(ctx, db, ".")` — таблица «применено/нет»;
- `-down` → `goose.DownContext(ctx, db, ".")` — откат **одной** последней версии.

И в конце — печать текущей версии: `goose.GetDBVersion(db)` вернёт `int64`
(0 = ничего не применено). Это твой главный индикатор состояния.

**Почему не оставить `pgxpool` для раннера.** Можно, но тогда миграции идут не в
тех транзакциях, которых ждёт goose, и `goose.Up` просто не примет аргумент.
Держи два разных инструмента для двух разных задач: пул — для HTTP-сервера,
`*sql.DB` — для миграций. Это нормально.

**Проверка компиляции:**

```bash
cd apps/go-api
go build ./cmd/migrate
go vet ./cmd/migrate
```

Отдельная проблема: в репозитории **уже закоммичен** бинарь `apps/go-api/migrate`
(12.5 МБ, результат старого `go build ./cmd/migrate`). Удали его из индекса и с
диска, иначе следующие коммиты будут тащить мусор:

```bash
git rm --cached apps/go-api/migrate
rm apps/go-api/migrate
printf 'migrate\n' >> apps/go-api/.gitignore
```

Проверь, что `git status` больше не показывает бинарь.

## Шаг 6. Привести порядок версий в соответствие со смыслом

goose накатывает **строго по возрастанию числа в имени**. Значит после перехода
порядок будет `001 → 002 → … → 008`, а не ручной
`005 → 002 → 003 → 006 → 007 → 004 → 008`.

Разберёмся, ломается ли что-то на самом деле:

- `db.sql` уже содержит финальную схему статистики, поэтому `003`/`004`
  (бывшие backfill'ы `002`/`003`) отработают и без предварительной `002`
  (бывшего `005`).
- Но `002` (rescope) делает `TRUNCATE` обеих таблиц и пересобирает их заново —
  то есть **сотрёт результат** `003`/`004`, включая `created_at`, который они
  старательно вычисляли. Работа сделана и выброшена.
- Кроме того, старый комментарий «005 must run before 002/003» перестал быть
  правдой ровно в тот момент, когда `db.sql` обзавёлся финальной схемой. Он
  остался в коде как исторический артефакт.

Поэтому **перенумеровываем файлы**: номер версии — это контракт о порядке,
а не порядковый номер по дате создания. Делай это до первого запуска goose —
тогда таблицы версий ещё нет и переименование ничего не ломает.

```bash
cd apps/go-api/migrations
mkdir .renumber
mv *.sql .renumber/

mv .renumber/001_*.sql 001_add_description_to_expenses.sql
mv .renumber/005_*.sql 002_add_resident_location_to_expense_stats.sql
mv .renumber/002_*.sql 003_backfill_monthly_totals.sql
mv .renumber/003_*.sql 004_backfill_monthly_averages.sql
mv .renumber/006_*.sql 005_add_applications.sql
mv .renumber/007_*.sql 006_add_user_id_to_expenses.sql
mv .renumber/004_*.sql 007_add_user_settings.sql
mv .renumber/008_*.sql 008_add_vue_torrent_bot_application.sql

rmdir .renumber
ls
```

Проверь: `ls` показывает `embed.go` и ровно `001…008` в том порядке, в котором
миграции должны выполняться. После `git add -A apps/go-api/migrations` команда
`git status` покажет переименования (`renamed:`), а не удаления + создания —
содержимое файлов не менялось, менялись только имена.

Затем убери исторические комментарии, которые теперь врут:

- шапка `002_add_resident_location_to_expense_stats.sql` («Previously both tables
  were keyed only by (month, year)…») — оставь, это объясняет *зачем*, но
  добавь строку, что порядок теперь задан номером;
- комментарий в `cmd/migrate/main.go` про `005 must run before 002/003` — удали
  вместе со slice;
- в `db.sql` около строки 552 есть комментарий
  `-- (equivalent to migrations/002 + migrations/003 + migrations/005)` —
  после перенумерации он неверен, обнови номера (`002 + 003 + 004`).

**Альтернатива, если перенумерация кажется шумом:** слить `002`+`003`+`004`
в одну миграцию `002_expense_stats.sql` (ALTER + один INSERT для totals,
один для averages) и удалить три файла. Результат тот же, истории меньше.
Для портфолио перенумерация честнее: видно, как схема развивалась.

## Шаг 7. Разобраться с уже применённой dev-базой

Здесь два пути. Выбери осознанно, не «на удачу».

### Путь A (рекомендуется) — пересоздать dev-базу

Данные в dev — сид из `db.sql`, терять нечего.

```bash
# из корня репозитория
podman-compose -f docker-compose.yml --env-file .env.dev down
podman volume ls | grep -i postgres     # найди volume с данными postgres
podman volume rm <имя-volume>

podman-compose -f docker-compose.yml --env-file .env.dev up -d postgres init-db
```

Затем (после того как nest-graphql создаст таблицу `"user"` — при работе на
хосте это `npm run dev` или `nx serve nest-graphql`):

```bash
cd apps/go-api && go run ./cmd/migrate
```

> `nx migrate-db go-api` из шага 9 делает то же самое — но таргета пока нет,
> так что на этом шаге запускай бинарь напрямую.

Плюс этого пути: ты видишь «чистый» прогон всех восьми версий и знаешь, что
`goose_db_version` отражает реальность.

Полезное свойство goose: каждая миграция идёт в своей транзакции. Если `007`
(`user_settings`) упадёт из-за отсутствующей таблицы `"user"`, версии `001…006`
останутся применёнными, и повторный запуск продолжит с `007`, а не с нуля.
Именно поэтому порядок «сначала nest-graphql (создаёт `"user"`), потом мигратор»
важнее, чем кажется.

### Путь B — сохранить данные и зафиксировать baseline

Если в базе есть то, что жалко (ручные траты, настройки), не пересоздавай её.
Нужно вручную сказать goose: «всё до версии 8 уже применено».

1. Убедись, что все восемь миграций действительно применены. Проверь по схеме:
   `\d expenses` (есть `description`, `user_id`, индекс `expenses_user_id_idx`),
   `\d applications`, `\d user_settings`, ключ `expense_monthly_totals`.
   Если чего-то нет — сначала накати это старым способом, потом baseline.
2. Создай таблицу версий. Проще всего один раз запустить мигратор с флагом
   baseline, который делает `goose.EnsureDBVersion(db)` (создаёт
   `goose_db_version` с версией 0) и **не** вызывает `Up`.
   Если такого флага нет — создай таблицу руками тем же DDL, что печатает goose:

   ```
   CREATE TABLE goose_db_version (
       id serial NOT NULL,
       version_id bigint NOT NULL,
       is_applied boolean NOT NULL,
       tstamp timestamp NULL DEFAULT now(),
       PRIMARY KEY (id)
   );
   ```

3. Отметь все версии применёнными:

   ```bash
   psql "$DATABASE_URL" -c "INSERT INTO goose_db_version (version_id, is_applied) VALUES (1,true),(2,true),(3,true),(4,true),(5,true),(6,true),(7,true),(8,true);"
   psql "$DATABASE_URL" -c "SELECT max(version_id) FROM goose_db_version WHERE is_applied;"
   ```

   Ожидаемо: `8`.

4. Запусти `go run ./cmd/migrate -status` — все восемь должны быть `Applied`.

**Почему не «просто запустить up и посмотреть».** На этой базе `007`
(`CREATE TABLE user_settings` без `IF NOT EXISTS`) упадёт с «relation already
exists», а `002` устроит `TRUNCATE` живым данным. Идемпотентность остальных
файлов тут не спасает — она не про версии, а про повторное выполнение SQL.

## Шаг 8. Убрать второй механизм наката

Пока `init-db.sh` и `init-db-auth.sh` гоняют те же файлы через `psql`, у схемы
два владельца, и `goose_db_version` будет врать: таблица скажет «не применено»,
а колонка уже есть.

**`scripts/init-db.sh`:** удали блок `Applying application registry migrations...`
(цикл по `/migrations/005…` и `/migrations/008…`). Оставь:
- ожидание Postgres,
- `psql -f /db.sql`,
- создание базы `torrent_bot`.

Обнови комментарий: `db.sql` — базовая схема, goose — всё остальное.

**`scripts/init-db-auth.sh`:** удали файл. Его работа (миграции `006` и `007`,
зависящие от таблицы `"user"`) теперь у goose.

**`docker-compose.yml`:**
- удали сервис `init-db-auth` целиком;
- у сервиса `init-db` удали монтирование `./apps/go-api/migrations:/migrations:ro`
  — он больше не читает миграции;
- добавь сервис `migrate`, который гоняет тот же бинарь, что и go-api:
  - `build` с тем же `apps/go-api/Dockerfile`;
  - `command: ["/migrate"]` (или как назовёшь бинарь);
  - **те же `environment`, что у `go-api`** — в первую очередь `DATABASE_URL`
    с хостом `postgres` (не `localhost`) и `sslmode=disable`;
  - `depends_on`: `init-db` → `condition: service_completed_successfully`,
    `nest-graphql` → `condition: service_healthy` (иначе `"user"` ещё нет);
  - `restart: "no"`;
- у `go-api` замени зависимость `init-db: service_completed_successfully` на
  `migrate: service_completed_successfully`.

**`docker-compose.backend.yml`:** то же самое, но с `network_mode: host` и
`DATABASE_URL` на `localhost:55000`.

**`apps/go-api/Dockerfile`:** собери второй бинарь в builder-стадии
(`CGO_ENABLED=0 go build -o migrate ./cmd/migrate/`) и скопируй его в runtime-образ
рядом с `api`. В builder уже есть `go test ./...` — он останется зелёным.

Проверка:

```bash
grep -rn "migrations/" scripts/ docker-compose.yml   # пусто
podman-compose -f docker-compose.yml --env-file .env.dev build migrate
podman-compose -f docker-compose.yml --env-file .env.dev up -d migrate
podman-compose -f docker-compose.yml --env-file .env.dev logs migrate
```

В логах — `OK 001_…` … `OK 008_…`, контейнер завершился с кодом 0.

**Минимальный вариант, если не хочешь трогать compose.** Сделай только шаг с
`init-db.sh` (убери цикл) и **не запускай** `init-db-auth` — работай на хосте
(`nx migrate-db go-api`). Тогда второй механизм фактически не участвует, но файл
и сервис остаются в репозитории как мина. Для портфолио лучше убрать.

## Шаг 9. Добавить Nx-таргет

В `apps/go-api/project.json` добавь таргет по образцу `build`/`test`:

- имя: **`migrate-db`** (не `migrate`!);
- `executor: nx:run-commands`;
- `options.command: "go run ./cmd/migrate"`;
- `options.cwd: "apps/go-api"`.

**Почему не `migrate`.** `nx migrate` — встроенная команда Nx для обновления
пакетов (`nx migrate @nx/workspace@latest`). Целевой таргет с таким именем
перекрывается ею: `nx migrate go-api` уйдёт в апдейтер и попытается «мигрировать
пакет go-api». Ошибка будет неочевидной. Имя `migrate-db` такой коллизии не даёт.

Проверка:

```bash
nx migrate-db go-api
nx show project go-api | grep -A4 migrate-db
```

Ожидаемо: первый запуск — `OK 001_…` … `OK 008_…`; повторный — пустой вывод
(ничего не применено, `goose` видит версию 8). Это и есть проверка
версионирования: **второй запуск обязан быть no-op**.

## Шаг 10. Обновить документацию

Три места, иначе следующий человек снова пойдёт гонять `psql` руками.

1. **`apps/go-api/AGENTS.md`**, таблица команд: добавь строку
   `| nx migrate-db go-api | Apply DB migrations (goose) |`.
2. **`apps/go-api/AGENTS.md`**, раздел про БД (добавь, если нет): «Базовая схема —
   `db.sql` (накатывается сервисом `init-db`). Всё остальное — goose-миграции в
   `apps/go-api/migrations/`, встроенные в бинарь. Запуск: `nx migrate-db go-api`.
   Повторный запуск безопасен».
3. **`README.md`**, Quick Start: после `podman-compose … up -d postgres init-db …`
   при работе на хосте добавь шаг `nx migrate-db go-api` с пояснением, что
   миграции, зависящие от таблицы `"user"`, требуют запущенного `nest-graphql`
   (или сервиса `migrate` в полном compose-стеке).

---

## Критерии готовности

- [ ] `apps/go-api/migrations/embed.go` существует, `go build ./migrations`
      проходит.
- [ ] Во всех восьми `.sql` есть `-- +goose Up`; в файлах с PL/pgSQL-блоками
      (`005`, `006` после перенумерации) блоки обёрнуты в
      `StatementBegin`/`StatementEnd`.
- [ ] `cmd/migrate/main.go` не содержит slice путей и `os.ReadFile`; использует
      `sql.Open("pgx", …)` + `goose.SetBaseFS` + `goose.UpContext`.
- [ ] Порядок файлов `001…008` совпадает с реальным порядком зависимостей
      (rescope до backfill'ов), старые комментарии про `005` обновлены.
- [ ] На чистой БД `nx migrate-db go-api` накатывает всё и завершается с кодом 0;
      **повторный** запуск — no-op.
- [ ] `go run ./cmd/migrate -status` показывает восемь `Applied` и версию `8`.
- [ ] `go run ./cmd/migrate -down` откатывает последнюю версию, повторный
      `up` возвращает её. Проверено на чистой БД.
- [ ] `scripts/init-db.sh` больше не гоняет миграции; `init-db-auth` удалён
      (или явно задокументировано, почему оставлен).
- [ ] `nx migrate-db go-api` работает; таргет `migrate-db` есть в
      `apps/go-api/project.json`.
- [ ] `apps/go-api/AGENTS.md` и корневой `README.md` описывают запуск миграций.
- [ ] `git status` чистый после прогона — в репозиторий не попал бинарь
      `apps/go-api/migrate`.

## Подсказки и грабли

- **`//go:embed` и `..`.** Встроить `../migrations/*.sql` нельзя. Либо пакет
  живёт рядом с `.sql` (наш вариант), либо файлы переезжают в подкаталог пакета.
- **Аргумент `dir` должен совпадать с layout embed.** `*.sql` → `"."`,
  `migrations/*.sql` → `"migrations"`. Если goose говорит «no migrations found» —
  почти всегда это.
- **`-- +goose Up` ровно один и первым.** Два `Up` или `Up` после `Down` — ошибка
  парсинга, и она проявляется только при запуске, не при компиляции.
- **`$$` ломает наивный сплит по `;`.** Любая PL/pgSQL-функция и любой `DO $$ … $$`
  требуют `StatementBegin/StatementEnd`. Проверяй `goose validate` или прогоном.
- **goose — глобальное состояние.** `SetBaseFS` и `SetDialect` влияют на весь
  процесс. Вызывай их один раз на старте, до любых `Up`/`Status`/`Down`.
- **`pgxpool.Pool` ≠ `*sql.DB`.** Не пытайся передать пул в goose. Драйвер
  `pgx/v5/stdlib` + `sql.Open` — правильный мост, и он уже в зависимостях.
- **`nx migrate` занято.** Целевой таргет называй `migrate-db`, иначе встроенная
  команда Nx перехватит вызов.
- **Идемпотентность ≠ версионирование.** `IF NOT EXISTS` спасает от повторного
  запуска, но не отменяет необходимость таблицы версий: `TRUNCATE`, `DELETE`,
  `UPDATE` и `ALTER … DROP` идемпотентными не бывают.
- **Baseline — это ручное утверждение.** `INSERT INTO goose_db_version` говорит
  goose «поверь мне». Если соврёшь (отметил непримененную версию), схема и история
  разъедутся молча. Сверяйся с `\d` перед вставкой.
- **`Down` — не «отменить накат», а «вернуть схему».** Если данные потеряны
  безвозвратно, честнее оставить `Down` пустым с комментарием, чем писать
  SQL, который делает вид, что откат возможен.
- **Транзакции.** По умолчанию goose выполняет каждый файл в транзакции.
  `CREATE INDEX CONCURRENTLY` и `CREATE DATABASE` в транзакции не работают —
  для них нужен `-- +goose NO TRANSACTION`. В текущих восьми файлах такого нет,
  но помни на будущее.
- **Не удаляй `update_updated_at_column()` в down'ах.** Функция создана в
  `db.sql` и используется триггерами `resident_locations` и таблиц статистики.
  `CREATE OR REPLACE` в `005` её просто переопределяет.
- **psql гонял файл целиком — вместе с `Down`.** Старый цикл в `init-db.sh`
  скармливал `psql` весь `008_add_vue_torrent_bot_application.sql`, включая
  `-- +goose Down` с `DROP TABLE applications`. Поэтому `005` «применялась»,
  `applications` тут же удалялась, и `008` падала с
  `relation "applications" does not exist`. Любой инструмент, который не понимает
  аннотации goose, выполняет `Down` как обычный SQL — это и есть причина, почему
  у миграций должен быть ровно один владелец.
- **`migrate` не может стартовать раньше `nest-graphql`.** `006` и `007`
  ссылаются на таблицу `"user"`, которую создаёт better-auth в nest-graphql.
  `depends_on: init-db` недостаточно — нужен `nest-graphql: service_healthy`,
  иначе миграции падают на `relation "user" does not exist`. На чистой базе это
  выглядит как «goose сломан», хотя схема просто ещё не доросла.
- **`db.sql` теперь «версия 0».** Если поменяешь базовую схему, помни: goose
  увидит миграции, рассчитанные на старую базу. Изменения базовой схемы должны
  либо оставаться совместимыми, либо оформляться новой миграцией, а не правкой
  `db.sql` задним числом.

## Что почитать

- `goose` README, раздел «Embedded sql migrations» — ровно наш сценарий.
- `goose` README, раздел «SQL Migrations» — правила аннотаций и
  `StatementBegin/End`.
- `pressly/goose` → `provider_types.go` / `migrate.go` — если хочется понять,
  как таблица версий используется при `Up`.
- `database/sql` docs про `sql.Open` и пул соединений — почему `Open` не
  подключается, и почему `Ping` обязателен для ранней диагностики.
