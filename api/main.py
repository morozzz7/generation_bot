from io import BytesIO
import time
import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI, HTTPException, Request
from fastapi.responses import StreamingResponse
import torch

from .model.inference import ImageGenerator
from .schemas import GenerateRequest

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    logger.info("Loading image generation model...")

    app.state.generator = ImageGenerator()

    logger.info("Model loaded successfully")

    yield

    logger.info("Shutting down FastAPI")

app = FastAPI(
    title="Image Generation API",
    description="Local Stable Diffusion image generation service",
    version="1.0.0",
    lifespan=lifespan,
)


@app.get("/health")
async def health(request: Request):
    generator = getattr(request.app.state, "generator", None)

    if generator is None:
        raise HTTPException(
            status_code=503,
            detail="Model is not loaded",
        )

    return {
        "status": "ok",
        "model_loaded": True,
        "device": str(generator.pipe.device),
    }


@app.post("/generate")
def generate(
    data: GenerateRequest,
    request: Request,
):
    generator = getattr(request.app.state, "generator", None)

    if generator is None:
        raise HTTPException(
            status_code=503,
            detail="Model is not loaded",
        )

    try:
        start_time = time.perf_counter()

        image, seed = generator.generate(
            prompt=data.prompt,
            negative_prompt=data.negative_prompt,
            steps=data.steps,
            seed=data.seed,
        )

        buffer = BytesIO()
        image.save(buffer, format="PNG")
        buffer.seek(0)

        elapsed = time.perf_counter() - start_time

        logger.info(
            "Image generated in %.2fs, seed=%d",
            elapsed,
            seed,
        )

        return StreamingResponse(
            buffer,
            media_type="image/png",
            headers={
                "X-Generation-Seed": str(seed),
                "X-Generation-Time": f"{elapsed:.2f}",
            },
        )

    except torch.cuda.OutOfMemoryError:
        logger.exception("CUDA out of memory")

        raise HTTPException(
            status_code=503,
            detail="GPU memory exhausted. Try again later.",
        )

    except Exception:
        logger.exception("Image generation failed")

        raise HTTPException(
            status_code=500,
            detail="Image generation failed",
        )


