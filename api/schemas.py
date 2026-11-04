from pydantic import BaseModel, Field
from uuid import UUID
from typing import Literal


class GenerateRequest(BaseModel):
    task_id: UUID

    model_id: Literal["dreamshaper", "sdxl_turbo"] = "dreamshaper"

    prompt: str = Field(min_length=1, max_length=2000)
    negative_prompt: str = Field(default="", max_length=2000)

    steps: int = Field(default=25, ge=1, le=50)
    seed: int = Field(default=-1, ge=-1, le=4294967295)
