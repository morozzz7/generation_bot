import time
import secrets
import torch
import threading

from .loader import load_model


class ImageGenerator:
    def __init__(self):
        self.pipe = load_model()
        self._lock = threading.Lock()

    @torch.inference_mode()
    def generate(
    self,
    prompt: str,
    negative_prompt: str = "",
    steps: int = 25,
    seed: int = -1,
    cancel_event: threading.Event | None = None,
):
        if not prompt.strip():
            raise ValueError("Prompt cannot be empty")

        if seed < 0:
            seed = secrets.randbits(32)

        generator = torch.Generator(
            device=self.pipe.device
        ).manual_seed(seed)

        def callback_on_step_end(pipe, step, timestep, callback_kwargs):
            if cancel_event is not None and cancel_event.is_set():
                raise GenerationCancelled()

            return callback_kwargs

        start_time = time.perf_counter()

        with self._lock:
            if cancel_event is not None and cancel_event.is_set():
                raise GenerationCancelled()

            image = self.pipe(
                prompt=prompt,
                negative_prompt=negative_prompt or None,
                num_inference_steps=steps,
                guidance_scale=7.0,
                width=512,
                height=512,
                generator=generator,
                callback_on_step_end=callback_on_step_end,
            ).images[0]

            if self.pipe.device.type == "cuda":
                torch.cuda.synchronize()

        elapsed = time.perf_counter() - start_time

        print(
            f"Generation completed in {elapsed:.2f}s "
            f"(seed={seed}, steps={steps})"
        )

        return image, seed

    
class GenerationCancelled(Exception):
    """Raised when an image generation task is cancelled."""