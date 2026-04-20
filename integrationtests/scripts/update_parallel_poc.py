#!/usr/bin/env python3
import re
from collections import Counter
from pathlib import Path


def find_repo_root(start: Path) -> Path:
    for parent in [start] + list(start.parents):
        if (parent / "bitrise.yml").exists():
            return parent
    raise SystemExit("Could not find repo root (bitrise.yml).")


def read_sequential_tests(bitrise_yml: Path) -> list[str]:
    data = bitrise_yml.read_text()
    tests = []
    tests += re.findall(r'sequential_re\+="(Test[^"]+?)\|"', data)
    last = re.findall(r'sequential_re\+="(Test[^"]+?)"', data)
    if last:
        tests.append(last[-1])
    return tests


def map_tests_to_files(root: Path, tests: list[str]) -> tuple[Counter, Counter, list[str]]:
    base = root / "integrationtests"
    file_map = {}
    for path in base.rglob("*_test.go"):
        text = path.read_text()
        for name in tests:
            if f"func {name}(" in text:
                file_map[name] = path
    pkg_counts = Counter()
    file_counts = Counter()
    missing = []
    for name in tests:
        path = file_map.get(name)
        if not path:
            missing.append(name)
            continue
        rel = path.relative_to(base)
        pkg = rel.parts[0]
        pkg_counts[pkg] += 1
        file_counts[str(rel)] += 1
    return pkg_counts, file_counts, missing


def render_report(pkg_counts: Counter, file_counts: Counter, missing: list[str]) -> str:
    lines = []
    lines.append("## Where sequential tests concentrate")
    lines.append("")
    lines.append("- Regenerate: `python3 integrationtests/scripts/update_parallel_poc.py`")
    lines.append("")
    lines.append("- By subpackage (count of sequential tests):")
    for name, count in pkg_counts.most_common():
        lines.append(f"  - `{name}`: {count}")
    lines.append("- By file (count of sequential tests):")
    for name, count in file_counts.most_common():
        lines.append(f"  - `integrationtests/{name}`: {count}")
    if missing:
        lines.append("- Unmapped tests (not found in *_test.go):")
        for name in missing:
            lines.append(f"  - `{name}`")
    return "\n".join(lines) + "\n"


def update_doc(doc_path: Path, report: str) -> None:
    data = doc_path.read_text()
    start = "<!-- SEQ_REPORT_START -->"
    end = "<!-- SEQ_REPORT_END -->"
    if start not in data or end not in data:
        raise SystemExit("Missing SEQ_REPORT_START/SEQ_REPORT_END markers in PARALLEL_POC.md.")
    before, rest = data.split(start, 1)
    _, after = rest.split(end, 1)
    new_data = before + start + "\n" + report + end + after
    doc_path.write_text(new_data)


def main() -> None:
    root = find_repo_root(Path(__file__).resolve())
    tests = read_sequential_tests(root / "bitrise.yml")
    pkg_counts, file_counts, missing = map_tests_to_files(root, tests)
    report = render_report(pkg_counts, file_counts, missing)
    update_doc(root / "integrationtests" / "PARALLEL_POC.md", report)


if __name__ == "__main__":
    main()
