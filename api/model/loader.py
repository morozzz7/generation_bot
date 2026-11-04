import torch

from diffusers import (
    StableDiffusionPipeline,
    AutoPipelineForText2Image,
    DPMSolverMultistepScheduler,
)

DEVICE = "cuda" if torch.cuda.is_available() else "cpu"

DREAMSHAPER_ID = "Lykon/dreamshaper-8"
SDXL_TURBO_ID = "stabilityai/sdxl-turbo"


def load_model(model_id: str):
    print(f"Loading {model_id}...")
    print(f"Device: {DEVICE}")

    if model_id == "dreamshaper":
        pipe = StableDiffusionPipeline.from_pretrained(
            DREAMSHAPER_ID,
            torch_dtype=torch.float32,
            use_safetensors=True,
        )

        pipe.scheduler = DPMSolverMultistepScheduler.from_config(
            pipe.scheduler.config,
            algorithm_type="dpmsolver++",
            final_sigmas_type="sigma_min",
        )

        pipe = pipe.to(DEVICE)

    elif model_id == "sdxl_turbo":
        pipe = AutoPipelineForText2Image.from_pretrained(
            SDXL_TURBO_ID,
            torch_dtype=torch.float16,
            variant="fp16",
            use_safetensors=True,
        )

        if DEVICE == "cuda":
            pipe.enable_model_cpu_offload()
        else:
            pipe = pipe.to(DEVICE)

    else:
        raise ValueError(f"Unknown model: {model_id}")

    pipe.set_progress_bar_config(disable=False)

    print(f"{model_id} loaded successfully!")

    return pipe
