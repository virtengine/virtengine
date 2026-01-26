"""
Model export for TensorFlow-Go inference.

Exports trained models to TensorFlow SavedModel format with
fixed input/output signatures compatible with Go inference.
"""

import logging
import os
import json
import hashlib
import time
from dataclasses import dataclass, field
from typing import Dict, Any, Optional
from pathlib import Path
from datetime import datetime

import numpy as np

from ml.training.config import ExportConfig, ModelConfig
from ml.training.model.architecture import TrustScoreModel

logger = logging.getLogger(__name__)

# TensorFlow import
try:
    import tensorflow as tf
    TF_AVAILABLE = True
except ImportError:
    TF_AVAILABLE = False


@dataclass
class ExportResult:
    """Result of model export operation."""
    
    # Export info
    model_path: str = ""
    model_hash: str = ""
    version: str = ""
    
    # Signatures
    input_signature: Dict[str, Any] = field(default_factory=dict)
    output_signature: Dict[str, Any] = field(default_factory=dict)
    
    # Metadata
    export_timestamp: str = ""
    tensorflow_version: str = ""
    export_format: str = "SavedModel"
    
    # Size info
    model_size_bytes: int = 0
    
    # Status
    success: bool = True
    error_message: Optional[str] = None
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization."""
        return {
            "model_path": self.model_path,
            "model_hash": self.model_hash,
            "version": self.version,
            "input_signature": self.input_signature,
            "output_signature": self.output_signature,
            "export_timestamp": self.export_timestamp,
            "tensorflow_version": self.tensorflow_version,
            "export_format": self.export_format,
            "model_size_bytes": self.model_size_bytes,
            "success": self.success,
            "error_message": self.error_message,
        }
    
    def saVIRTENGINE_metadata(self, filepath: str) -> None:
        """Save export metadata to JSON."""
        Path(filepath).parent.mkdir(parents=True, exist_ok=True)
        with open(filepath, 'w') as f:
            json.dump(self.to_dict(), f, indent=2)


class ModelExporter:
    """
    Exports trained models to TensorFlow SavedModel format.
    
    Features:
    - Fixed input/output signatures for Go inference
    - Model versioning with hash
    - Optional quantization for smaller size
    - Metadata export for reproducibility
    """
    
    def __init__(self, config: Optional[ExportConfig] = None):
        """
        Initialize the exporter.
        
        Args:
            config: Export configuration
        """
        self.config = config or ExportConfig()
    
    def export_savedmodel(
        self,
        model: TrustScoreModel,
        export_path: Optional[str] = None,
        version: Optional[str] = None,
    ) -> ExportResult:
        """
        Export model as TensorFlow SavedModel.
        
        The exported model has fixed input/output signatures compatible
        with TensorFlow-Go inference in the blockchain node.
        
        Args:
            model: Trained TrustScoreModel
            export_path: Path to export to (uses config if not provided)
            version: Version string (generated if not provided)
            
        Returns:
            ExportResult with export information
        """
        if not TF_AVAILABLE:
            return ExportResult(
                success=False,
                error_message="TensorFlow is required for export"
            )
        
        try:
            # Determine paths and version
            export_dir = export_path or self.config.export_dir
            version = version or self._generate_version()
            
            versioned_path = os.path.join(export_dir, version)
            Path(versioned_path).mkdir(parents=True, exist_ok=True)
            
            model_path = os.path.join(versioned_path, "model")
            
            # Get the Keras model
            keras_model = model.keras_model
            
            # Create concrete function with fixed signature
            input_dim = keras_model.input_shape[1]
            
            @tf.function(input_signature=[
                tf.TensorSpec(shape=[None, input_dim], dtype=tf.float32, name=self.config.input_name)
            ])
            def serving_fn(features):
                """Serving function with fixed signature."""
                return {self.config.output_name: keras_model(features, training=False)}
            
            # Export with signature
            signatures = {
                self.config.signature_name: serving_fn,
            }
            
            # Export model
            tf.saved_model.save(
                keras_model,
                model_path,
                signatures=signatures,
            )
            
            # Compute model hash
            model_hash = self.compute_model_hash(model_path)
            
            # Get model size
            model_size = self._get_directory_size(model_path)
            
            # Build result
            result = ExportResult(
                model_path=model_path,
                model_hash=model_hash,
                version=version,
                input_signature={
                    "name": self.config.input_name,
                    "shape": [None, input_dim],
                    "dtype": "float32",
                },
                output_signature={
                    "name": self.config.output_name,
                    "shape": [None, 1],
                    "dtype": "float32",
                    "range": [0, 100],
                },
                export_timestamp=datetime.utcnow().isoformat(),
                tensorflow_version=tf.__version__,
                model_size_bytes=model_size,
                success=True,
            )
            
            # Save metadata
            metadata_path = os.path.join(versioned_path, "export_metadata.json")
            result.saVIRTENGINE_metadata(metadata_path)
            
            # Optionally quantize
            if self.config.quantize:
                self._quantize_model(model_path, result)
            
            logger.info(f"Model exported to {model_path}")
            logger.info(f"Version: {version}, Hash: {model_hash}")
            
            return result
            
        except Exception as e:
            logger.error(f"Export failed: {e}")
            return ExportResult(
                success=False,
                error_message=str(e),
            )
    
    def _generate_version(self) -> str:
        """Generate a version string."""
        prefix = self.config.version_prefix
        
        if self.config.include_timestamp:
            timestamp = datetime.utcnow().strftime("%Y%m%d_%H%M%S")
            return f"{prefix}{timestamp}"
        else:
            return f"{prefix}1.0.0"
    
    def compute_model_hash(self, model_path: str) -> str:
        """
        Compute SHA256 hash of model weights.
        
        Args:
            model_path: Path to SavedModel directory
            
        Returns:
            SHA256 hash string
        """
        if not TF_AVAILABLE:
            return ""
        
        try:
            # Load the saved model
            loaded_model = tf.saved_model.load(model_path)
            
            # Get all variables
            variables = loaded_model.variables
            
            # Concatenate all weights
            all_weights = []
            for var in variables:
                all_weights.append(var.numpy().tobytes())
            
            # Compute hash
            combined = b"".join(all_weights)
            return hashlib.sha256(combined).hexdigest()
            
        except Exception as e:
            logger.warning(f"Could not compute model hash: {e}")
            
            # Fallback: hash the files
            return self._hash_directory(model_path)
    
    def _hash_directory(self, dir_path: str) -> str:
        """Hash all files in a directory."""
        hasher = hashlib.sha256()
        
        for root, dirs, files in os.walk(dir_path):
            dirs.sort()
            files.sort()
            for filename in files:
                filepath = os.path.join(root, filename)
                with open(filepath, 'rb') as f:
                    hasher.update(f.read())
        
        return hasher.hexdigest()
    
    def _get_directory_size(self, dir_path: str) -> int:
        """Get total size of a directory in bytes."""
        total_size = 0
        for root, dirs, files in os.walk(dir_path):
            for filename in files:
                filepath = os.path.join(root, filename)
                total_size += os.path.getsize(filepath)
        return total_size
    
    def _quantize_model(self, model_path: str, result: ExportResult) -> None:
        """Apply quantization to the exported model."""
        try:
            # TensorFlow Lite conversion for quantization
            converter = tf.lite.TFLiteConverter.from_saved_model(model_path)
            
            if self.config.quantization_type == "float16":
                converter.optimizations = [tf.lite.Optimize.DEFAULT]
                converter.target_spec.supported_types = [tf.float16]
            elif self.config.quantization_type == "int8":
                converter.optimizations = [tf.lite.Optimize.DEFAULT]
                # Would need representative dataset for full int8
            
            tflite_model = converter.convert()
            
            # Save TFLite model
            tflite_path = os.path.join(
                os.path.dirname(model_path),
                "model.tflite"
            )
            with open(tflite_path, 'wb') as f:
                f.write(tflite_model)
            
            logger.info(f"Quantized model saved to {tflite_path}")
            
        except Exception as e:
            logger.warning(f"Quantization failed: {e}")
    
    def verify_export(self, export_path: str) -> bool:
        """
        Verify that an exported model can be loaded and used.
        
        Args:
            export_path: Path to exported SavedModel
            
        Returns:
            True if verification passes
        """
        if not TF_AVAILABLE:
            return False
        
        try:
            # Load the model
            loaded = tf.saved_model.load(export_path)
            
            # Get the serving function
            serving_fn = loaded.signatures[self.config.signature_name]
            
            # Create dummy input
            input_spec = serving_fn.structured_input_signature[1]
            input_tensor = list(input_spec.values())[0]
            input_dim = input_tensor.shape[1] if input_tensor.shape[1] else 768
            
            dummy_input = tf.constant(np.random.randn(1, input_dim).astype(np.float32))
            
            # Run inference
            output = serving_fn(dummy_input)
            
            # Check output
            output_name = self.config.output_name
            if output_name in output:
                result = output[output_name].numpy()
                logger.info(f"Verification passed: output shape {result.shape}")
                return True
            else:
                logger.error(f"Output '{output_name}' not found in model")
                return False
                
        except Exception as e:
            logger.error(f"Verification failed: {e}")
            return False
    
    def get_go_inference_example(self, result: ExportResult) -> str:
        """
        Generate example Go code for inference.
        
        Args:
            result: Export result
            
        Returns:
            Go code example as string
        """
        return f'''// Go inference example for trust score model
// Model version: {result.version}
// Model hash: {result.model_hash}

package main

import (
    tf "github.com/tensorflow/tensorflow/tensorflow/go"
)

func LoadTrustScoreModel(modelPath string) (*tf.SavedModel, error) {{
    return tf.LoadSavedModel(modelPath, []string{{"serve"}}, nil)
}}

func Predict(model *tf.SavedModel, features []float32) (float32, error) {{
    // Create input tensor
    // Input shape: [1, {result.input_signature.get("shape", [None, 768])[1]}]
    // Input name: "{result.input_signature.get("name", "features")}"
    
    inputTensor, err := tf.NewTensor([][]float32{{features}})
    if err != nil {{
        return 0, err
    }}
    
    // Run inference
    // Output name: "{result.output_signature.get("name", "trust_score")}"
    results, err := model.Session.Run(
        map[tf.Output]*tf.Tensor{{
            model.Graph.Operation("{result.input_signature.get("name", "features")}").Output(0): inputTensor,
        }},
        []tf.Output{{
            model.Graph.Operation("{result.output_signature.get("name", "trust_score")}").Output(0),
        }},
        nil,
    )
    if err != nil {{
        return 0, err
    }}
    
    // Extract trust score (0-100)
    trustScore := results[0].Value().([][]float32)[0][0]
    return trustScore, nil
}}
'''


def export_model(
    model: TrustScoreModel,
    export_path: str,
    version: Optional[str] = None,
    config: Optional[ExportConfig] = None,
) -> ExportResult:
    """
    Convenience function to export a model.
    
    Args:
        model: Trained model
        export_path: Export directory
        version: Version string
        config: Export configuration
        
    Returns:
        ExportResult
    """
    exporter = ModelExporter(config)
    return exporter.export_savedmodel(model, export_path, version)
