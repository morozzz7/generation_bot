from io import BytesIO

from fastapi import FastAPI
from fastapi.responses import StreamingResponse
from pydantic import BaseModel
from PIL import Image

app = FastAPI()


class GenerateRequest(BaseModel):
    prompt: str


@app.post("/generate")
async def generate(request: GenerateRequest):
    print(f"Received prompt: {request.prompt}")

    # Временное тестовое изображение
    image = Image.new("RGB", (512, 512), color=(100, 120, 200))

    buffer = BytesIO()
    image.save(buffer, format="PNG")
    buffer.seek(0)

    return StreamingResponse(
        buffer,
        media_type="image/png"
    )


@app.get("/health")
async def health():
    return {"status": "ok"}
