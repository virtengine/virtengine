"""
Tests for model training, evaluation, and export.
"""

import pytest
import numpy as np
import tempfile
from datetime import datetime
from pathlib import Path

from ml.training.config import TrainingConfig, ModelConfig, ExportConfig
from ml.training.model.evaluation import ModelEvaluator, EvaluationMetrics

# Skip tests if TensorFlow not available
try:
    import tensorflow as tf
    from ml.training.model.architecture import TrustScoreModel, create_trust_score_model
    from ml.training.model.training import ModelTrainer, TrainingResult
    from ml.training.model.export import ModelExporter, ExportResult
    TF_AVAILABLE = True
except ImportError:
    TF_AVAILABLE = False


@pytest.mark.skipif(not TF_AVAILABLE, reason="TensorFlow not available")
class TestModelArchitecture:
    """Tests for model architecture."""
    
    def test_create_model(self):
        """Test creating a trust score model."""
        config = ModelConfig(
            input_dim=128,
            hidden_layers=[64, 32],
            dropout_rate=0.2,
        )
        
        model = create_trust_score_model(config)
        
        assert model is not None
        assert model.input_shape == (None, 128)
    
    def test_trust_score_model_class(self):
        """Test TrustScoreModel class."""
        config = ModelConfig(
            input_dim=64,
            hidden_layers=[32, 16],
        )
        
        model = TrustScoreModel(config)
        
        assert model is not None
        assert model.keras_model is not None
    
    def test_model_compile(self):
        """Test model compilation."""
        config = ModelConfig(input_dim=64, hidden_layers=[32])
        model = TrustScoreModel(config)
        
        model.compile(learning_rate=0.001)
        
        # Should be compiled successfully
        assert model._is_compiled is True
    
    def test_model_predict(self):
        """Test model prediction."""
        config = ModelConfig(input_dim=64, hidden_layers=[32])
        model = TrustScoreModel(config)
        model.compile()
        
        # Random input
        X = np.random.randn(10, 64).astype(np.float32)
        
        predictions = model.predict(X)
        
        assert predictions.shape == (10, 1)
        # Predictions should be in 0-100 range (sigmoid * 100)
        assert np.all(predictions >= 0)
        assert np.all(predictions <= 100)
    
    def test_model_summary(self):
        """Test getting model summary."""
        config = ModelConfig(input_dim=64, hidden_layers=[32])
        model = TrustScoreModel(config)
        
        summary = model.summary()
        
        assert "Model:" in summary
        assert "dense" in summary.lower()


@pytest.mark.skipif(not TF_AVAILABLE, reason="TensorFlow not available")
class TestModelTraining:
    """Tests for model training."""
    
    def test_trainer_creation(self, sample_config):
        """Test creating a model trainer."""
        trainer = ModelTrainer(sample_config)
        
        assert trainer is not None
    
    def test_train_from_arrays(self, sample_config, sample_train_arrays):
        """Test training from numpy arrays."""
        train_X, train_y, val_X, val_y, test_X, test_y = sample_train_arrays
        
        # Use small config for fast testing
        sample_config.model.epochs = 2
        sample_config.model.batch_size = 16
        
        trainer = ModelTrainer(sample_config)
        
        result = trainer.train_from_arrays(
            train_X, train_y,
            val_X, val_y,
        )
        
        assert isinstance(result, TrainingResult)
        assert result.model is not None
        assert result.total_epochs == 2
        assert "loss" in result.history
    
    def test_training_result_to_dict(self, sample_config, sample_train_arrays):
        """Test converting training result to dictionary."""
        train_X, train_y, val_X, val_y, _, _ = sample_train_arrays
        
        sample_config.model.epochs = 1
        trainer = ModelTrainer(sample_config)
        
        result = trainer.train_from_arrays(train_X, train_y, val_X, val_y)
        
        result_dict = result.to_dict()
        
        assert "best_val_loss" in result_dict
        assert "total_epochs" in result_dict
        assert "model_hash" in result_dict


class TestModelEvaluation:
    """Tests for model evaluation."""
    
    def test_evaluator_creation(self):
        """Test creating a model evaluator."""
        evaluator = ModelEvaluator()
        
        assert evaluator is not None
    
    def test_evaluate_predictions(self):
        """Test evaluating predictions directly."""
        evaluator = ModelEvaluator()
        
        # Perfect predictions
        predictions = np.array([50.0, 60.0, 70.0, 80.0, 90.0])
        labels = np.array([50.0, 60.0, 70.0, 80.0, 90.0])
        
        metrics = evaluator.evaluate_predictions(predictions, labels)
        
        assert metrics.mae == 0.0
        assert metrics.mse == 0.0
        assert metrics.accuracy_5 == 1.0
    
    def test_metrics_with_errors(self):
        """Test metrics with prediction errors."""
        evaluator = ModelEvaluator()
        
        predictions = np.array([55.0, 65.0, 75.0, 85.0, 95.0])
        labels = np.array([50.0, 60.0, 70.0, 80.0, 90.0])
        
        metrics = evaluator.evaluate_predictions(predictions, labels)
        
        assert metrics.mae == 5.0
        assert metrics.mse == 25.0
        assert metrics.accuracy_5 == 1.0  # All within 5 points
        assert metrics.accuracy_10 == 1.0
    
    def test_metrics_accuracy_thresholds(self):
        """Test accuracy at different thresholds."""
        evaluator = ModelEvaluator()
        
        predictions = np.array([53.0, 68.0, 85.0, 75.0, 100.0])
        labels = np.array([50.0, 60.0, 70.0, 80.0, 90.0])
        
        metrics = evaluator.evaluate_predictions(predictions, labels)
        
        # Errors: 3, 8, 15, 5, 10
        assert metrics.accuracy_5 == 0.4  # 2 out of 5
        assert metrics.accuracy_10 == 0.8  # 4 out of 5
        assert metrics.accuracy_15 == 1.0  # 5 out of 5
    
    def test_metrics_to_dict(self):
        """Test converting metrics to dictionary."""
        evaluator = ModelEvaluator()
        
        predictions = np.array([50.0, 60.0])
        labels = np.array([55.0, 65.0])
        
        metrics = evaluator.evaluate_predictions(predictions, labels)
        metrics_dict = metrics.to_dict()
        
        assert "mae" in metrics_dict
        assert "rmse" in metrics_dict
        assert "accuracy_5" in metrics_dict
    
    def test_evaluation_report(self):
        """Test generating evaluation report."""
        evaluator = ModelEvaluator()
        
        predictions = np.array([50.0, 60.0, 70.0])
        labels = np.array([52.0, 58.0, 75.0])
        
        metrics = evaluator.evaluate_predictions(predictions, labels)
        
        with tempfile.TemporaryDirectory() as tmpdir:
            report_path = Path(tmpdir) / "report.txt"
            report = evaluator.generate_report(metrics, str(report_path))
            
            assert "EVALUATION REPORT" in report
            assert "MAE" in report
            assert report_path.exists()
    
    def test_metrics_summary(self):
        """Test metrics summary string."""
        metrics = EvaluationMetrics(
            mae=5.0,
            rmse=6.5,
            r2=0.85,
            accuracy_5=0.6,
            accuracy_10=0.8,
            accuracy_20=1.0,
            num_samples=100,
        )
        
        summary = metrics.summary()
        
        assert "5.0" in summary
        assert "0.85" in summary
        assert "100" in summary


@pytest.mark.skipif(not TF_AVAILABLE, reason="TensorFlow not available")
class TestModelExport:
    """Tests for model export."""
    
    def test_exporter_creation(self):
        """Test creating a model exporter."""
        config = ExportConfig()
        exporter = ModelExporter(config)
        
        assert exporter is not None
    
    def test_export_savedmodel(self):
        """Test exporting model as SavedModel."""
        config = ModelConfig(input_dim=64, hidden_layers=[32])
        model = TrustScoreModel(config)
        model.compile()
        
        exporter = ModelExporter()
        
        with tempfile.TemporaryDirectory() as tmpdir:
            result = exporter.export_savedmodel(
                model,
                export_path=tmpdir,
                version="v1.0.0_test",
            )
            
            assert result.success
            assert result.model_path is not None
            assert result.model_hash is not None
            assert Path(result.model_path).exists()
    
    def test_compute_model_hash(self):
        """Test computing model hash."""
        config = ModelConfig(input_dim=64, hidden_layers=[32])
        model = TrustScoreModel(config)
        model.compile()
        
        exporter = ModelExporter()
        
        with tempfile.TemporaryDirectory() as tmpdir:
            result = exporter.export_savedmodel(model, tmpdir, "v_test")
            
            hash1 = exporter.compute_model_hash(result.model_path)
            hash2 = exporter.compute_model_hash(result.model_path)
            
            assert hash1 == hash2  # Same model = same hash
            assert len(hash1) == 64  # SHA256 hex
    
    def test_verify_export(self):
        """Test verifying exported model."""
        config = ModelConfig(input_dim=64, hidden_layers=[32])
        model = TrustScoreModel(config)
        model.compile()
        
        exporter = ModelExporter()
        
        with tempfile.TemporaryDirectory() as tmpdir:
            result = exporter.export_savedmodel(model, tmpdir, "v_test")
            
            is_valid = exporter.verify_export(result.model_path)
            
            assert is_valid is True
    
    def test_export_result_metadata(self):
        """Test export result contains proper metadata."""
        config = ModelConfig(input_dim=64, hidden_layers=[32])
        model = TrustScoreModel(config)
        model.compile()
        
        exporter = ModelExporter()
        
        with tempfile.TemporaryDirectory() as tmpdir:
            result = exporter.export_savedmodel(model, tmpdir, "v1.0.0")
            
            assert result.input_signature is not None
            assert result.output_signature is not None
            assert "shape" in result.input_signature
            assert "dtype" in result.input_signature
    
    def test_go_inference_example(self):
        """Test generating Go inference example."""
        config = ModelConfig(input_dim=64, hidden_layers=[32])
        model = TrustScoreModel(config)
        model.compile()
        
        exporter = ModelExporter()
        
        with tempfile.TemporaryDirectory() as tmpdir:
            result = exporter.export_savedmodel(model, tmpdir, "v1.0.0")
            
            go_code = exporter.get_go_inference_example(result)
            
            assert "package main" in go_code
            assert "LoadTrustScoreModel" in go_code
            assert "tensorflow/go" in go_code


@pytest.mark.skipif(not TF_AVAILABLE, reason="TensorFlow not available")
class TestEndToEnd:
    """End-to-end tests for the training pipeline."""
    
    def test_full_training_pipeline(self, sample_config, sample_train_arrays):
        """Test the complete training pipeline."""
        train_X, train_y, val_X, val_y, test_X, test_y = sample_train_arrays
        
        # Configure for fast testing
        sample_config.model.epochs = 3
        sample_config.model.batch_size = 16
        sample_config.model.early_stopping = False
        
        with tempfile.TemporaryDirectory() as tmpdir:
            sample_config.model.checkpoint_dir = tmpdir
            sample_config.export.export_dir = tmpdir
            
            # Train
            trainer = ModelTrainer(sample_config)
            training_result = trainer.train_from_arrays(
                train_X, train_y, val_X, val_y
            )
            
            assert training_result.model is not None
            
            # Evaluate
            evaluator = ModelEvaluator()
            metrics = evaluator.evaluate(
                training_result.model,
                test_X,
                test_y
            )
            
            assert metrics.num_samples == len(test_y)
            
            # Export
            exporter = ModelExporter(sample_config.export)
            export_result = exporter.export_savedmodel(
                training_result.model,
                sample_config.export.export_dir,
                "v_e2e_test"
            )
            
            assert export_result.success
            assert Path(export_result.model_path).exists()


class TestSyntheticDataset:
    """Tests for synthetic dataset generation."""
    
    def test_generate_synthetic_dataset(self):
        """Test generating synthetic dataset."""
        from ml.training.model.train_and_export import generate_synthetic_dataset
        
        train_X, train_y, val_X, val_y, test_X, test_y = generate_synthetic_dataset(
            num_samples=100,
            feature_dim=128,
            seed=42,
        )
        
        # Check shapes
        assert train_X.shape == (80, 128)  # 80% of 100
        assert len(train_y) == 80
        assert val_X.shape == (10, 128)    # 10% of 100
        assert len(val_y) == 10
        assert test_X.shape == (10, 128)   # 10% of 100
        assert len(test_y) == 10
        
        # Check data types
        assert train_X.dtype == np.float32
        assert train_y.dtype == np.float32
        
        # Check trust score range
        assert np.all(train_y >= 0) and np.all(train_y <= 100)
        assert np.all(val_y >= 0) and np.all(val_y <= 100)
        assert np.all(test_y >= 0) and np.all(test_y <= 100)
    
    def test_synthetic_dataset_reproducibility(self):
        """Test that synthetic dataset is reproducible with same seed."""
        from ml.training.model.train_and_export import generate_synthetic_dataset
        
        # Generate twice with same seed
        data1 = generate_synthetic_dataset(num_samples=50, seed=42)
        data2 = generate_synthetic_dataset(num_samples=50, seed=42)
        
        # Should be identical
        np.testing.assert_array_equal(data1[0], data2[0])  # train_X
        np.testing.assert_array_equal(data1[1], data2[1])  # train_y


class TestModelHash:
    """Tests for model hash computation."""
    
    @pytest.mark.skipif(not TF_AVAILABLE, reason="TensorFlow not available")
    def test_compute_model_hash(self):
        """Test computing model hash."""
        from ml.training.model.train_and_export import compute_model_hash
        
        config = ModelConfig(input_dim=64, hidden_layers=[32])
        model = TrustScoreModel(config)
        model.compile()
        
        exporter = ModelExporter()
        
        with tempfile.TemporaryDirectory() as tmpdir:
            result = exporter.export_savedmodel(model, tmpdir, "v_hash_test")
            
            hash1 = compute_model_hash(result.model_path)
            hash2 = compute_model_hash(result.model_path)
            
            # Same model should give same hash
            assert hash1 == hash2
            
            # Hash should be 64 hex characters (SHA-256)
            assert len(hash1) == 64
            assert all(c in '0123456789abcdef' for c in hash1)
    
    @pytest.mark.skipif(not TF_AVAILABLE, reason="TensorFlow not available")
    def test_save_model_hash(self):
        """Test saving MODEL_HASH.txt."""
        from ml.training.model.train_and_export import compute_model_hash, save_model_hash
        
        config = ModelConfig(input_dim=64, hidden_layers=[32])
        model = TrustScoreModel(config)
        model.compile()
        
        exporter = ModelExporter()
        
        with tempfile.TemporaryDirectory() as tmpdir:
            result = exporter.export_savedmodel(model, tmpdir, "v_hash_save_test")
            
            model_hash = compute_model_hash(result.model_path)
            hash_path = save_model_hash(result.model_path, model_hash, "v1.0.0")
            
            # Check file was created
            assert Path(hash_path).exists()
            
            # Check content
            content = Path(hash_path).read_text()
            assert f"SHA256={model_hash}" in content
            assert "VERSION=v1.0.0" in content


class TestGovernanceWorkflow:
    """Tests for governance workflow generation."""
    
    def test_governance_workflow_checklist(self):
        """Test generating pre-proposal checklist."""
        from ml.training.model.governance import GovernanceWorkflow
        from ml.training.model.manifest import ModelManifest
        
        # Create mock manifest
        manifest = ModelManifest(
            model_version="v1.0.0",
            model_hash="a" * 64,
            evaluation_passed=True,
            tensorflow_version="2.15.0",
            determinism_settings={
                "force_cpu": True,
                "deterministic": True,
                "tf_deterministic_ops": True,
                "random_seed": 42,
            },
        )
        
        metrics = {
            "r2": 0.90,
            "mae": 5.0,
            "num_samples": 500,
        }
        
        workflow = GovernanceWorkflow()
        checklist = workflow.generate_pre_proposal_checklist(manifest, metrics)
        
        assert checklist["model_version"] == "v1.0.0"
        assert len(checklist["checks"]) >= 4
        assert checklist["all_passed"] is True
    
    def test_governance_proposal_generation(self):
        """Test generating governance proposal."""
        from ml.training.model.governance import GovernanceProposalGenerator, GovernanceProposal
        from ml.training.model.manifest import ModelManifest
        
        manifest = ModelManifest(
            model_version="v1.0.0",
            model_hash="b" * 64,
            tensorflow_version="2.15.0",
            export_timestamp=datetime.utcnow().isoformat(),
            evaluation_metrics={
                "mae": 5.0,
                "rmse": 6.5,
                "r2": 0.90,
                "accuracy_5": 0.65,
                "accuracy_10": 0.85,
                "accuracy_20": 0.98,
                "num_samples": 1000,
            },
            determinism_settings={
                "deterministic": True,
                "random_seed": 42,
                "force_cpu": True,
            },
        )
        
        generator = GovernanceProposalGenerator()
        proposal = generator.generate(
            manifest=manifest,
            model_url="ipfs://test-hash",
        )
        
        assert isinstance(proposal, GovernanceProposal)
        assert proposal.model_version == "v1.0.0"
        assert proposal.model_hash == "b" * 64
        assert "Trust Score Model" in proposal.title
        assert "v1.0.0" in proposal.title
    
    def test_validator_notification(self):
        """Test generating validator notification."""
        from ml.training.model.governance import GovernanceWorkflow, GovernanceProposal
        
        proposal = GovernanceProposal(
            title="Update Trust Score Model to v1.0.0",
            model_version="v1.0.0",
            model_hash="c" * 64,
            voting_period="604800s",
            metrics={
                "mae": 5.0,
                "rmse": 6.5,
                "r2": 0.90,
                "accuracy_10": 0.85,
            },
        )
        
        workflow = GovernanceWorkflow()
        notification = workflow.generate_validator_notification(proposal)
        
        assert "MODEL UPDATE NOTIFICATION" in notification
        assert "v1.0.0" in notification
        assert "VALIDATOR ACTIONS REQUIRED" in notification


@pytest.mark.skipif(not TF_AVAILABLE, reason="TensorFlow not available")
class TestTrainAndExportDryRun:
    """Tests for train_and_export dry run functionality."""
    
    def test_dry_run_mode(self):
        """Test dry run mode validates without training."""
        from ml.training.model.train_and_export import train_and_export
        
        with tempfile.TemporaryDirectory() as tmpdir:
            result = train_and_export(
                output_dir=tmpdir,
                synthetic=True,
                num_samples=100,
                dry_run=True,
            )
            
            assert result["success"] is True
            assert result.get("dry_run") is True
