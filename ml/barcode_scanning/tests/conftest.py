"""
Test fixtures and configuration for barcode scanning tests.
"""

import pytest
import numpy as np


# Sample MRZ data for testing (fictional/synthetic data)
SAMPLE_MRZ_TD3 = [
    "P<UTOERIKSSON<<ANNA<MARIA<<<<<<<<<<<<<<<<<<<",
    "L898902C36UTO7408122F1204159ZE184226B<<<<<10",
]

SAMPLE_MRZ_TD1 = [
    "I<UTOD231458907<<<<<<<<<<<<<<<",
    "7408122F1204159UTO<<<<<<<<<<<6",
    "ERIKSSON<<ANNA<MARIA<<<<<<<<<<",
]

# Sample PDF417 data (AAMVA format - fictional)
SAMPLE_PDF417_DATA = b"""@

ANSI 6360010102DL00410230ZN02490101DLDAQD12345678
DAAJOHN
DCSSMITH
DCTJOHN
DDEN
DDFN
DDGN
DBB01011990
DBC1
DAU070 in
DAYBRO
DAG123 MAIN STREET
DAICITYVILLE
DAJNY
DAK12345
DCF1234567890
DCGUSA
DBA01012030
DCS
"""


@pytest.fixture
def sample_mrz_td3():
    """Sample TD3 MRZ (passport format)."""
    return "\n".join(SAMPLE_MRZ_TD3)


@pytest.fixture
def sample_mrz_td1():
    """Sample TD1 MRZ (ID card format)."""
    return "\n".join(SAMPLE_MRZ_TD1)


@pytest.fixture
def sample_pdf417_bytes():
    """Sample PDF417 barcode data bytes."""
    return SAMPLE_PDF417_DATA


@pytest.fixture
def sample_grayscale_image():
    """Create a sample grayscale image for testing."""
    return np.zeros((480, 640), dtype=np.uint8)


@pytest.fixture
def sample_color_image():
    """Create a sample color image for testing."""
    return np.zeros((480, 640, 3), dtype=np.uint8)


@pytest.fixture
def sample_ocr_data():
    """Sample OCR extraction data for cross-validation."""
    return {
        "full_name": "ANNA MARIA ERIKSSON",
        "surname": "ERIKSSON",
        "given_names": "ANNA MARIA",
        "date_of_birth": "1974-08-12",
        "document_number": "L898902C3",
        "expiry_date": "2012-04-15",
        "nationality": "UTO",
        "sex": "F",
    }


@pytest.fixture
def sample_ocr_data_mismatch():
    """Sample OCR data with intentional mismatches."""
    return {
        "full_name": "ANA MARIA ERIKSON",  # Typos
        "surname": "ERIKSON",
        "given_names": "ANA MARIA",
        "date_of_birth": "1974-08-12",  # Correct
        "document_number": "L898902C4",  # Wrong digit
        "expiry_date": "2012-04-15",
        "nationality": "UTO",
        "sex": "F",
    }
