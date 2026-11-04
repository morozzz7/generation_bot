# Generation Bot

Telegram-бот для генерации изображений через Stable Diffusion. Go-бот общается с Telegram, Python/FastAPI-сервис держит модель на GPU.

## Возможности

- Генерация изображения по текстовому промпту.
- Выбор модели через inline-кнопки: DreamShaper 8 или SDXL Turbo.
- Отмена активной генерации командой `/cancel`.
- Отмена ожидания промпта до запуска генерации.
- Генерации выполняются последовательно; в памяти держится только одна модель.

## Архитектура

```text
Telegram <-> Go bot (internal/) <-> FastAPI (api/) <-> Stable Diffusion (GPU)
```

## Модели


| `model_id`    | Модель                                                                  | База                 | Шаги | Guidance | Dtype              |
| ------------- | ----------------------------------------------------------------------- | -------------------- | ---- | -------- | ------------------ |
| `dreamshaper` | [Lykon/dreamshaper-8](https://huggingface.co/Lykon/dreamshaper-8)       | Stable Diffusion 1.5 | 25   | 7.0      | FP32               |
| `sdxl_turbo`  | [stabilityai/sdxl-turbo](https://huggingface.co/stabilityai/sdxl-turbo) | SDXL Turbo           | 4    | 0.0      | FP16 + CPU offload |


Модель загружается лениво при первом запросе. При переключении на другую модель текущий pipeline выгружается из памяти, после чего загружается новый.

## Структура проекта

```text
generation_bot/
├── cmd/bot/main.go            # точка входа Telegram-бота
├── internal/
│   ├── bot/
│   │   ├── handlers.go        # сообщения, выбор модели, запуск генерации
│   │   ├── commands.go        # /start, /help, /generate, /about, /cancel
│   │   └── state.go           # состояния пользователя, task ID, выбранная модель
│   └── generation/clients.go  # HTTP-клиент FastAPI
├── api/
│   ├── main.py                # FastAPI: /health, /generate, /cancel
│   ├── schemas.py             # валидация запроса и model_id
│   ├── task_manager.py        # реестр задач и Event для отмены
│   └── model/
│       ├── loader.py          # загрузка DreamShaper 8 / SDXL Turbo
│       └── inference.py       # инференс, переключение моделей, отмена
├── go.mod
├── api/requirements.txt
└── .env                       # не коммитится
```



## Требования

- Go 1.27.3+
- Python 3.10+
- NVIDIA GPU с CUDA (на CPU работает, но медленно)
- Интернет для первой загрузки моделей с Hugging Face



## Настройка

```bash
python -m venv myvenv
source myvenv/Scripts/activate   # Windows Git Bash
pip install -r api/requirements.txt
go mod tidy
```

Файл `.env` в корне проекта:

```env
TELEGRAM_BOT_TOKEN=your_telegram_bot_token
API_BASE_URL=http://127.0.0.1:8000
```



## Запуск

API (терминал 1):

```bash
uvicorn api.main:app --host 127.0.0.1 --port 8000
```

Бот (терминал 2):

```bash
go run ./cmd/bot
```

Swagger UI: [http://127.0.0.1:8000/docs](http://127.0.0.1:8000/docs)

## Команды


| Команда     | Действие                                         |
| ----------- | ------------------------------------------------ |
| `/start`    | Приветствие                                      |
| `/help`     | Список команд                                    |
| `/generate` | Выбрать модель и запросить промпт                |
| `/about`    | Информация о боте и моделях                      |
| `/cancel`   | Отменить ожидание промпта или активную генерацию |


Сценарий: `/generate` → кнопка модели → текстовый промпт → изображение.

## API


| Метод | Endpoint            | Описание                |
| ----- | ------------------- | ----------------------- |
| GET   | `/health`           | Статус сервиса          |
| POST  | `/generate`         | Генерация, ответ — PNG  |
| POST  | `/cancel/{task_id}` | Запросить отмену задачи |


Тело `POST /generate`:

```json
{
  "task_id": "550e8400-e29b-41d4-a716-446655440000",
  "model_id": "dreamshaper",
  "prompt": "futuristic city at sunset",
  "negative_prompt": "",
  "steps": 25,
  "seed": -1
}
```

`model_id`: `dreamshaper` (по умолчанию) или `sdxl_turbo`. Для SDXL Turbo шаги и guidance задаются сервером.

Ответ — `image/png` с заголовками `X-Generation-Task-ID`, `X-Generation-Seed`, `X-Generation-Model`.

Коды ошибок `/generate`: `409` — задача уже существует или отменена, `503` — не хватило VRAM, `500` — прочая ошибка. `/cancel` возвращает `404`, если задача не найдена или уже завершена.

## Как работает отмена

1. Бот создаёт UUID задачи и передаёт его в `/generate` вместе с `model_id`.
2. `/cancel` вызывает `/cancel/{task_id}`.
3. `TaskManager` устанавливает `threading.Event`.
4. Инференс проверяет событие перед стартом и после каждого шага, после чего прерывается.



## Известные ограничения

- Первый запрос к каждой модели медленнее из-за загрузки весов и возможной скачки с Hugging Face.

