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
    ):
        if not prompt:
            raise ValueError('prompt cannot be empty')

        if seed < 0:
            seed = secrets.randbits(32)

        generator = torch.Generator(
            device=self.pipe.device
        ).manual_seed(seed)

        start_time = time.perf_counter()

        with self._lock:
            image = self.pipe(
                prompt=prompt,
                negative_prompt=negative_prompt,
                num_inference_steps=steps,
                guidance_scale=7.0,
                width=512,
                height=512,
                generator=generator,
            ).images[0]

            if self.pipe.device.type == "cuda":
                torch.cuda.synchronize()

        elapsed = time.perf_counter() - start_time

        print(f"Generation time: {elapsed:.2f} sec")
        print(f"Seed: {seed}")

        return image, seed
