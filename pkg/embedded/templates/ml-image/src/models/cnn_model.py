"""CNN model architecture for image classification."""

import torch
import torch.nn as nn
import pytorch_lightning as pl
from torchvision import models


class ImageClassifier(pl.LightningModule):
    """Lightning module for image classification."""

    def __init__(
        self,
        num_classes: int,
        learning_rate: float = 1e-3,
        pretrained: bool = True,
    ):
        """
        Initialize the classifier.

        Args:
            num_classes: Number of output classes
            learning_rate: Learning rate for optimizer
            pretrained: Whether to use pretrained weights
        """
        super().__init__()
        self.save_hyperparameters()

        # Use ResNet18 as backbone
        weights = models.ResNet18_Weights.DEFAULT if pretrained else None
        self.backbone = models.resnet18(weights=weights)

        # Replace classifier head
        in_features = self.backbone.fc.in_features
        self.backbone.fc = nn.Linear(in_features, num_classes)

        self.criterion = nn.CrossEntropyLoss()

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        return self.backbone(x)

    def training_step(self, batch: tuple, batch_idx: int) -> torch.Tensor:
        images, labels = batch
        outputs = self(images)
        loss = self.criterion(outputs, labels)
        self.log("train_loss", loss, prog_bar=True)
        return loss

    def validation_step(self, batch: tuple, batch_idx: int) -> torch.Tensor:
        images, labels = batch
        outputs = self(images)
        loss = self.criterion(outputs, labels)
        acc = (outputs.argmax(dim=1) == labels).float().mean()
        self.log("val_loss", loss, prog_bar=True)
        self.log("val_acc", acc, prog_bar=True)
        return loss

    def configure_optimizers(self):
        return torch.optim.AdamW(
            self.parameters(),
            lr=self.hparams.learning_rate,
        )
