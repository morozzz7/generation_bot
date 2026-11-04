import threading

class TaskManager:
    def __init__(self):
        self._lock = threading.Lock()
        self._tasks: dict[str, threading.Event] = {}

    def register(self, task_id: str) -> threading.Event:
        with self._lock:
            if task_id in self._tasks:
                raise ValueError("Task already exists")

            event = threading.Event()
            self._tasks[task_id] = event

            return event

    def cancel(self, task_id: str) -> bool:
        with self._lock:
            event = self._tasks.get(task_id)

            if event is None:
                return False

            event.set()
            return True

    def remove(self, task_id: str):
        with self._lock:
            self._tasks.pop(task_id, None)