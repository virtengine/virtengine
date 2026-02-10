"""
Tests for barcode scanning module initialization.
"""


def test_import_module():
    """Test that the module can be imported."""
    from ml import barcode_scanning
    assert hasattr(barcode_scanning, "__version__")


def test_import_pipeline():
    """Test pipeline imports."""
    from ml.barcode_scanning import (
        BarcodeScanningPipeline,
        BarcodeScanningResult,
        BarcodeScanningConfig,
    )
    
    assert BarcodeScanningPipeline is not None
    assert BarcodeScanningResult is not None
    assert BarcodeScanningConfig is not None


def test_import_pdf417():
    """Test PDF417 parser imports."""
    from ml.barcode_scanning import (
        PDF417Parser,
        PDF417Data,
        AAMVAField,
    )
    
    assert PDF417Parser is not None
    assert PDF417Data is not None
    assert AAMVAField is not None


def test_import_mrz():
    """Test MRZ parser imports."""
    from ml.barcode_scanning import (
        MRZParser,
        MRZData,
        MRZLine,
        MRZCheckDigit,
    )
    
    assert MRZParser is not None
    assert MRZData is not None
    assert MRZLine is not None
    assert MRZCheckDigit is not None


def test_import_cross_validator():
    """Test cross-validator imports."""
    from ml.barcode_scanning import (
        CrossValidator,
        CrossValidationResult,
        FieldMatch,
        MatchType,
    )
    
    assert CrossValidator is not None
    assert CrossValidationResult is not None
    assert FieldMatch is not None
    assert MatchType is not None


def test_import_config():
    """Test config imports."""
    from ml.barcode_scanning import (
        BarcodeType,
        MRZFormat,
        PDF417Config,
        MRZConfig,
        CrossValidationConfig,
    )
    
    assert BarcodeType is not None
    assert MRZFormat is not None
    assert PDF417Config is not None
    assert MRZConfig is not None
    assert CrossValidationConfig is not None
