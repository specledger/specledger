# ML Image Processing Template

## Overview

This template provides a machine learning project structure for image processing tasks including classification, detection, and segmentation. It follows the standard data science project layout with clear separation of data, models, and experiments.

## Technology Stack

- **Language**: Python 3.10+
- **ML Framework**: PyTorch or TensorFlow
- **Data Processing**: NumPy, Pillow, OpenCV
- **Experiment Tracking**: Weights & Biases or MLflow
- **Environment**: pyproject.toml + requirements.txt

## Directory Structure

```
.
├── src/
│   ├── data/
│   │   └── dataset.py      # Dataset loading and augmentation
│   ├── models/
│   │   └── cnn_model.py    # Model architecture definitions
│   ├── training/
│   │   └── train.py        # Training loop and utilities
│   ├── inference/          # Model serving and prediction
│   └── utils/              # Helper utilities
├── data/
│   ├── raw/                # Original, unprocessed data
│   ├── processed/          # Cleaned and transformed data
│   └── interim/            # Intermediate processing
├── models/
│   └── checkpoints/        # Saved model checkpoints
├── notebooks/              # Jupyter notebooks for exploration
├── configs/                # Training configuration files
├── tests/                  # Unit and integration tests
├── requirements.txt
└── pyproject.toml
```

## Development Commands

### Setup Environment
```bash
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

### Training
```bash
python -m src.training.train --config configs/train.yaml
```

### Evaluation
```bash
python -m src.training.evaluate --model models/checkpoints/best.pt
```

### Testing
```bash
pytest tests/
```

## ML Development Guidelines

### Data Pipeline
1. Place raw data in `data/raw/`
2. Create preprocessing scripts in `src/data/`
3. Save processed data to `data/processed/`

### Model Development
1. Define architectures in `src/models/`
2. Training logic in `src/training/`
3. Save checkpoints to `models/checkpoints/`

### Experiment Tracking
- Log hyperparameters and metrics
- Version datasets alongside models
- Document model performance

## Code Guidelines

- Use type hints consistently
- Document functions with docstrings
- Keep notebooks clean (restart & run all before commit)
- Pin dependency versions

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
