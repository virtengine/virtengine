"""VirtEngine VEID ML pipeline package.

The subpackages under ``ml/`` (document_preprocessing, face_extraction,
facial_verification, ocr_extraction, text_detection, training) implement the
deterministic identity-verification pipeline (VE-219). The runnable entry
point for the container image is :mod:`ml.pipeline_runner`.

All pipeline code assumes absolute ``ml.*`` imports with the repository root
(or ``/app`` in the image) on ``sys.path``, CPU-only execution, and the
determinism environment enforced by :func:`ml.pipeline_runner.main`.
"""

PIPELINE_VERSION = "1.0.0"

__all__ = ["PIPELINE_VERSION"]
