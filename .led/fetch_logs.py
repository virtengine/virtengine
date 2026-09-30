"""Recover pre-lock failed-job logs for the ledger. Writes one .log per failed job."""
import json
import os
import re
import subprocess
import sys

REPO = "virtengine/virtengine"
OUT = os.path.join(os.path.dirname(os.path.abspath(__file__)), "logs")
os.makedirs(OUT, exist_ok=True)

RUNS = {
    36073455679: "CI",
    36073455483: "Quality Gate",
    36073455096: "Security",
    36073455817: "Supply Chain Security",
    36073455131: "compatibility",
}

NOISE = re.compile(
    r"^(\s*$|##\[group\]|##\[endgroup\]|##\[command\]|Current runner version|"
    r"Runner Image|Operating System|.*virtual.*environment|Runner name|"
    r"Received signal|##\[error\]Process completed|Set up job|"
    r"Post job cleanup|Download action repository|Prepare workflow directory|"
    r"Complete job|Starting: |Restore cache|Save cache)",
    re.I,
)


def gh(*args, timeout=900):
    return subprocess.run(
        ["gh", *args, "-R", REPO],
        capture_output=True,
        text=True,
        timeout=timeout,
    )


summary = {}
for run_id, wf in RUNS.items():
    r = gh("run", "view", str(run_id), "--json", "jobs")
    if r.returncode != 0:
        print(f"!! run {run_id} view failed: {r.stderr[:300]}", flush=True)
        continue
    jobs = json.loads(r.stdout)["jobs"]
    for job in jobs:
        if job.get("conclusion") != "failure":
            continue
        jid = job["databaseId"]
        slug = re.sub(r"[^A-Za-z0-9_.-]+", "-", job["name"])[:60]
        path = os.path.join(OUT, f"{run_id}-{jid}-{slug}.log")
        # Job logs via the API endpoint gh wraps: gh run view --job <id> --log
        lg = gh("run", "view", "--job", str(jid), "--log")
        if lg.returncode != 0 or not lg.stdout.strip():
            summary[job["name"]] = {"run": run_id, "job": jid, "log": "UNAVAILABLE",
                                    "err": (lg.stderr or lg.stdout)[:200]}
            continue
        with open(path, "w", encoding="utf-8", errors="replace") as fh:
            fh.write(lg.stdout)
        lines = [l for l in lg.stdout.splitlines() if not NOISE.match(l)]
        tail = lines[-45:]
        summary[job["name"]] = {"run": run_id, "job": jid, "log": path,
                                "tail": "\n".join(tail)[-3000:]}
        print(f"== {wf} run {run_id} job {jid} :: {job['name']} -> {os.path.basename(path)}",
              flush=True)

with open(os.path.join(OUT, "index.json"), "w", encoding="utf-8") as fh:
    json.dump(summary, fh, indent=1)
print("\n\n########## INDEX ##########")
print(json.dumps(summary, indent=1))