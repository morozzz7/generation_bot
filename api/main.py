import io
from uuid import UUID
import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI, HTTPException, Request
from fastapi.responses import StreamingResponse
import torch

from .model.inference import ImageGenerator, GenerationCancelled
from .schemas import GenerateRequest
from .task_manager import TaskManager


logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    logger.info("Loading image generation model...")

    app.state.generator = ImageGenerator()
    app.state.task_manager = TaskManager()

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
def generate(request: GenerateRequest):
    task_id = str(request.task_id)
    task_manager: TaskManager = app.state.task_manager
    generator: ImageGenerator = app.state.generator

    try:
        cancel_event = task_manager.register(task_id)
    except ValueError:
        raise HTTPException(
            status_code=409,
            detail="Task ID already exists",
        )

    try:
        logger.info("Starting generation: task_id=%s", task_id)

        image, seed = generator.generate(
            prompt=request.prompt,
            negative_prompt=request.negative_prompt,
            steps=request.steps,
            seed=request.seed,
            cancel_event=cancel_event,
        )

        if cancel_event.is_set():
            raise GenerationCancelled()

        image_buffer = io.BytesIO()
        image.save(image_buffer, format="PNG")
        image_buffer.seek(0)

        logger.info("Generation completed: task_id=%s", task_id)

        return StreamingResponse(
            image_buffer,
            media_type="image/png",
            headers={
                "X-Generation-Task-ID": task_id,
                "X-Generation-Seed": str(seed),
            },
        )

    except GenerationCancelled:
        logger.info("Generation cancelled: task_id=%s", task_id)

        raise HTTPException(
            status_code=409,
            detail="Generation cancelled",
        )

    except torch.cuda.OutOfMemoryError:
        logger.exception("CUDA out of memory: task_id=%s", task_id)

        raise HTTPException(
            status_code=503,
            detail="GPU memory exhausted",
        )

    except Exception:
        logger.exception("Generation failed: task_id=%s", task_id)

        raise HTTPException(
            status_code=500,
            detail="Image generation failed",
        )

    finally:
        task_manager.remove(task_id)


@app.post("/cancel/{task_id}")
def cancel_generation(task_id: UUID):
    task_manager: TaskManager = app.state.task_manager

    cancelled = task_manager.cancel(str(task_id))

    if not cancelled:
        raise HTTPException(
            status_code=404,
            detail="Task not found or already finished",
        )

    logger.info("Cancellation requested: task_id=%s", task_id)

    return {
        "task_id": str(task_id),
        "status": "cancellation_requested",
    }
