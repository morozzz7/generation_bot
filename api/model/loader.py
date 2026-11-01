import torch

from diffusers import (
    StableDiffusionPipeline,
    DPMSolverMultistepScheduler,
)

MODEL_ID = "Lykon/dreamshaper-8"

DEVICE = "cuda" if torch.cuda.is_available() else "cpu"

DTYPE = torch.float32


def load_model():
    print(f"Loading {MODEL_ID}...")
    print(f"Device: {DEVICE}")
    print(f"Dtype: {DTYPE}")

    pipe = StableDiffusionPipeline.from_pretrained(
        MODEL_ID,
        torch_dtype=DTYPE,
        use_safetensors=True,
    )

    pipe.scheduler = DPMSolverMultistepScheduler.from_config(
        pipe.scheduler.config,
        algorithm_type="dpmsolver++",
        final_sigmas_type="sigma_min",
    )

    pipe = pipe.to(DEVICE)

    pipe.set_progress_bar_config(disable=False)

    print("Model loaded successfully!")

    return pipe
