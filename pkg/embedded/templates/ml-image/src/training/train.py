"""Training script for image classification."""

import argparse
from pathlib import Path

import pytorch_lightning as pl
from pytorch_lightning.callbacks import ModelCheckpoint, EarlyStopping
from pytorch_lightning.loggers import WandbLogger
from torch.utils.data import DataLoader
from torchvision import transforms

from src.data.dataset import ImageDataset
from src.models.cnn_model import ImageClassifier


def get_transforms(train: bool = True) -> transforms.Compose:
    """Get image transforms for training or validation."""
    if train:
        return transforms.Compose([
            transforms.RandomResizedCrop(224),
            transforms.RandomHorizontalFlip(),
            transforms.ToTensor(),
            transforms.Normalize(
                mean=[0.485, 0.456, 0.406],
                std=[0.229, 0.224, 0.225],
            ),
        ])
    return transforms.Compose([
        transforms.Resize(256),
        transforms.CenterCrop(224),
        transforms.ToTensor(),
        transforms.Normalize(
            mean=[0.485, 0.456, 0.406],
            std=[0.229, 0.224, 0.225],
        ),
    ])


def main(args: argparse.Namespace) -> None:
    """Main training function."""
    # Create datasets
    train_dataset = ImageDataset(
        args.data_dir / "train",
        transform=get_transforms(train=True),
    )
    val_dataset = ImageDataset(
        args.data_dir / "val",
        transform=get_transforms(train=False),
    )

    # Create dataloaders
    train_loader = DataLoader(
        train_dataset,
        batch_size=args.batch_size,
        shuffle=True,
        num_workers=args.num_workers,
    )
    val_loader = DataLoader(
        val_dataset,
        batch_size=args.batch_size,
        num_workers=args.num_workers,
    )

    # Create model
    model = ImageClassifier(
        num_classes=args.num_classes,
        learning_rate=args.lr,
    )

    # Setup callbacks
    callbacks = [
        ModelCheckpoint(
            dirpath=args.checkpoint_dir,
            filename="best-{epoch:02d}-{val_acc:.2f}",
            monitor="val_acc",
            mode="max",
        ),
        EarlyStopping(monitor="val_loss", patience=5),
    ]

    # Setup logger
    logger = WandbLogger(project=args.project_name) if args.use_wandb else None

    # Train
    trainer = pl.Trainer(
        max_epochs=args.epochs,
        callbacks=callbacks,
        logger=logger,
        accelerator="auto",
    )
    trainer.fit(model, train_loader, val_loader)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--data-dir", type=Path, default=Path("data/processed"))
    parser.add_argument("--checkpoint-dir", type=Path, default=Path("models/checkpoints"))
    parser.add_argument("--num-classes", type=int, required=True)
    parser.add_argument("--batch-size", type=int, default=32)
    parser.add_argument("--epochs", type=int, default=100)
    parser.add_argument("--lr", type=float, default=1e-3)
    parser.add_argument("--num-workers", type=int, default=4)
    parser.add_argument("--project-name", type=str, default="ml-image")
    parser.add_argument("--use-wandb", action="store_true")
    main(parser.parse_args())
