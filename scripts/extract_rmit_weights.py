#!/usr/bin/env python3
"""
Extract and register U-Net weights from RMIT intern project.

This script copies the trained ResNet34 U-Net weights from the RMIT
OCR_Document_Scan intern project to the VirtEngine ML weights directory.

Usage:
    python scripts/extract_rmit_weights.py

Task Reference: VE-3040 - Extract RMIT U-Net ResNet34 Weights
"""

import hashlib
import shutil
import sys
from pathlib import Path


def calculate_sha256(filepath: Path) -> str:
    """Calculate SHA256 hash of a file."""
    sha256_hash = hashlib.sha256()
    with open(filepath, "rb") as f:
        for byte_block in iter(lambda: f.read(65536), b""):
            sha256_hash.update(byte_block)
    return sha256_hash.hexdigest().lower()


def main() -> int:
    """Main extraction function."""
    # Get project root (parent of scripts/)
    script_dir = Path(__file__).resolve().parent
    project_root = script_dir.parent
    
    # Source path - the actual model file location
    # Note: The model is UNet_sig.pth in resnet34 subfolder, not unet_resnet34_new.pt
    src = project_root / "temp" / "OCR_Document_Scan-main" / "model" / "resnet34" / "UNet_sig.pth"
    
    # Alternative source path (if renamed)
    alt_src = project_root / "temp" / "OCR_Document_Scan-main" / "model" / "unet_resnet34_new.pt"
    
    # Destination
    dst_dir = project_root / "ml" / "face_extraction" / "weights"
    dst = dst_dir / "unet_resnet34.pt"
    
    print("=" * 60)
    print("VE-3040: Extract RMIT U-Net ResNet34 Weights")
    print("=" * 60)
    print()
    
    # Find source file
    source_file = None
    if src.exists():
        source_file = src
        print(f"Found model at: {src}")
    elif alt_src.exists():
        source_file = alt_src
        print(f"Found model at: {alt_src}")
    else:
        print(f"ERROR: Source model not found!")
        print(f"  Checked: {src}")
        print(f"  Checked: {alt_src}")
        print()
        print("Please ensure the RMIT intern project is in temp/OCR_Document_Scan-main/")
        return 1
    
    # Create destination directory
    dst_dir.mkdir(parents=True, exist_ok=True)
    print(f"Destination: {dst}")
    print()
    
    # Calculate source hash before copy
    print("Calculating source file hash...")
    src_hash = calculate_sha256(source_file)
    print(f"  Source SHA256: {src_hash}")
    
    # Get file size
    size_bytes = source_file.stat().st_size
    size_mb = size_bytes / (1024 * 1024)
    print(f"  Size: {size_mb:.2f} MB ({size_bytes:,} bytes)")
    print()
    
    # Copy the model
    print("Copying model...")
    shutil.copy2(source_file, dst)
    print(f"  Copied to: {dst}")
    
    # Verify copy
    print()
    print("Verifying copy...")
    dst_hash = calculate_sha256(dst)
    print(f"  Destination SHA256: {dst_hash}")
    
    if src_hash != dst_hash:
        print()
        print("ERROR: Hash mismatch after copy! File may be corrupted.")
        return 1
    
    print("  ✓ Hash verified successfully")
    print()
    
    # Expected hash from model_registry.py
    expected_hash = "0ea89b9d7b249d04ebe767cc38e78d04545067eabdda0f9fc22a1ae2c19bca57"
    
    if dst_hash == expected_hash:
        print("✓ Hash matches model_registry.py - no update needed")
    else:
        print("!" * 60)
        print("IMPORTANT: Update ml/face_extraction/model_registry.py with:")
        print()
        print(f'        "sha256": "{dst_hash}",')
        print()
        print("!" * 60)
    
    print()
    print("=" * 60)
    print("Extraction complete!")
    print("=" * 60)
    print()
    print("Model details:")
    print(f"  Source: RMIT OCR_Document_Scan intern project (2023)")
    print(f"  Architecture: ResNet34 backbone U-Net")
    print(f"  Framework: PyTorch")
    print(f"  Training data: Turkish ID documents")
    print(f"  SHA256: {dst_hash}")
    print()
    
    return 0


if __name__ == "__main__":
    sys.exit(main())
