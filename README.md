# Generation Bot 🎨

Telegram-бот для генерации изображений на основе **Stable Diffusion**. Проект использует Go для взаимодействия с Telegram и Python + FastAPI для выполнения генерации изображений на GPU.

Пользователь отправляет команду `/generate`, вводит текстовый промпт и получает сгенерированное изображение непосредственно в Telegram.

## Содержание

- [Архитектура](#архитектура)
- [Технологический стек](#технологический-стек)
- [Модель генерации](#модель-генерации)
- [Структура проекта](#структура-проекта)
- [Требования](#требования)
- [Установка и настройка](#установка-и-настройка)
- [Запуск проекта](#запуск-проекта)
- [Команды Telegram](#команды-telegram)
- [Принцип работы генерации](#принцип-работы-генерации)
- [API](#api)
- [Возможные проблемы](#возможные-проблемы)
- [Планы развития](#планы-развития)

---

## Архитектура

Проект разделён на два независимых компонента:

**1. Go Telegram Bot**
- Получение и обработка сообщений пользователей.
- Обработка команд Telegram.
- Управление состояниями диалога.
- Отправка запросов к Python API.
- Отправка сгенерированных изображений пользователям.

**2. Python Image Generation API**
- Предоставление HTTP API через FastAPI.
- Загрузка модели Stable Diffusion.
- Обработка пользовательских промптов.
- Генерация изображений с использованием GPU.
- Возвращение изображений в формате PNG.

## Технологический стек

| Компонент | Технология |
|---|---|
| Основной язык бота | Go |
| Telegram API | go-telegram/bot |
| Конфигурация | godotenv |
| Backend API | Python |
| API Framework | FastAPI |
| ASGI Server | Uvicorn |
| ML Framework | PyTorch |
| Image Generation | Hugging Face Diffusers |
| Image Model | Stable Diffusion |
| Scheduler | DPM-Solver++ |
| Формат изображений | PNG |
| Протокол взаимодействия | HTTP / JSON |

## Модель генерации

В качестве генерационной модели используется:

**DreamShaper 8**

- Repository: https://huggingface.co/Lykon/dreamshaper-8
- Framework: Stable Diffusion 1.5
- Architecture: Latent Diffusion Model
- Library: Hugging Face Diffusers
- Weights format: SafeTensors

DreamShaper 8 — модель генерации изображений, основанная на Stable Diffusion 1.5 и дообученная для получения более выразительных и детализированных изображений.

### Параметры генерации

| Параметр | Значение |
|---|---|
| Model | DreamShaper 8 |
| Image width | 512 px |
| Image height | 512 px |
| Inference steps | 25 |
| Guidance scale | 7.0 |
| Scheduler | DPM-Solver++ |
| Data type | FP32 |
| Output format | PNG |
| Seed | Random / User-defined |

Для генерации используется GPU NVIDIA GTX 1080 Ti с 11 GB VRAM.

Модель загружается один раз при запуске FastAPI и остаётся в памяти до завершения работы сервиса.

**Лицензия:** CreativeML Open RAIL-M. Использование модели допускает локальный запуск, однако необходимо соблюдать условия лицензии и ограничения на использование сгенерированного контента.

## Структура проекта

```text
generation_bot/
│
├── cmd/
│   ├── bot/
│   │   └── main.go
│   │
│   └── test-api/
│       └── main.go
│
├── internal/
│   ├── bot/
│   │   ├── handler.go
│   │   ├── commands.go
│   │   └── state.go
│   │
│   └── generation/
│       └── client.go
│
├── api/
│   ├── main.py
│   ├── schemas.py
│   ├── requirements.txt
│   │
│   └── model/
│       ├── __init__.py
│       ├── loader.py
│       └── inference.py
│
├── .env
├── .gitignore
├── go.mod
└── README.md
```

## Требования

Перед запуском необходимо установить:

- Go 1.21 или новее.
- Python 3.10 или новее.
- NVIDIA GPU с поддержкой CUDA (либо запуск на CPU с существенно меньшей производительностью).
- NVIDIA GPU Driver.

## Установка и настройка

### 1. Клонирование проекта

```bash
git clone <repository-url>
cd generation_bot
```

### 2. Настройка Python-окружения

Создание виртуального окружения:

**Windows:**

```bash
python -m venv myvenv
```

Активация в PowerShell:

```powershell
.\myvenv\Scripts\Activate.ps1
```

Активация в Git Bash:

```bash
source myvenv/Scripts/activate
```

Установка зависимостей:

```bash
pip install -r api/requirements.txt
```

### 3. Проверка CUDA

Проверить доступность GPU через PyTorch:

```bash
python -c "import torch; print(torch.cuda.is_available()); print(torch.cuda.get_device_name(0) if torch.cuda.is_available() else 'CUDA unavailable')"
```

Ожидаемый результат:

```text
True
NVIDIA GeForce GTX 1080 Ti
```

### 4. Загрузка модели

При первом запуске Diffusers загружает веса модели DreamShaper 8 из Hugging Face.

Необходимо подключение к интернету и достаточное место на диске.

После загрузки модель сохраняется в локальном кэше Hugging Face, поэтому повторная загрузка обычно не требуется.

### 5. Установка Go-зависимостей

Из корневой директории проекта:

```bash
go mod tidy
```

### 6. Настройка переменных окружения

Создать файл `.env` в корне проекта:

```env
TELEGRAM_BOT_TOKEN=your_telegram_bot_token
API_BASE_URL=http://127.0.0.1:8000
```

`TELEGRAM_BOT_TOKEN` необходимо получить у [@BotFather](https://t.me/BotFather).

Не публикуйте `.env` и не добавляйте его в Git.

## Запуск проекта

Для работы приложения необходимо запустить два независимых процесса.

### Терминал №1 — Python FastAPI

Активировать виртуальное окружение и перейти в корневую директорию проекта.

Запустить API:

```bash
uvicorn api.main:app --host 127.0.0.1 --port 8000
```

При первом запуске будет выполнена загрузка модели.

После успешного запуска API доступен по адресу:

```text
http://127.0.0.1:8000
```

Проверка состояния сервиса:

```text
http://127.0.0.1:8000/health
```

Swagger UI:

```text
http://127.0.0.1:8000/docs
```

### Терминал №2 — Go Telegram Bot

Из корневой директории проекта:

```bash
go run ./cmd/bot
```

После запуска бот начнёт получать сообщения от Telegram.

### Сборка исполняемого файла

При необходимости можно собрать приложение:

```bash
go build -o generation_bot.exe ./cmd/bot
```

Запуск:

```bash
./generation_bot.exe
```

## Команды Telegram

| Команда | Описание |
|---|---|
| `/start` | Начало взаимодействия с ботом |
| `/generate` | Запуск процесса генерации изображения |
| `/cancel` | Отмена ожидания промпта |
| `/help` | Список доступных команд |
| `/about` | Информация о проекте |

### Пример использования

1. Пользователь отправляет:

```text
/start
```

2. Запускает генерацию:

```text
/generate
```

3. Бот запрашивает промпт:

```text
Send me a prompt for image generation.
```

4. Пользователь отправляет описание:

```text
A beautiful futuristic city at sunset,
cinematic lighting, highly detailed,
volumetric light, 4k
```

5. Бот сообщает о начале генерации, отправляет запрос в FastAPI и возвращает готовое изображение.

## Принцип работы генерации

### HTTP-запрос

Go-клиент отправляет POST-запрос:

```http
POST /generate
Content-Type: application/json
```

Пример тела запроса:

```json
{
    "prompt": "A futuristic city at sunset",
    "negative_prompt": "",
    "steps": 25,
    "seed": -1
}
```

### Обработка в Python

FastAPI получает запрос и передаёт параметры в `ImageGenerator`.

Генератор:

1. Проверяет корректность промпта.
2. Выбирает seed (случайный при значении `-1`).
3. Создаёт генератор случайных чисел PyTorch.
4. Запускает Stable Diffusion.
5. Получает изображение.
6. Возвращает изображение и seed.

Для защиты модели от одновременных вызовов используется `threading.Lock`.

Таким образом, запросы к одной загруженной модели выполняются последовательно.

### HTTP-ответ

FastAPI возвращает изображение непосредственно в теле HTTP-ответа:

```http
Content-Type: image/png
X-Generation-Seed: 123456
X-Generation-Time: 9.72
```

Go-клиент получает байты изображения и передаёт их Telegram API.

Промежуточное сохранение изображения на диск не требуется.

## API

| Endpoint | Метод | Назначение |
|---|---|---|
| `/health` | GET | Проверка состояния сервиса |
| `/generate` | POST | Генерация изображения |
| `/docs` | GET | Swagger UI |

### Проверка генерации через Swagger

Открыть:

```text
http://127.0.0.1:8000/docs
```

Выбрать `POST /generate`, нажать `Try it out` и передать JSON с промптом.

API вернёт PNG-файл.

## Возможные проблемы

### CUDA out of memory

Причина: недостаточно видеопамяти для выполнения генерации.

Возможные решения:
- Уменьшить разрешение изображения.
- Уменьшить количество дополнительных операций на GPU.
- Проверить, не занята ли видеопамять другими приложениями.

### FastAPI недоступен

Проверить:
- Запущен ли Uvicorn.
- Совпадает ли `API_BASE_URL` в `.env` с адресом FastAPI.
- Доступен ли endpoint `/health`.

### Telegram Bot не запускается

Проверить:
- Корректность `TELEGRAM_BOT_TOKEN`.
- Наличие `.env` в корне проекта.
- Установку Go-зависимостей.

### Ошибка загрузки модели

Проверить:
- Подключение к Hugging Face.
- Наличие свободного места на диске.
- Корректность зависимостей Diffusers и PyTorch.

## Планы развития

Текущая версия представляет собой базовую рабочую реализацию генерационного Telegram-бота.

Планируемые направления развития:

- [ ] PostgreSQL для хранения пользователей и истории генераций.
- [ ] Очередь задач генерации.
- [ ] Управление параллельностью и ограничение количества запросов.
- [ ] Сохранение истории сгенерированных изображений.
- [ ] Настройка параметров генерации пользователем.
- [ ] Поддержка нескольких генерационных моделей.
- [ ] Отображение статуса выполнения задач.
- [ ] Отмена уже запущенных задач.
- [ ] Docker-контейнеризация сервисов.
- [ ] Улучшение обработки ошибок и логирования.

---

**Project status:** Initial working implementation.

**Current functionality:** Telegram interaction, prompt-based image generation, FastAPI integration, GPU inference and PNG delivery.