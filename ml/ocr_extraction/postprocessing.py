"""
OCR Post-processing for error correction.

This module provides text correction and normalization
after Tesseract OCR, handling common character confusions
and applying field-specific patterns.
"""

import re
from typing import Dict, List, Optional, Tuple
from dataclasses import dataclass

from ml.ocr_extraction.config import PostProcessingConfig


# Common OCR character confusions (source -> target)
# These are applied in context-sensitive manner
CHARACTER_CONFUSIONS: Dict[str, Dict[str, str]] = {
    # Confusions in numeric context
    "numeric": {
        "O": "0",
        "o": "0",
        "I": "1",
        "l": "1",
        "L": "1",
        "|": "1",
        "S": "5",
        "s": "5",
        "Z": "2",
        "z": "2",
        "B": "8",
        "G": "6",
        "g": "9",
        "q": "9",
        "T": "7",
        "A": "4",
    },
    # Confusions in alphabetic context
    "alpha": {
        "0": "O",
        "1": "I",
        "5": "S",
        "8": "B",
        "6": "G",
    },
    # General confusions (always apply)
    "general": {
        "—": "-",
        "–": "-",
        "−": "-",
        "'": "'",
        "'": "'",
        """: '"',
        """: '"',
        "…": "...",
    },
}


# MRZ-specific corrections
MRZ_CORRECTIONS: Dict[str, str] = {
    " ": "",  # Remove spaces
    "«": "<",
    "»": "<",
    "K": "<",  # Common confusion with filler
}


@dataclass
class CorrectionResult:
    """Result of post-processing correction."""
    original: str
    corrected: str
    corrections_made: List[Tuple[int, str, str]]  # (position, old, new)
    
    @property
    def was_corrected(self) -> bool:
        """Check if any corrections were made."""
        return len(self.corrections_made) > 0


class OCRPostProcessor:
    """
    Post-processor for OCR output.
    
    Handles common OCR errors, whitespace normalization,
    and field-specific text patterns.
    """
    
    def __init__(self, config: Optional[PostProcessingConfig] = None):
        """
        Initialize post-processor.
        
        Args:
            config: Post-processing configuration. Uses defaults if None.
        """
        self.config = config or PostProcessingConfig()
    
    def process(self, text: str, field_type: Optional[str] = None) -> str:
        """
        Apply all post-processing steps to text.
        
        Args:
            text: Raw OCR text
            field_type: Optional field type for specialized processing
            
        Returns:
            Processed text
        """
        result = text
        
        # Apply general corrections first
        result = self._apply_general_corrections(result)
        
        # Normalize whitespace
        if self.config.normalize_whitespace:
            result = self.normalize_whitespace(result)
        
        # Remove newlines
        if self.config.remoVIRTENGINE_newlines:
            result = result.replace("\n", " ").replace("\r", "")
        
        # Strip leading/trailing
        if self.config.strip_leading_trailing:
            result = result.strip()
        
        # Apply field-specific corrections
        if field_type:
            result = self.apply_field_patterns(result, field_type)
        
        # Case normalization
        if self.config.convert_to_uppercase:
            result = result.upper()
        elif self.config.convert_to_lowercase:
            result = result.lower()
        
        # Remove special characters if configured
        if self.config.remoVIRTENGINE_special_chars:
            result = self._remoVIRTENGINE_special_chars(result)
        
        # Apply custom replacements
        for old, new in self.config.custom_replacements.items():
            result = result.replace(old, new)
        
        # Final whitespace normalization
        if self.config.normalize_whitespace:
            result = self.normalize_whitespace(result)
        
        return result
    
    def correct_common_errors(
        self,
        text: str,
        context: str = "general"
    ) -> CorrectionResult:
        """
        Fix common OCR character confusions.
        
        Handles O/0, I/1/l, S/5 and similar confusions
        based on the expected context.
        
        Args:
            text: Text to correct
            context: "numeric", "alpha", or "general"
            
        Returns:
            CorrectionResult with corrections made
        """
        corrections = []
        result = list(text)
        
        # Get confusion map for context
        confusion_map = CHARACTER_CONFUSIONS.get(context, {})
        general_map = CHARACTER_CONFUSIONS["general"]
        
        # Combine maps (context-specific takes precedence)
        combined_map = {**general_map, **confusion_map}
        
        for i, char in enumerate(result):
            if char in combined_map:
                new_char = combined_map[char]
                corrections.append((i, char, new_char))
                result[i] = new_char
        
        return CorrectionResult(
            original=text,
            corrected="".join(result),
            corrections_made=corrections,
        )
    
    def correct_in_context(self, text: str) -> str:
        """
        Apply context-aware corrections.
        
        Analyzes surrounding characters to determine if
        a character is in numeric or alphabetic context.
        
        Args:
            text: Text to correct
            
        Returns:
            Corrected text
        """
        if not self.config.enable_confusion_correction:
            return text
        
        result = list(text)
        
        for i, char in enumerate(result):
            # Determine context from surrounding characters
            prev_char = result[i - 1] if i > 0 else " "
            next_char = result[i + 1] if i < len(result) - 1 else " "
            
            # Check if in numeric context
            prev_numeric = prev_char.isdigit()
            next_numeric = next_char.isdigit()
            
            if prev_numeric or next_numeric:
                # Numeric context
                if char in CHARACTER_CONFUSIONS["numeric"]:
                    result[i] = CHARACTER_CONFUSIONS["numeric"][char]
            else:
                # Alphabetic context
                prev_alpha = prev_char.isalpha()
                next_alpha = next_char.isalpha()
                
                if prev_alpha or next_alpha:
                    if char in CHARACTER_CONFUSIONS["alpha"]:
                        result[i] = CHARACTER_CONFUSIONS["alpha"][char]
        
        return "".join(result)
    
    def normalize_whitespace(self, text: str) -> str:
        """
        Normalize whitespace in text.
        
        Collapses multiple spaces and tabs to single space.
        
        Args:
            text: Text with potentially irregular whitespace
            
        Returns:
            Text with normalized whitespace
        """
        # Replace tabs with spaces
        result = text.replace("\t", " ")
        
        # Collapse multiple spaces
        result = re.sub(r" +", " ", result)
        
        return result
    
    def apply_field_patterns(self, text: str, field_type: str) -> str:
        """
        Apply field-specific regex patterns and corrections.
        
        Args:
            text: Input text
            field_type: Type of field (e.g., "date", "id_number", "name")
            
        Returns:
            Corrected text
        """
        handlers = {
            "date": self._process_date,
            "date_of_birth": self._process_date,
            "expiry_date": self._process_date,
            "issue_date": self._process_date,
            "id_number": self._process_id_number,
            "document_number": self._process_id_number,
            "name": self._process_name,
            "full_name": self._process_name,
            "first_name": self._process_name,
            "last_name": self._process_name,
            "surname": self._process_name,
            "given_names": self._process_name,
            "mrz": self._process_mrz,
            "mrz_line1": self._process_mrz,
            "mrz_line2": self._process_mrz,
        }
        
        handler = handlers.get(field_type, lambda x: x)
        return handler(text)
    
    def _process_date(self, text: str) -> str:
        """Process date field."""
        # Correct numeric characters
        result = self.correct_common_errors(text, "numeric")
        text = result.corrected
        
        # Remove non-date characters
        text = re.sub(r"[^\d/\-.]", "", text)
        
        # Standardize separators
        text = text.replace(".", "/").replace("-", "/")
        
        # Fix common patterns
        # DD/MM/YYYY or MM/DD/YYYY
        match = re.search(r"(\d{1,2})/(\d{1,2})/(\d{2,4})", text)
        if match:
            d, m, y = match.groups()
            d = d.zfill(2)
            m = m.zfill(2)
            if len(y) == 2:
                # Assume 20xx for years < 50, 19xx otherwise
                y = f"20{y}" if int(y) < 50 else f"19{y}"
            text = f"{d}/{m}/{y}"
        
        return text
    
    def _process_id_number(self, text: str) -> str:
        """Process ID number field."""
        # Remove non-alphanumeric except hyphen
        text = re.sub(r"[^A-Za-z0-9\-]", "", text)
        
        # Uppercase
        text = text.upper()
        
        return text
    
    def _process_name(self, text: str) -> str:
        """Process name field."""
        # Remove digits and most special characters
        text = re.sub(r"[^A-Za-z\s\-']", "", text)
        
        # Normalize whitespace
        text = self.normalize_whitespace(text)
        
        # Title case
        text = text.title()
        
        # Fix common OCR issues with names
        # McX -> McX (preserve case after Mc)
        text = re.sub(r"\bMc([a-z])", lambda m: f"Mc{m.group(1).upper()}", text)
        
        return text
    
    def _process_mrz(self, text: str) -> str:
        """Process MRZ line."""
        # Apply MRZ-specific corrections
        for old, new in MRZ_CORRECTIONS.items():
            text = text.replace(old, new)
        
        # Only allow MRZ characters
        text = re.sub(r"[^A-Z0-9<]", "", text.upper())
        
        return text
    
    def _apply_general_corrections(self, text: str) -> str:
        """Apply general character corrections."""
        for old, new in CHARACTER_CONFUSIONS["general"].items():
            text = text.replace(old, new)
        return text
    
    def _remoVIRTENGINE_special_chars(self, text: str) -> str:
        """Remove special characters except allowed ones."""
        allowed = self.config.allowed_special_chars
        result = []
        
        for char in text:
            if char.isalnum() or char.isspace() or char in allowed:
                result.append(char)
        
        return "".join(result)
    
    def extract_numbers(self, text: str) -> str:
        """Extract only numeric characters."""
        return "".join(c for c in text if c.isdigit())
    
    def extract_letters(self, text: str) -> str:
        """Extract only alphabetic characters."""
        return "".join(c for c in text if c.isalpha())
    
    def split_by_pattern(
        self,
        text: str,
        pattern: str
    ) -> List[str]:
        """
        Split text by regex pattern.
        
        Args:
            text: Text to split
            pattern: Regex pattern
            
        Returns:
            List of matched groups
        """
        matches = re.findall(pattern, text)
        return matches if matches else [text]


def create_name_corrector() -> OCRPostProcessor:
    """Create post-processor optimized for names."""
    config = PostProcessingConfig(
        enable_confusion_correction=True,
        normalize_whitespace=True,
        strip_leading_trailing=True,
        remoVIRTENGINE_newlines=True,
        convert_to_uppercase=False,
    )
    return OCRPostProcessor(config)


def create_id_corrector() -> OCRPostProcessor:
    """Create post-processor optimized for ID numbers."""
    config = PostProcessingConfig(
        enable_confusion_correction=True,
        normalize_whitespace=True,
        strip_leading_trailing=True,
        remoVIRTENGINE_newlines=True,
        convert_to_uppercase=True,
        remoVIRTENGINE_special_chars=True,
        allowed_special_chars="-",
    )
    return OCRPostProcessor(config)


def create_mrz_corrector() -> OCRPostProcessor:
    """Create post-processor optimized for MRZ."""
    config = PostProcessingConfig(
        enable_confusion_correction=False,  # MRZ has its own corrections
        normalize_whitespace=False,  # Preserve structure
        strip_leading_trailing=True,
        remoVIRTENGINE_newlines=True,
        convert_to_uppercase=True,
    )
    return OCRPostProcessor(config)
