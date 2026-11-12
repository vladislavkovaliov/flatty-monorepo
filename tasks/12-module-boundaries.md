# Задача 12. Контроль границ модулей (enforce-module-boundaries)

> Тема: **Nx depConstraints + @nx/enforce-module-boundaries** — запретить
> app→app импорты и зафиксировать, кто на кого может ссылаться.
> Результат: нелегальный импорт между приложениями падает на этапе lint, а не
> обнаруживается случайно на ревью.

---

## Зачем

Это **принципиальный пробел для архитектора**. В монолите это просто «плохой стиль»;
в монолито-репозитории с микрофронтендами это вопрос контроля архитектуры.

Сейчас ничто не мешает, например, `react-resident` импортировать компонент напрямую
из `react-entrypoint` (или наоборот). Такие импорты:

- создают **неявную связность** между приложениями, которые задуманы независимыми;
- ломают основную идею микрофронтендов — независимую сборку и деплой;
- приводят к «большому клубку», где изменение в одном приложении молча тянет другое.

На собеседовании вопрос «как ты защищаешь границы модулей в монорепе?» — один из
маркеров solution-архитектора. Правильный ответ: теги на проектах + `depConstraints`
в `nx.json` + линт-правило, которое это **автоматически** проверяет. Сейчас в проекте
нет ни `depConstraints`, ни правила `@nx/enforce-module-boundaries` (проверено —
в `nx.json` их нет).

## Контекст проекта

- `@nx/eslint` уже стоит (`^23.0.1` в `package.json`) — правило доступно, осталось
  его подключить.
- **Единого корневого ESLint-конфига нет.** Конфиги лежат по приложениям:
  - `apps/react-launcher/eslint.config.js`
  - `apps/react-settings/eslint.config.js`
  - `apps/react-resident/eslint.config.js`
  - `apps/vue-torrent-bot/eslint.config.ts`
- Теги есть **только** у `apps/go-api/project.json`:
  `"tags": ["scope:backend", "type:api"]`. Остальные приложения без тегов.
- В `nx.json` есть `targetDefaults` и `plugins` (`@nx-go/nx-go`, `@nx/next/plugin`),
  но ключа `depConstraints` нет.

---

## Шаг 1. Придумать схему тегов

Минимально достаточная схема — два измерения: **scope** (домен) и **type** (роль):

| Тег | Значение |
|-----|----------|
| `scope:backend` | go-api, nest-graphql |
| `scope:frontend` | все React/Vue приложения |
| `scope:shell` | react-launcher, react-entrypoint (хосты) |
| `scope:remote` | react-settings, react-resident, vue-torrent-bot |
| `scope:agent` | python-agent, python-torrent-bot |
| `scope:shared` | packages/sdk, packages/mfe-host |
| `type:app` | все приложения |
| `type:api` | go-api, nest-graphql |
| `type:lib` | пакеты в packages/ |

Не усложняй. Начни с `type:app`/`type:lib` и `scope:...`, добавляй теги по мере
реальной необходимости (YAGNI).

## Шаг 2. Проставить теги в project.json

У каждого приложения в `apps/<name>/project.json` и у каждого пакета в
`packages/<name>/project.json` добавь поле `"tags": [...]`. Образец уже есть —
посмотри `apps/go-api/project.json`:

- `go-api` → `["scope:backend", "type:api"]`
- `nest-graphql` → `["scope:backend", "type:api"]`
- `react-launcher` → `["scope:shell", "type:app"]`
- `react-entrypoint` → `["scope:shell", "type:app"]`
- `react-settings`, `react-resident` → `["scope:remote", "type:app"]`
- `vue-torrent-bot` → `["scope:remote", "type:app"]`
- `python-agent`, `python-torrent-bot` → `["scope:agent", "type:app"]`
- `packages/sdk`, `packages/mfe-host` → `["scope:shared", "type:lib"]`

Проверь, что теги не противоречат уже существующим (у go-api теги уже есть — не
затирай их, а дополни при необходимости).

## Шаг 3. Добавить depConstraints в nx.json

В корень `nx.json` (на уровень `targetDefaults`, не внутрь него) добавь ключ
`depConstraints`. Суть: описать, каким `sourceTag` разрешено ссылаться на какие
`onlyDependOnLibsWithTags`.

Ключевые правила, которые надо зафиксировать:

- **app → app запрещён.** Приложение может зависеть только от `type:lib` (пакетов)
  и своих зависимостей, но **не** от другого `type:app`.
- **shell → remote запрещён напрямую** (иначе потеряется смысл runtime-инъекции
  скриптов). Хосты общаются с ремоутами через `packages/mfe-host`/`packages/sdk`
  и через рантайм-загрузку, а не через импорт кода.
- **remote → shell запрещён** (ремоуты — изолированные приложения).
- **shared → app запрещён** (общие пакеты не должны знать про приложения).

Если какое-то приложение сейчас легально импортирует пакет из `packages/`, добавь
соответствующее разрешение (например, всё `scope:frontend` может зависеть от
`scope:shared`).

## Шаг 4. Подключить правило @nx/enforce-module-boundaries

Правило читает `depConstraints` и `tags` автоматически. Его надо добавить в каждый
ESLint-конфиг приложения (корневого конфига нет):

- `apps/react-launcher/eslint.config.js`
- `apps/react-settings/eslint.config.js`
- `apps/react-resident/eslint.config.js`
- `apps/vue-torrent-bot/eslint.config.ts`

В каждом конфиге в секцию `rules` добавь `@nx/enforce-module-boundaries`
(с включением или «error»). Если хочешь не дублировать — вынеси общую часть в
shared-конфиг, но начни с прямого добавления в каждый файл (проще и нагляднее).

> **Грабли с Next.js.** `react-entrypoint` — Next.js приложение; его lint-таргет
> генерируется плагином `@nx/next/plugin` (`lintTargetName: "lint"` в `nx.json`).
> Убедись, что правило попало в конфиг, который Next-линтер реально использует.
> Остальные React-приложения (react-launcher/settings/resident) — без явного
> `"lint"` таргета в project.json (проверь, что их ESLint вообще запускается через
> `nx lint`), иначе правило будет висеть мёртвым.

## Шаг 5. Прогнать и починить существующие нарушения

```bash
npx nx run-many -t lint --all
```

Скорее всего, после включения правила всплывут существующие нелегальные импорты
(например, если какое-то приложение уже тянет код из другого). Разбери каждое:

- если это **легальная зависимость от пакета** → добавь тег/разрешение в
  `depConstraints`, а не глуши правило;
- если это **реальный app→app импорт** → вынеси общий код в `packages/` (правильное
  решение) или оформи как публичный API пакета.

## Критерии готовности

- [ ] У всех приложений и пакетов проставлены `tags` в project.json.
- [ ] В `nx.json` есть `depConstraints`, запрещающие app→app.
- [ ] Правило `@nx/enforce-module-boundaries` подключено во все frontend ESLint-конфиги.
- [ ] `npx nx run-many -t lint --all` проходит чисто.
- [ ] Попытка добавить app→app импорт падает на lint (проверь вручную: временно
      добавь импорт, убедись что линтер ругается, откати).

## Подсказки и грабли

- **Не глуши правило через `// eslint-disable`.** Каждый disable — это дыра в
  контроле границ. Лучше исправить архитектурно (вынести в пакет).
- **Теги — это контракт.** Меняй их осознанно: переименование тега без правки
  `depConstraints` молча снимает защиту.
- **Начинай с минимума правил.** Одно правило «app не зависит от app» даёт 80%
  пользы. Детальная иерархия shell/remote — поверх, когда разберёшься с базой.
- **Проверяй не только React.** `vue-torrent-bot` — отдельный конфиг
  (`eslint.config.ts`), его легко забыть.
- **График зависимостей.** `nx graph` покажет, кто на кого ссылается, — удобно
  сверяться до и после.
