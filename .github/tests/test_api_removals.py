"""Tests for .github/scripts/api_removals.py.

The regression case is a real unified diff captured from PR #1120
(run 36825740607): a protobuf response-type split that the previous
implementation reported as 15 API removals when it was 0.

Run:  python -m pytest .github/tests/test_api_removals.py -v
   or: python .github/tests/test_api_removals.py
"""

from __future__ import annotations

import importlib.util
import io
import os
import sys
import contextlib
import unittest

_HERE = os.path.dirname(os.path.abspath(__file__))
_SCRIPT = os.path.join(_HERE, "..", "scripts", "api_removals.py")

spec = importlib.util.spec_from_file_location("api_removals", _SCRIPT)
api_removals = importlib.util.module_from_spec(spec)
assert spec.loader is not None
spec.loader.exec_module(api_removals)


def run(diff: str) -> tuple[int, str]:
    """Run main() against `diff` captured on stdin, returning (exit, stdout)."""
    old = sys.stdin
    sys.stdin = io.StringIO(diff)
    buf = io.StringIO()
    try:
        with contextlib.redirect_stdout(buf):
            code = api_removals.main([])
    finally:
        sys.stdin = old
    return code, buf.getvalue()


# The real diff from PR #1120, abridged but structurally exact: each removed
# line has a same-named replacement that differs only in its return type.
PR_1120_DIFF = """diff --git a/x/resources/keeper/grpc_query.go b/x/resources/keeper/grpc_query.go
index 1111111..2222222 100644
--- a/x/resources/keeper/grpc_query.go
+++ b/x/resources/keeper/grpc_query.go
@@ -10,3 +10,3 @@
-func (q *Querier) ReservationByOrder(ctx context.Context, req *resourcesv1.QueryReservationByOrderRequest) (*resourcesv1.QueryReservationResponse, error) {
+func (q *Querier) ReservationByOrder(ctx context.Context, req *resourcesv1.QueryReservationByOrderRequest) (*resourcesv1.QueryReservationByOrderResponse, error) {
@@ -20,3 +20,3 @@
-func (q *Querier) ReservationByBid(ctx context.Context, req *resourcesv1.QueryReservationByBidRequest) (*resourcesv1.QueryReservationResponse, error) {
+func (q *Querier) ReservationByBid(ctx context.Context, req *resourcesv1.QueryReservationByBidRequest) (*resourcesv1.QueryReservationByBidResponse, error) {
"""


class TestSignatureChangeIsNotARemoval(unittest.TestCase):
    def test_pr1120_real_diff_does_not_fail_the_gate(self):
        code, out = run(PR_1120_DIFF)
        self.assertEqual(code, 0, f"gate failed on a pure signature change:\n{out}")
        self.assertNotIn("Potential API removals", out)
        self.assertIn("signature changes", out)

    def test_every_changed_symbol_is_named(self):
        _, out = run(PR_1120_DIFF)
        self.assertIn("ReservationByOrder", out)
        self.assertIn("ReservationByBid", out)

    def test_reports_two_signature_changes(self):
        _, out = run(PR_1120_DIFF)
        self.assertIn("(2)", out)


class TestRealRemovalsStillFail(unittest.TestCase):
    def test_unmatched_removal_fails(self):
        diff = (
            "--- a/x/settlement/keeper/grpc_query.go\n"
            "+++ b/x/settlement/keeper/grpc_query.go\n"
            "-func (q GRPCQuerier) LegacyCaseLookup(ctx context.Context) error {\n"
        )
        code, out = run(diff)
        self.assertEqual(code, 1)
        self.assertIn("Potential API removals", out)
        self.assertIn("LegacyCaseLookup", out)

    def test_mixed_diff_fails_and_lists_only_the_real_removal(self):
        diff = (
            "--- a/x/resources/keeper/grpc_query.go\n"
            "+++ b/x/resources/keeper/grpc_query.go\n"
            "-func (q *Querier) Kept(ctx context.Context) (*Old, error) {\n"
            "+func (q *Querier) Kept(ctx context.Context) (*New, error) {\n"
            "-func (q *Querier) Deleted(ctx context.Context) error {\n"
        )
        code, out = run(diff)
        self.assertEqual(code, 1)
        # Only the genuinely deleted symbol is a removal...
        self.assertIn("Querier.Deleted", out)
        self.assertIn("Potential API removals", out)
        # ...and the redeclared one is reported as a signature change, never
        # inside the removal list.
        removals_block = out.split("Potential API removals", 1)[1]
        self.assertNotIn("Querier.Kept", removals_block)


class TestReceiverHandling(unittest.TestCase):
    def test_pointer_vs_value_receiver_is_the_same_symbol(self):
        diff = (
            "--- a/x/keeper/grpc_query.go\n"
            "+++ b/x/keeper/grpc_query.go\n"
            "-func (q Querier) Thing() error {\n"
            "+func (q *Querier) Thing() error {\n"
        )
        code, _ = run(diff)
        self.assertEqual(code, 0, "receiver pointer-ness must not read as a removal")

    def test_method_moved_to_package_function_is_not_a_removal(self):
        # financialCases went from a GRPCQuerier method to a free function.
        diff = (
            "--- a/x/settlement/keeper/grpc_query.go\n"
            "+++ b/x/settlement/keeper/grpc_query.go\n"
            "-func (q GRPCQuerier) financialCases(kind string) error {\n"
            "+func financialCases(q GRPCQuerier, kind string) error {\n"
        )
        code, _ = run(diff)
        self.assertEqual(code, 0)

    def test_receiver_not_first_param_is_still_recognised_as_a_move(self):
        # The real PR #1120 case: ctx precedes the receiver, so a first-param
        # only check misses it and the gate false-reds again.
        diff = (
            "--- a/x/settlement/keeper/grpc_query.go\n"
            "+++ b/x/settlement/keeper/grpc_query.go\n"
            "-func (q GRPCQuerier) financialCases(ctx context.Context, kind string) error {\n"
            "+func financialCases(ctx context.Context, q GRPCQuerier, kind, key string) error {\n"
        )
        code, out = run(diff)
        self.assertEqual(code, 0, f"receiver in non-first position:\n{out}")

    def test_nested_commas_do_not_corrupt_param_matching(self):
        # map[string]any and a func-typed param both contain commas.
        diff = (
            "--- a/x/settlement/keeper/grpc_query.go\n"
            "+++ b/x/settlement/keeper/grpc_query.go\n"
            "-func (q GRPCQuerier) Cases(ctx context.Context) error {\n"
            "+func Cases(ctx context.Context, q GRPCQuerier, opts map[string]any, "
            "cb func(a, b int) error) error {\n"
        )
        code, out = run(diff)
        self.assertEqual(code, 0, f"nested commas broke param parsing:\n{out}")

    def test_different_receiver_type_same_name_is_still_a_removal(self):
        diff = (
            "--- a/x/keeper/grpc_query.go\n"
            "+++ b/x/keeper/grpc_query.go\n"
            "-func (q *OtherQuerier) Thing() error {\n"
            "+func (q *Querier) Thing() error {\n"
        )
        code, out = run(diff)
        self.assertEqual(code, 1, "a method moved between types is a real change")
        self.assertIn("Thing", out)


class TestCleanAndEmptyInputs(unittest.TestCase):
    def test_empty_diff_passes(self):
        code, out = run("")
        self.assertEqual(code, 0)
        self.assertIn("No removed public functions", out)

    def test_non_func_changes_pass(self):
        diff = "--- a/api/openapi/x.swagger.json\n+++ b/api/openapi/x.swagger.json\n-+ a new field\n"
        code, _ = run(diff)
        self.assertEqual(code, 0)

    def test_diff_metadata_is_not_a_removal(self):
        code, out = run("--- a/x/keeper/grpc_query.go\n+++ b/x/keeper/grpc_query.go\n")
        self.assertEqual(code, 0)
        self.assertIn("No removed public functions", out)

    def test_generic_function_declaration_is_parsed(self):
        diff = (
            "--- a/x/types/query.go\n"
            "+++ b/x/types/query.go\n"
            "-func Query[T any](k string) (T, error) {\n"
            "+func Query[K cmp.Ordered, T any](k K) (T, error) {\n"
        )
        code, _ = run(diff)
        self.assertEqual(code, 0)


if __name__ == "__main__":
    unittest.main(verbosity=2)
