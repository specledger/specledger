# ML Image Processing Template

PyTorch-based machine learning project for image processing.

## Structure

```
src/
├── data/
│   ├── dataset.py        # Dataset definitions
│   ├── preprocessing.py  # Image preprocessing
│   └── augmentation.py   # Data augmentation
├── models/
│   └── cnn_model.py      # Model architecture
├── training/
│   ├── train.py          # Training loop
│   └── evaluate.py       # Evaluation metrics
├── inference/
│   └── predict.py        # Inference pipeline
└── utils/                # Utility functions

data/
├── raw/                  # Original data
├── processed/            # Processed data
└── interim/              # Intermediate data

models/checkpoints/       # Model checkpoints
notebooks/                # Jupyter notebooks
configs/                  # Configuration files
tests/                    # Test files
```

## Technologies

- **Framework**: PyTorch + PyTorch Lightning
- **Vision**: TorchVision (pre-trained models)
- **Tracking**: Weights & Biases
- **Processing**: OpenCV

## Getting Started

1. Install: `pip install -r requirements.txt`
2. Train: `python src/training/train.py`
3. Evaluate: `python src/training/evaluate.py`
