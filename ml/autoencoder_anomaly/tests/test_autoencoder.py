"""
Tests for autoencoder model.

VE-924: Test encoder/decoder architecture.
"""

import pytest
import numpy as np

from ml.autoencoder_anomaly.autoencoder import (
    Autoencoder,
    AutoencoderOutput,
    ConvolutionalEncoder,
    ConvolutionalDecoder,
    EncoderOutput,
    DecoderOutput,
    create_autoencoder,
)
from ml.autoencoder_anomaly.config import (
    AutoencoderAnomalyConfig,
    EncoderConfig,
    DecoderConfig,
)


class TestConvolutionalEncoder:
    """Tests for ConvolutionalEncoder class."""
    
    def test_encoder_creation(self, encoder_config):
        """Test creating encoder with config."""
        encoder = ConvolutionalEncoder(encoder_config)
        
        assert encoder is not None
        assert encoder.config == encoder_config
        assert encoder.MODEL_VERSION == "1.0.0"
    
    def test_encoder_default_config(self):
        """Test creating encoder with default config."""
        encoder = ConvolutionalEncoder()
        assert encoder is not None
    
    def test_model_hash_generation(self, encoder_config):
        """Test model hash is generated."""
        encoder = ConvolutionalEncoder(encoder_config)
        
        assert encoder.model_hash is not None
        assert len(encoder.model_hash) == 32
    
    def test_model_hash_determinism(self, encoder_config):
        """Test model hash is deterministic."""
        encoder1 = ConvolutionalEncoder(encoder_config)
        encoder2 = ConvolutionalEncoder(encoder_config)
        
        assert encoder1.model_hash == encoder2.model_hash
    
    def test_encode_image(self, encoder_config, sample_image):
        """Test encoding an image."""
        encoder = ConvolutionalEncoder(encoder_config)
        
        output = encoder.encode(sample_image)
        
        assert isinstance(output, EncoderOutput)
        assert output.latent_vector is not None
        assert len(output.latent_vector) == encoder_config.latent_dim
    
    def test_encode_with_features(self, encoder_config, sample_image):
        """Test encoding with intermediate features."""
        encoder = ConvolutionalEncoder(encoder_config)
        
        output = encoder.encode(sample_image, return_features=True)
        
        assert len(output.layer_features) > 0
    
    def test_encode_determinism(self, encoder_config, sample_image):
        """Test encoding is deterministic."""
        encoder1 = ConvolutionalEncoder(encoder_config)
        encoder2 = ConvolutionalEncoder(encoder_config)
        
        output1 = encoder1.encode(sample_image)
        output2 = encoder2.encode(sample_image)
        
        np.testing.assert_array_almost_equal(
            output1.latent_vector,
            output2.latent_vector,
            decimal=5
        )
    
    def test_latent_statistics(self, encoder_config, sample_image):
        """Test latent statistics are computed."""
        encoder = ConvolutionalEncoder(encoder_config)
        
        output = encoder.encode(sample_image)
        
        assert output.latent_mean is not None
        assert output.latent_std is not None
        assert output.latent_min is not None
        assert output.latent_max is not None
    
    def test_feature_hash(self, encoder_config, sample_image):
        """Test feature hash is generated."""
        encoder = ConvolutionalEncoder(encoder_config)
        
        output = encoder.encode(sample_image)
        
        assert output.feature_hash is not None
        assert len(output.feature_hash) == 16
    
    def test_output_to_dict(self, encoder_config, sample_image):
        """Test EncoderOutput.to_dict()."""
        encoder = ConvolutionalEncoder(encoder_config)
        
        output = encoder.encode(sample_image)
        output_dict = output.to_dict()
        
        assert "latent_dim" in output_dict
        assert "latent_mean" in output_dict
        assert "feature_hash" in output_dict


class TestConvolutionalDecoder:
    """Tests for ConvolutionalDecoder class."""
    
    def test_decoder_creation(self, decoder_config):
        """Test creating decoder with config."""
        decoder = ConvolutionalDecoder(
            config=decoder_config,
            latent_dim=128,
            spatial_size=(8, 8)
        )
        
        assert decoder is not None
        assert decoder.config == decoder_config
        assert decoder.MODEL_VERSION == "1.0.0"
    
    def test_decoder_default_config(self):
        """Test creating decoder with default config."""
        decoder = ConvolutionalDecoder(latent_dim=128, spatial_size=(8, 8))
        assert decoder is not None
    
    def test_model_hash_generation(self, decoder_config):
        """Test model hash is generated."""
        decoder = ConvolutionalDecoder(
            config=decoder_config,
            latent_dim=128,
            spatial_size=(8, 8)
        )
        
        assert decoder.model_hash is not None
        assert len(decoder.model_hash) == 32
    
    def test_decode_latent(self, decoder_config, sample_latent_vector):
        """Test decoding a latent vector."""
        decoder = ConvolutionalDecoder(
            config=decoder_config,
            latent_dim=len(sample_latent_vector),
            spatial_size=(8, 8)
        )
        
        output = decoder.decode(sample_latent_vector)
        
        assert isinstance(output, DecoderOutput)
        assert output.reconstruction is not None
        assert output.reconstruction.shape[-1] == decoder_config.output_channels
    
    def test_decode_determinism(self, decoder_config, sample_latent_vector):
        """Test decoding is deterministic."""
        decoder1 = ConvolutionalDecoder(
            config=decoder_config,
            latent_dim=len(sample_latent_vector),
            spatial_size=(8, 8)
        )
        decoder2 = ConvolutionalDecoder(
            config=decoder_config,
            latent_dim=len(sample_latent_vector),
            spatial_size=(8, 8)
        )
        
        output1 = decoder1.decode(sample_latent_vector)
        output2 = decoder2.decode(sample_latent_vector)
        
        np.testing.assert_array_almost_equal(
            output1.reconstruction,
            output2.reconstruction,
            decimal=5
        )
    
    def test_output_normalized(self, decoder_config, sample_latent_vector):
        """Test output is in [0, 1] range (sigmoid)."""
        decoder = ConvolutionalDecoder(
            config=decoder_config,
            latent_dim=len(sample_latent_vector),
            spatial_size=(8, 8)
        )
        
        output = decoder.decode(sample_latent_vector)
        
        assert output.reconstruction.min() >= 0.0
        assert output.reconstruction.max() <= 1.0
    
    def test_output_to_dict(self, decoder_config, sample_latent_vector):
        """Test DecoderOutput.to_dict()."""
        decoder = ConvolutionalDecoder(
            config=decoder_config,
            latent_dim=len(sample_latent_vector),
            spatial_size=(8, 8)
        )
        
        output = decoder.decode(sample_latent_vector)
        output_dict = output.to_dict()
        
        assert "reconstruction_shape" in output_dict
        assert "processing_time_ms" in output_dict


class TestAutoencoder:
    """Tests for complete Autoencoder class."""
    
    def test_autoencoder_creation(self, anomaly_config):
        """Test creating autoencoder with config."""
        autoencoder = Autoencoder(anomaly_config)
        
        assert autoencoder is not None
        assert autoencoder.config == anomaly_config
        assert autoencoder.MODEL_VERSION == "1.0.0"
    
    def test_autoencoder_default_config(self):
        """Test creating autoencoder with default config."""
        autoencoder = Autoencoder()
        assert autoencoder is not None
    
    def test_model_hash_generation(self, anomaly_config):
        """Test combined model hash is generated."""
        autoencoder = Autoencoder(anomaly_config)
        
        assert autoencoder.model_hash is not None
        assert len(autoencoder.model_hash) == 32
    
    def test_model_hash_determinism(self, anomaly_config):
        """Test model hash is deterministic."""
        ae1 = Autoencoder(anomaly_config)
        ae2 = Autoencoder(anomaly_config)
        
        assert ae1.model_hash == ae2.model_hash
    
    def test_forward_pass(self, anomaly_config, sample_image):
        """Test forward pass through autoencoder."""
        autoencoder = Autoencoder(anomaly_config)
        
        output = autoencoder.forward(sample_image)
        
        assert isinstance(output, AutoencoderOutput)
        assert output.encoder_output is not None
        assert output.decoder_output is not None
    
    def test_forward_with_features(self, anomaly_config, sample_image):
        """Test forward pass with features."""
        autoencoder = Autoencoder(anomaly_config)
        
        output = autoencoder.forward(sample_image, return_features=True)
        
        assert len(output.encoder_output.layer_features) > 0
    
    def test_encode_method(self, anomaly_config, sample_image):
        """Test encode method."""
        autoencoder = Autoencoder(anomaly_config)
        
        encoder_output = autoencoder.encode(sample_image)
        
        assert isinstance(encoder_output, EncoderOutput)
        assert len(encoder_output.latent_vector) == anomaly_config.encoder.latent_dim
    
    def test_decode_method(self, anomaly_config, sample_latent_vector):
        """Test decode method."""
        autoencoder = Autoencoder(anomaly_config)
        
        decoder_output = autoencoder.decode(sample_latent_vector)
        
        assert isinstance(decoder_output, DecoderOutput)
        assert decoder_output.reconstruction is not None
    
    def test_output_to_dict(self, anomaly_config, sample_image):
        """Test AutoencoderOutput.to_dict()."""
        autoencoder = Autoencoder(anomaly_config)
        
        output = autoencoder.forward(sample_image)
        output_dict = output.to_dict()
        
        assert "encoder" in output_dict
        assert "decoder" in output_dict
        assert "model_version" in output_dict
        assert "model_hash" in output_dict
    
    def test_timing_recorded(self, anomaly_config, sample_image):
        """Test processing time is recorded."""
        autoencoder = Autoencoder(anomaly_config)
        
        output = autoencoder.forward(sample_image)
        
        assert output.total_time_ms > 0


class TestCreateAutoencoder:
    """Tests for factory function."""
    
    def test_create_with_config(self, anomaly_config):
        """Test creating autoencoder with config."""
        autoencoder = create_autoencoder(anomaly_config)
        
        assert isinstance(autoencoder, Autoencoder)
        assert autoencoder.config == anomaly_config
    
    def test_create_without_config(self):
        """Test creating autoencoder without config."""
        autoencoder = create_autoencoder()
        
        assert isinstance(autoencoder, Autoencoder)
