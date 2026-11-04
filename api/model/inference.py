
import gc
import secrets
import threading
import time

import torch

from .loader import load_model


class GenerationCancelled(Exception):
    """Raised when an image generation task is cancelled."""


class ImageGenerator:
    def __init__(self):
        self.pipe = None
        self.current_model = None
        self._lock = threading.Lock()

    def _get_pipeline(self, model_id: str):
        if self.current_model == model_id:
            return self.pipe

        print(f"Switching model to {model_id}...")

        if self.pipe is not None:
            del self.pipe
            self.pipe = None
            self.current_model = None

            gc.collect()

            if torch.cuda.is_available():
                torch.cuda.empty_cache()

        self.pipe = load_model(model_id)
        self.current_model = model_id

        return self.pipe

    @torch.inference_mode()
    def generate(
        self,
        model_id: str,
        prompt: str,
        negative_prompt: str = "",
        steps: int | None = None,
        seed: int = -1,
        cancel_event: threading.Event | None = None,
    ):
        if not prompt.strip():
            raise ValueError("Prompt cannot be empty")

        if model_id not in ("dreamshaper", "sdxl_turbo"):
            raise ValueError(f"Unknown model: {model_id}")

        if model_id == "sdxl_turbo":
            steps = 4
        elif model_id == "dreamshaper":
            steps = 25
        else:
            raise ValueError(f"Unknown model: {model_id}")

        if seed < 0:
            seed = secrets.randbits(32)

        start_time = time.perf_counter()

        with self._lock:
            if cancel_event is not None and cancel_event.is_set():
                raise GenerationCancelled()

            pipe = self._get_pipeline(model_id)

            if cancel_event is not None and cancel_event.is_set():
                raise GenerationCancelled()

            generator = torch.Generator(device="cpu").manual_seed(seed)

            def callback_on_step_end(pipe, step, timestep, callback_kwargs):
                if cancel_event is not None and cancel_event.is_set():
                    raise GenerationCancelled()

                return callback_kwargs

            params = {
                "prompt": prompt,
                "num_inference_steps": steps,
                "generator": generator,
                "callback_on_step_end": callback_on_step_end,
                "width": 512,
                "height": 512,
            }

            if model_id == "sdxl_turbo":
                params["guidance_scale"] = 0.0
            else:
                params["guidance_scale"] = 7.0
                params["negative_prompt"] = negative_prompt or None

            image = pipe(**params).images[0]

            if torch.cuda.is_available():
                torch.cuda.synchronize()

        elapsed = time.perf_counter() - start_time

        print(
            f"Generation completed in {elapsed:.2f}s "
            f"(model={model_id}, seed={seed}, steps={steps})"
        )

        return image, seed
