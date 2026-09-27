# Vendored CRAFT text detector

This directory contains a vendored copy of the CRAFT (Character-Region Awareness
for Text detection) inference implementation.

## Provenance

| Item | Value |
| --- | --- |
| Upstream project | https://github.com/fcakyon/craft-text-detector |
| Upstream release | `craft-text-detector==0.4.3` (sdist `craft-text-detector-0.4.3.tar.gz`) |
| Upstream release date | 2022-05-09 |
| Upstream status | Archived by the owner on 2022-12-02; no further releases |
| License | MIT (see `LICENSE`) — Copyright (c) 2019 NAVER Corp., 2020 Fatih C Akyon |
| Model lineage | CRAFT, https://github.com/clovaai/CRAFT-pytorch (CVPR 2019) |

## Why this is vendored instead of installed

`ml/text_detection/requirements.txt` cannot depend on `craft-text-detector` on the
CI interpreter (Python 3.11), for two independent reasons:

1. **Unsatisfiable dependency metadata.** `craft-text-detector==0.4.3` declares
   `opencv-python>=3.4.8.29,<4.5.4.62` in its `install_requires`. Every published
   version of that project carries the same cap and the project is archived, so no
   version allows a 3.11-compatible OpenCV. `opencv-python<4.5.4.62` only ships
   wheels for CPython 3.6–3.10, so there is no installable combination at all on
   3.11. `pip-audit -r` therefore could not resolve the file and produced no report.

2. **Incompatible with the pinned torchvision.** `models/basenet/vgg16_bn.py`
   imports `model_urls` from `torchvision.models.vgg`. That symbol was removed in
   torchvision 0.14, so the module fails to import against the `torchvision==0.27.1+cpu`
   this project pins.

Removing the PyPI dependency and vendoring the implementation is the fix that keeps
CRAFT available while making the requirements file honestly resolvable and auditable.

## Changes relative to upstream 0.4.3

The vendored sources are byte-for-byte upstream except for the following, all of
which are required to import and run under the pinned toolchain:

1. **Absolute self-imports rewritten to relative imports.** Upstream imported its own
   submodules as `import craft_text_detector.craft_utils as craft_utils`, which only
   resolves when the package sits on `sys.path` under its own top-level name. The
   vendored copy uses `from . import ...` so it can live under
   `ml.text_detection.third_party`.
2. **`models/basenet/vgg16_bn.py` ported to the modern torchvision API.** The
   `model_urls` import and the `model_urls["vgg16_bn"].replace("https://", "http://")`
   line were removed (the symbol no longer exists), and the constructor now selects
   `VGG16_BN_Weights` explicitly instead of relying on the deprecated `pretrained=`
   keyword. `pretrained=False` maps to `weights=None`, preserving upstream behaviour:
   the CRAFT checkpoint supplies every parameter, so no ImageNet weights are fetched.
3. **`predict.get_prediction` also returns the raw score maps.** Upstream returned only
   the *colourised* heatmaps (`cv2.applyColorMap` output, `uint8` BGR) under
   `"heatmaps"`. `CRAFTDetector.get_bounding_boxes` needs the raw `float32` region and
   affinity maps in `[0, 1]`, so `get_prediction` additionally returns
   `result["score_maps"]["text_score_map"]` / `["link_score_map"]`. The change is
   additive; the `"heatmaps"` key keeps its upstream meaning and shape.

No model architecture, threshold, or post-processing behaviour was modified.

## Weights

The CRAFT weights are not vendored. They are fetched on first use by
`craft_utils.load_craftnet_model` into `~/.craft_text_detector/weights/` via `gdown`,
exactly as upstream did. `gdown` is pinned in `ml/text_detection/requirements.txt`.
