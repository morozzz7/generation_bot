package bot

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	telegrambot "github.com/go-telegram/bot"

	"generation_bot/internal/generation"
)

type telegramRecorder struct {
	mu       sync.Mutex
	messages []string
}

func (r *telegramRecorder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if strings.HasSuffix(req.URL.Path, "/sendMessage") {
		if err := req.ParseMultipartForm(1 << 20); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		text := req.FormValue("text")

		r.mu.Lock()
		r.messages = append(r.messages, text)
		r.mu.Unlock()
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, `{"ok":true,"result":{}}`)
}

func (r *telegramRecorder) Messages() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	return append([]string(nil), r.messages...)
}

func (r *telegramRecorder) LastMessage() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.messages) == 0 {
		return ""
	}

	return r.messages[len(r.messages)-1]
}

type apiRecorder struct {
	mu     sync.Mutex
	paths  []string
	status int
	body   string
}

func (r *apiRecorder) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mu.Lock()
	r.paths = append(r.paths, req.URL.Path)
	status := r.status
	body := r.body
	r.mu.Unlock()

	if status == 0 {
		status = http.StatusOK
	}
	if body == "" {
		body = `{"status":"cancellation_requested"}`
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = io.WriteString(w, body)
}

func (r *apiRecorder) Paths() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	return append([]string(nil), r.paths...)
}

type testEnv struct {
	handler *Handler
	bot     *telegrambot.Bot
	tg      *telegramRecorder
	api     *apiRecorder
}

func newTestEnv(t *testing.T, apiStatus int) *testEnv {
	t.Helper()

	tgRecorder := &telegramRecorder{}
	tgServer := httptest.NewServer(tgRecorder)
	t.Cleanup(tgServer.Close)

	apiRec := &apiRecorder{status: apiStatus}
	apiServer := httptest.NewServer(apiRec)
	t.Cleanup(apiServer.Close)

	b, err := telegrambot.New(
		"123456:TEST",
		telegrambot.WithServerURL(tgServer.URL),
		telegrambot.WithSkipGetMe(),
	)
	if err != nil {
		t.Fatalf("create telegram bot: %v", err)
	}

	return &testEnv{
		handler: &Handler{
			client: generation.NewClient(apiServer.URL),
			state:  NewStateManager(),
		},
		bot: b,
		tg:  tgRecorder,
		api: apiRec,
	}
}

func (e *testEnv) run(t *testing.T, chatID, userID int64, text string) {
	t.Helper()

	e.handler.handleCommand(context.Background(), e.bot, chatID, userID, text)
}

func TestHandleCommand_Start(t *testing.T) {
	env := newTestEnv(t, http.StatusOK)

	env.run(t, 10, 10, "/start")

	want := "Welcome to Generation Bot! 🎨\n\nUse /generate to create an image."
	if got := env.tg.LastMessage(); got != want {
		t.Fatalf("unexpected message\n got: %q\nwant: %q", got, want)
	}
}

func TestHandleCommand_Help(t *testing.T) {
	env := newTestEnv(t, http.StatusOK)

	env.run(t, 10, 10, "/help")

	got := env.tg.LastMessage()
	for _, want := range []string{"/start", "/generate", "/cancel", "/help", "/about"} {
		if !strings.Contains(got, want) {
			t.Fatalf("help message does not mention %q: %q", want, got)
		}
	}
}

func TestHandleCommand_Generate(t *testing.T) {
	env := newTestEnv(t, http.StatusOK)

	env.run(t, 10, 10, "/generate")

	want := "Send me a prompt for image generation. 📝"
	if got := env.tg.LastMessage(); got != want {
		t.Fatalf("unexpected message\n got: %q\nwant: %q", got, want)
	}

	if got := env.handler.state.Get(10); got != StateWaitingForPrompt {
		t.Fatalf("unexpected state: got %v, want %v", got, StateWaitingForPrompt)
	}
}

func TestHandleCommand_CancelWithoutActiveGeneration(t *testing.T) {
	env := newTestEnv(t, http.StatusOK)

	env.run(t, 10, 10, "/cancel")

	want := "There is no active generation."
	if got := env.tg.LastMessage(); got != want {
		t.Fatalf("unexpected message\n got: %q\nwant: %q", got, want)
	}

	if paths := env.api.Paths(); len(paths) != 0 {
		t.Fatalf("cancel API must not be called, got %v", paths)
	}
}

func TestHandleCommand_CancelWhileWaitingForPrompt(t *testing.T) {
	env := newTestEnv(t, http.StatusOK)
	env.handler.state.Set(10, StateWaitingForPrompt)

	env.run(t, 10, 10, "/cancel")

	want := "Prompt cancelled."
	if got := env.tg.LastMessage(); got != want {
		t.Fatalf("unexpected message\n got: %q\nwant: %q", got, want)
	}

	if got := env.handler.state.Get(10); got != StateIdle {
		t.Fatalf("unexpected state: got %v, want %v", got, StateIdle)
	}

	if paths := env.api.Paths(); len(paths) != 0 {
		t.Fatalf("cancel API must not be called, got %v", paths)
	}
}

func TestHandleCommand_CancelActiveGeneration(t *testing.T) {
	env := newTestEnv(t, http.StatusOK)
	env.handler.state.SetTask(10, "task-abc")

	env.run(t, 10, 10, "/cancel")

	want := "Cancellation requested. ✋"
	if got := env.tg.LastMessage(); got != want {
		t.Fatalf("unexpected message\n got: %q\nwant: %q", got, want)
	}

	paths := env.api.Paths()
	if len(paths) != 1 || paths[0] != "/cancel/task-abc" {
		t.Fatalf("unexpected cancel API calls: %v", paths)
	}
}

func TestHandleCommand_CancelActiveGenerationAPIError(t *testing.T) {
	env := newTestEnv(t, http.StatusNotFound)
	env.handler.state.SetTask(10, "task-abc")

	env.run(t, 10, 10, "/cancel")

	want := "Could not cancel generation. It may have already finished."
	if got := env.tg.LastMessage(); got != want {
		t.Fatalf("unexpected message\n got: %q\nwant: %q", got, want)
	}

	paths := env.api.Paths()
	if len(paths) != 1 || paths[0] != "/cancel/task-abc" {
		t.Fatalf("unexpected cancel API calls: %v", paths)
	}
}

func TestHandleCommand_CancelOfAnotherUserTask(t *testing.T) {
	env := newTestEnv(t, http.StatusOK)
	env.handler.state.SetTask(20, "task-other")

	env.run(t, 10, 10, "/cancel")

	want := "There is no active generation."
	if got := env.tg.LastMessage(); got != want {
		t.Fatalf("unexpected message\n got: %q\nwant: %q", got, want)
	}

	if paths := env.api.Paths(); len(paths) != 0 {
		t.Fatalf("cancel API must not be called, got %v", paths)
	}
}

func TestHandleCommand_Unknown(t *testing.T) {
	env := newTestEnv(t, http.StatusOK)

	env.run(t, 10, 10, "/unknown")

	want := "Unknown command. Use /help to see available commands."
	if got := env.tg.LastMessage(); got != want {
		t.Fatalf("unexpected message\n got: %q\nwant: %q", got, want)
	}
}

func TestHandleCommand_IgnoresBotSuffix(t *testing.T) {
	env := newTestEnv(t, http.StatusOK)

	env.run(t, 10, 10, "/generate@my_generation_bot")

	want := "Send me a prompt for image generation. 📝"
	if got := env.tg.LastMessage(); got != want {
		t.Fatalf("unexpected message\n got: %q\nwant: %q", got, want)
	}

	if got := env.handler.state.Get(10); got != StateWaitingForPrompt {
		t.Fatalf("unexpected state: got %v, want %v", got, StateWaitingForPrompt)
	}
}
