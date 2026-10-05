#!/usr/bin/env python3
"""Regression tests for GitHub's matrix-output aggregation.

Measured on Security run 37295691699: each govulncheck shard emits a ONE-ELEMENT
JSON object as a job output, and GitHub CONCATENATES matrix outputs verbatim, so
the job output arrived as

    {"core":"clean"}{"platform":"clean"}{"tooling":"clean"}{"tests-and-sim":"clean"}

The consumer passed that straight to `json.loads`, which rejects trailing data,
so the gate failed closed with "not valid JSON" and the run reported
`{tests-and-sim:clean}` - only the LAST shard survived. The workflow now wraps
the aggregation in `fromJSON`, and these tests pin the arithmetic that depends
on, using the exact strings the shard step writes.

The last test is the load-bearing one: it asserts plain `json.loads` STILL
rejects the concatenation. Without it this suite would pass just as happily
against the broken design it exists to catch.

Run: python -m unittest discover -s .github/tests -p "test_matrix_output_aggregation*.py"
"""
from __future__ import annotations

import json
import unittest

SHARDS = ["core", "platform", "tooling", "tests-and-sim"]


def shard_output(shard: str, verdict: str = "clean") -> str:
    """Exactly what the shard step appends to $GITHUB_OUTPUT."""
    return 'verdicts={"%s":"%s"}' % (shard, verdict)


def github_aggregates(shards: list[str], verdict: str = "clean") -> str:
    """What GitHub delivers: the `name=` prefix stripped, values concatenated."""
    return "".join(shard_output(s, verdict).split("=", 1)[1] for s in shards)


def from_json(raw: str) -> dict[str, str]:
    """GitHub's `fromJSON` over a run of concatenated one-element objects.

    Decodes each object in sequence and merges them, which is what makes the
    concatenation reassemble instead of erroring.
    """
    merged: dict[str, str] = {}
    decoder = json.JSONDecoder()
    idx, n = 0, len(raw)
    while idx < n:
        while idx < n and raw[idx].isspace():
            idx += 1
        if idx >= n:
            break
        obj, idx = decoder.raw_decode(raw, idx)
        if not isinstance(obj, dict):
            raise ValueError(f"not a JSON object at offset {idx}")
        merged.update(obj)
    return merged


class TestAggregation(unittest.TestCase):
    def test_all_shards_survive_the_concatenation(self):
        merged = from_json(github_aggregates(SHARDS))
        self.assertEqual(sorted(merged), sorted(SHARDS))
        self.assertEqual(set(merged.values()), {"clean"})

    def test_cancelled_shard_is_absent_after_reassembly(self):
        """The whole point: an absent key is the no-verdict signal."""
        merged = from_json(github_aggregates(SHARDS[:3]))
        self.assertNotIn("tests-and-sim", merged)
        self.assertEqual(sorted(merged), sorted(SHARDS[:3]))

    def test_single_shard_is_still_valid(self):
        self.assertEqual(from_json(github_aggregates(["core"])), {"core": "clean"})

    def test_empty_default_object_is_accepted(self):
        """The workflow's `|| '{}'` default, for a wholly cancelled matrix."""
        self.assertEqual(from_json("{}"), {})

    def test_verdicts_are_preserved_not_flattened_to_clean(self):
        merged = from_json(
            github_aggregates(["core"], "clean")
            + shard_output("platform", "vulns").split("=", 1)[1]
            + shard_output("tooling", "incomplete").split("=", 1)[1]
        )
        self.assertEqual(merged["core"], "clean")
        self.assertEqual(merged["platform"], "vulns")
        self.assertEqual(merged["tooling"], "incomplete")

    def test_plain_json_loads_still_rejects_the_concatenation(self):
        """Guards the fix: if this ever PASSES, the suite models nothing.

        This is the exact failure observed in CI - `json.loads` on the
        aggregated output raised, the gate reported only the last shard, and the
        step was red for a transport reason while looking like a shard finding.
        """
        raw = github_aggregates(SHARDS)
        with self.assertRaises(json.JSONDecodeError):
            json.loads(raw)


if __name__ == "__main__":
    unittest.main(verbosity=2)