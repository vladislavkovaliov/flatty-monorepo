# Задача 14. Дрейф документации: react-wrapper и dead-код

> Тема: **актуализировать доки** — убрать ссылки на несуществующий `react-wrapper`
> и устаревшие описания dead-кода.
> Результат: README/AGENTS/документы описывают реальное состояние репозитория.

---

## Зачем

**Устаревшая документация хуже, чем её отсутствие.** Человек (или LLM) читает README,
верит архитектурной диаграмме — а приложение, на которое всё ссылается, уже удалено.
Это подрывает доверие ко всей документации: если одна часть врёт, читатель начинает
сомневаться во всём остальном.

На собеседовании это прямой удар: ты показываешь репозиторий, а интервьюер видит
диаграмму с `react-wrapper` (порт 5174, «full shell with nav»), которого нет в
`apps/`. Мгновенный вывод — «автор не следит за собственным репо».

---

## Уже сделано

- **`README.md`** — `react-wrapper` вычищен из ASCII-диаграммы и из таблицы
  «Project Structure»; диаграмма осталась цельной (рамки и стрелки `Script injection`
  сходятся на `react-launcher` / `react-entrypoint`).
- **`AGENTS.md`** — `react-wrapper` убран из «Shared Frontend Conventions»;
  описание `dev:frontend` приведено к фактическому составу скрипта
  (`launcher + settings + resident + entrypoint + vue-torrent-bot`);
  shells в разделе «Micro-frontends» → `react-launcher`, `react-entrypoint`.
- **`dinamic-app.md`** — добавлен статус-блок «что реализовано / что намеренно не
  делалось»; §4 переписан в «`react-launcher` — no changes (by design)»; из Benefits,
  описания хука и Verification убраны утверждения, что launcher ещё предстоит мигрировать.
- **`apps/react-launcher/src/applications/*.ts` и ключ `localStorage`
  `react-applications`** — сознательно оставлены: launcher это демо-шелл с одним MFE,
  поход в `/api/applications` ему не нужен. Централизованный реестр
  (`GET/POST/PUT/DELETE /api/applications`, миграции `006`/`008`, htmx `/admin/applications`,
  `useApplications()` в `@flatty-budget/mfe-host`) использует только `react-entrypoint`.

**Проверка:** `grep -rn "react-wrapper" README.md AGENTS.md` → пусто.
Ссылки в `thoughts/shared/designs/`, `thoughts/shared/plans/` и старых
`thoughts/ledgers/` — исторический архив, их не трогаем. В `tasks/11-git-hygiene.md`
упоминание осталось как пример commit-message.

---

## Добито по ходу (остаточный дрейф)

Дрейф был не про `react-wrapper`, а про состав/порты приложений. Проверено по
`ls apps/` и конфигам dev-серверов:

- `README.md` — таблица «Project Structure»: `react-resident` `8082` → **8083**
  (в диаграмме было 8083, `webpack.apps.config.mjs` слушает 8083; `8082` — host-порт
  из `docker-compose.yml`, `["8082:80"]`).
- `AGENTS.md` — таблица Projects: добавлены `vue-torrent-bot` (Vite IIFE, 8084) и
  `python-agent` (Python + MCP, без порта, ссылка на `README.md` — `AGENTS.md` у него нет);
  `react-resident` → 8083; `react-entrypoint` — Stack `Nestjs` → **Next.js**.
- `AGENTS.md` — раздел Micro-frontends: добавлен Vue-remote `vue-torrent-bot`
  (IIFE, `window["ext-apps"]`, `IAppComponent` из `@flatty-budget/mfe-host`).

---

## Критерии готовности

- [x] `grep -rn "react-wrapper" README.md AGENTS.md` — пусто.
- [x] ASCII-диаграмма в README целостная (нет оборванных рамок/стрелок).
- [x] Таблица проектов в README соответствует реальному `apps/` по портам.
- [x] Таблица Projects в `AGENTS.md` покрывает все десять каталогов `apps/`.
- [x] `dinamic-app.md` помечен как реализованный.

## Подсказки и грабли

- **Не «подчищай» исторические design-документы под ноль.** Они фиксируют, как было
  принято решение. Достаточно актуализировать статус, а не переписывать текст.
- **Сверяйся с `apps/` и конфигами, а не с памятью.** Порты живут в
  `webpack.apps.config.mjs` / `vite.config.ts`, а не в доках.
- **ASCII-диаграммы ломаются легко.** После правок открой README в markdown-preview:
  кривая рамка заметна сразу.
