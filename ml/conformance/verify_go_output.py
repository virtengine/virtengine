"""
Verify Go inference outputs match Python.

This script validates that Go inference produces identical outputs
to Python for the conformance test vectors.

Task Reference: VE-3006 - Go-Python Conformance Testing

Usage:
    python ml/conformance/verify_go_output.py go_outputs.json
    python -m ml.conformance.verify_go_output go_outputs.json
"""

import argparse
import json
import sys
from dataclasses import dataclass
from pathlib import Path
from typing import List, Dict, Any, Optional, Tuple

# Tolerance for floating-point comparison
DEFAULT_TOLERANCE = 1e-6


@dataclass
class VerificationResult:
    """Result of verifying a single test vector."""
    vector_name: str
    passed: bool
    score_match: bool
    tier_match: bool
    codes_match: bool
    confidence_match: bool
    score_diff: float
    confidence_diff: float
    missing_codes: List[str]
    extra_codes: List[str]
    error_message: Optional[str] = None


def load_test_vectors(path: Optional[str] = None) -> Dict[str, Any]:
    """Load test vectors from JSON file."""
    if path is None:
        path = Path(__file__).parent / "test_vectors.json"
    else:
        path = Path(path)
    
    if not path.exists():
        raise FileNotFoundError(f"Test vectors not found: {path}")
    
    with open(path, "r") as f:
        return json.load(f)


def load_go_outputs(path: str) -> Dict[str, Any]:
    """Load Go inference outputs from JSON file."""
    path = Path(path)
    
    if not path.exists():
        raise FileNotFoundError(f"Go outputs not found: {path}")
    
    with open(path, "r") as f:
        return json.load(f)


def floats_equal(a: float, b: float, tolerance: float = DEFAULT_TOLERANCE) -> bool:
    """Compare two floats with tolerance."""
    return abs(a - b) <= tolerance


def verify_single_vector(
    vector: Dict[str, Any],
    go_output: Dict[str, Any],
    tolerance: float = DEFAULT_TOLERANCE
) -> VerificationResult:
    """
    Verify a single test vector against Go output.
    
    Args:
        vector: Expected test vector from Python
        go_output: Actual output from Go inference
        tolerance: Maximum allowed floating-point difference
        
    Returns:
        VerificationResult with detailed comparison
    """
    vector_name = vector.get("name", "unknown")
    
    try:
        # Compare scores
        expected_score = vector.get("expected_score", 0.0)
        actual_score = go_output.get("score", 0.0)
        score_diff = abs(expected_score - actual_score)
        score_match = floats_equal(expected_score, actual_score, tolerance)
        
        # Compare tiers
        expected_tier = vector.get("expected_tier", 0)
        actual_tier = go_output.get("tier", 0)
        tier_match = expected_tier == actual_tier
        
        # Compare confidence
        expected_confidence = vector.get("expected_confidence", 0.0)
        actual_confidence = go_output.get("confidence", 0.0)
        confidence_diff = abs(expected_confidence - actual_confidence)
        confidence_match = floats_equal(expected_confidence, actual_confidence, tolerance)
        
        # Compare reason codes
        expected_codes = set(vector.get("expected_codes", []))
        actual_codes = set(go_output.get("reason_codes", []))
        missing_codes = list(expected_codes - actual_codes)
        extra_codes = list(actual_codes - expected_codes)
        codes_match = len(missing_codes) == 0 and len(extra_codes) == 0
        
        # Overall pass/fail
        passed = score_match and tier_match and codes_match and confidence_match
        
        return VerificationResult(
            vector_name=vector_name,
            passed=passed,
            score_match=score_match,
            tier_match=tier_match,
            codes_match=codes_match,
            confidence_match=confidence_match,
            score_diff=score_diff,
            confidence_diff=confidence_diff,
            missing_codes=missing_codes,
            extra_codes=extra_codes,
        )
        
    except Exception as e:
        return VerificationResult(
            vector_name=vector_name,
            passed=False,
            score_match=False,
            tier_match=False,
            codes_match=False,
            confidence_match=False,
            score_diff=0.0,
            confidence_diff=0.0,
            missing_codes=[],
            extra_codes=[],
            error_message=str(e),
        )


def verify_outputs(
    go_outputs_path: str,
    test_vectors_path: Optional[str] = None,
    tolerance: float = DEFAULT_TOLERANCE,
    verbose: bool = False,
) -> Tuple[bool, List[VerificationResult]]:
    """
    Verify Go outputs match expected Python outputs.
    
    Args:
        go_outputs_path: Path to JSON file with Go inference outputs
        test_vectors_path: Optional path to test vectors (uses default if None)
        tolerance: Maximum allowed floating-point difference
        verbose: Print detailed results
        
    Returns:
        Tuple of (all_passed, list of results)
    """
    # Load data
    test_vectors_data = load_test_vectors(test_vectors_path)
    go_outputs_data = load_go_outputs(go_outputs_path)
    
    # Extract vectors
    vectors = test_vectors_data.get("vectors", [])
    go_results = go_outputs_data.get("results", [])
    
    # Create lookup by name
    go_results_by_name = {r.get("name", ""): r for r in go_results}
    
    results = []
    
    for vector in vectors:
        vector_name = vector.get("name", "")
        go_output = go_results_by_name.get(vector_name, {})
        
        if not go_output:
            results.append(VerificationResult(
                vector_name=vector_name,
                passed=False,
                score_match=False,
                tier_match=False,
                codes_match=False,
                confidence_match=False,
                score_diff=0.0,
                confidence_diff=0.0,
                missing_codes=[],
                extra_codes=[],
                error_message=f"No Go output found for vector: {vector_name}",
            ))
            continue
        
        result = verify_single_vector(vector, go_output, tolerance)
        results.append(result)
        
        if verbose:
            status = "✓ PASS" if result.passed else "✗ FAIL"
            print(f"  {status}: {vector_name}")
            
            if not result.passed:
                if not result.score_match:
                    print(f"      Score: expected vs actual, diff={result.score_diff:.6f}")
                if not result.tier_match:
                    print(f"      Tier mismatch")
                if not result.confidence_match:
                    print(f"      Confidence: diff={result.confidence_diff:.6f}")
                if result.missing_codes:
                    print(f"      Missing codes: {result.missing_codes}")
                if result.extra_codes:
                    print(f"      Extra codes: {result.extra_codes}")
                if result.error_message:
                    print(f"      Error: {result.error_message}")
    
    all_passed = all(r.passed for r in results)
    return all_passed, results


def generate_report(
    results: List[VerificationResult],
    output_path: Optional[str] = None
) -> str:
    """
    Generate a verification report.
    
    Args:
        results: List of verification results
        output_path: Optional path to write report
        
    Returns:
        Report as string
    """
    passed_count = sum(1 for r in results if r.passed)
    failed_count = len(results) - passed_count
    
    lines = [
        "=" * 60,
        "Go-Python Conformance Verification Report",
        "=" * 60,
        "",
        f"Total vectors: {len(results)}",
        f"Passed: {passed_count}",
        f"Failed: {failed_count}",
        "",
        "-" * 60,
        "Results:",
        "-" * 60,
    ]
    
    for result in results:
        status = "PASS" if result.passed else "FAIL"
        lines.append(f"\n{status}: {result.vector_name}")
        
        if not result.passed:
            if not result.score_match:
                lines.append(f"  - Score difference: {result.score_diff:.8f}")
            if not result.tier_match:
                lines.append(f"  - Tier mismatch")
            if not result.confidence_match:
                lines.append(f"  - Confidence difference: {result.confidence_diff:.8f}")
            if result.missing_codes:
                lines.append(f"  - Missing reason codes: {', '.join(result.missing_codes)}")
            if result.extra_codes:
                lines.append(f"  - Extra reason codes: {', '.join(result.extra_codes)}")
            if result.error_message:
                lines.append(f"  - Error: {result.error_message}")
    
    lines.extend([
        "",
        "-" * 60,
        f"Result: {'ALL PASSED' if failed_count == 0 else 'FAILURES DETECTED'}",
        "=" * 60,
    ])
    
    report = "\n".join(lines)
    
    if output_path:
        Path(output_path).write_text(report)
    
    return report


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description="Verify Go inference outputs match Python conformance test vectors"
    )
    parser.add_argument(
        "go_outputs",
        help="Path to JSON file with Go inference outputs"
    )
    parser.add_argument(
        "--test-vectors",
        help="Path to test vectors JSON (uses default if not specified)"
    )
    parser.add_argument(
        "--tolerance",
        type=float,
        default=DEFAULT_TOLERANCE,
        help=f"Floating-point comparison tolerance (default: {DEFAULT_TOLERANCE})"
    )
    parser.add_argument(
        "--report",
        help="Path to write verification report"
    )
    parser.add_argument(
        "-v", "--verbose",
        action="store_true",
        help="Print detailed results"
    )
    
    args = parser.parse_args()
    
    print("=" * 60)
    print("VE-3006: Go-Python Conformance Verification")
    print("=" * 60)
    print()
    
    try:
        all_passed, results = verify_outputs(
            go_outputs_path=args.go_outputs,
            test_vectors_path=args.test_vectors,
            tolerance=args.tolerance,
            verbose=args.verbose,
        )
        
        # Generate and print report
        report = generate_report(results, args.report)
        print(report)
        
        # Exit with appropriate code
        sys.exit(0 if all_passed else 1)
        
    except FileNotFoundError as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(2)
    except json.JSONDecodeError as e:
        print(f"Error parsing JSON: {e}", file=sys.stderr)
        sys.exit(2)
    except Exception as e:
        print(f"Unexpected error: {e}", file=sys.stderr)
        sys.exit(2)


if __name__ == "__main__":
    main()
