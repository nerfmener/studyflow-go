# StudyFlow

Небольшой трекер обучения на Go. Он помогает не просто собирать список курсов, а фиксировать реальные учебные сессии и видеть свой прогресс.

## Возможности

- добавление учебных задач с категорией;
- запись времени и заметки после занятия;
- общий прогресс и streak по дням;
- простое JSON-хранилище без внешней базы данных;
- HTTP API и минимальный веб-интерфейс.

Проект намеренно сделан простым: это учебный backend-проект, который легко читать и постепенно расширять.

## Запуск

Нужен Go 1.23 или новее.

```bash
go test ./...
go run ./cmd/studyflow
```

Открой http://localhost:8080.

По умолчанию данные сохраняются в `data/studyflow.json`. Можно указать другой путь:

```bash
STUDYFLOW_DATA=/tmp/studyflow.json go run ./cmd/studyflow
```

## API

```bash
curl http://localhost:8080/health
curl -X POST http://localhost:8080/api/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"Изучить HTTP в Go","category":"Go"}'
curl -X POST http://localhost:8080/api/sessions \
  -H 'Content-Type: application/json' \
  -d '{"minutes":45,"note":"Написал первый handler"}'
```

## Идеи для следующих шагов

- добавить отметку задачи как выполненной;
- заменить JSON на SQLite;
- добавить авторизацию;
- покрыть HTTP-слой интеграционными тестами;
- сделать Dockerfile и CI.

## Лицензия

MIT
