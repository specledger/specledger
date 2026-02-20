"""Dataset definitions for image processing."""

from pathlib import Path
from typing import Optional, Callable

import torch
from torch.utils.data import Dataset
from PIL import Image


class ImageDataset(Dataset):
    """Custom dataset for image loading and processing."""

    def __init__(
        self,
        root_dir: str | Path,
        transform: Optional[Callable] = None,
        extensions: tuple = (".jpg", ".jpeg", ".png"),
    ):
        """
        Initialize the dataset.

        Args:
            root_dir: Root directory containing images
            transform: Optional transform to apply to images
            extensions: Valid image file extensions
        """
        self.root_dir = Path(root_dir)
        self.transform = transform
        self.extensions = extensions
        self.image_paths = self._load_image_paths()

    def _load_image_paths(self) -> list[Path]:
        """Load all valid image paths from root directory."""
        paths = []
        for ext in self.extensions:
            paths.extend(self.root_dir.rglob(f"*{ext}"))
        return sorted(paths)

    def __len__(self) -> int:
        return len(self.image_paths)

    def __getitem__(self, idx: int) -> torch.Tensor:
        image_path = self.image_paths[idx]
        image = Image.open(image_path).convert("RGB")

        if self.transform:
            image = self.transform(image)

        return image
