"""
Tests for the vendored CRAFT implementation and the CRAFTDetector integration.

CRAFT used to be supplied by the PyPI ``craft-text-detector`` package. It is now
vendored under ``ml/text_detection/third_party/craft_text_detector`` because no
release of that package can be installed on Python 3.11 (its ``install_requires``
caps ``opencv-python`` below every cp311 wheel) and because it imports
``torchvision.models.vgg.model_urls``, removed in torchvision 0.14. See
``third_party/craft_text_detector/README.md``.
"""

import inspect
from pathlib import Path

import numpy as np
import pytest
from unittest.mock import Mock, patch

from ml.text_detection.craft_detector import CRAFTDetector

VENDORED_PACKAGE = "ml.text_detection.third_party.craft_text_detector"


def _detector_with_result(prediction_result):
    """Build a CRAFTDetector whose model returns ``prediction_result``."""
    with patch.object(CRAFTDetector, "_load_model", return_value=Mock()):
        detector = CRAFTDetector()

    model = Mock()
    model.detect_text = Mock(return_value=prediction_result)
    detector._model = model
    return detector


class TestVendoredPackageImportSurface:
    """The vendored copy must expose the upstream API and import cleanly."""

    def test_exposes_upstream_public_api(self):
        """Craft and the module-level helpers are re-exported like upstream."""
        craft_text_detector = pytest.importorskip(VENDORED_PACKAGE)

        for name in (
            "Craft",
            "get_prediction",
            "read_image",
            "load_craftnet_model",
            "load_refinenet_model",
            "export_detected_regions",
            "export_extra_results",
            "empty_cuda_cache",
        ):
            assert hasattr(craft_text_detector, name), f"missing {name}"

    def test_model_modules_import_against_pinned_torchvision(self):
        """
        Upstream imported ``model_urls`` from torchvision; the vendored copy must
        not, otherwise it cannot import against torchvision >= 0.14.
        """
        pytest.importorskip("torch")
        pytest.importorskip("torchvision")

        from ml.text_detection.third_party.craft_text_detector.models.craftnet import (  # noqa: E501
            CraftNet,
        )
        from ml.text_detection.third_party.craft_text_detector.models.refinenet import (  # noqa: E501
            RefineNet,
        )

        assert inspect.isclass(CraftNet)
        assert inspect.isclass(RefineNet)

    def test_vgg16_bn_backbone_builds_without_fetching_imagenet_weights(self):
        """The CRAFT backbone must build from local weights only."""
        torch = pytest.importorskip("torch")
        pytest.importorskip("torchvision")

        from ml.text_detection.third_party.craft_text_detector.models.basenet.vgg16_bn import (  # noqa: E501
            vgg16_bn,
        )

        backbone = vgg16_bn(pretrained=False, freeze=True)

        # Slices mirror torchvision's vgg16_bn layout: 12 / 7 / 10 / 10 modules.
        assert len(backbone.slice1) == 12
        assert len(backbone.slice2) == 7
        assert len(backbone.slice3) == 10
        assert len(backbone.slice4) == 10
        assert isinstance(backbone.slice1[0], torch.nn.Conv2d)

    def test_craft_net_builds_with_expected_parameter_count(self):
        """CraftNet() must construct the full CRAFT architecture."""
        pytest.importorskip("torch")
        pytest.importorskip("torchvision")

        from ml.text_detection.third_party.craft_text_detector.models.craftnet import (  # noqa: E501
            CraftNet,
        )

        net = CraftNet()
        assert sum(p.numel() for p in net.parameters()) == 20770466


class TestVendoredPredictionOutput:
    """``get_prediction`` must expose both heatmaps and raw score maps."""

    @pytest.fixture(scope="class")
    def prediction(self):
        """Run one real forward pass with a randomly initialised CraftNet."""
        pytest.importorskip("torch")
        pytest.importorskip("torchvision")

        import ml.text_detection.third_party.craft_text_detector as craft_text_detector
        from ml.text_detection.third_party.craft_text_detector.models.craftnet import (  # noqa: E501
            CraftNet,
        )

        image = (
            np.random.RandomState(42).rand(300, 400, 3).astype(np.float32) * 255
        ).astype(np.uint8)

        net = CraftNet()
        net.eval()

        return craft_text_detector.get_prediction(
            image=image, craft_net=net, refine_net=None, long_size=128
        )

    def test_returns_raw_float32_scalar_score_maps(self, prediction):
        """The addition used by CRAFTDetector: 2-D float32 maps in [0, 1]."""
        score_maps = prediction["score_maps"]

        for key in ("text_score_map", "link_score_map"):
            score_map = score_maps[key]
            assert score_map.ndim == 2
            assert score_map.dtype == np.float32
            assert np.isfinite(score_map).all()

    def test_score_maps_use_native_heatmap_resolution(self, prediction):
        """Raw maps keep the network's heatmap resolution (half of long_size)."""
        text_score_map = prediction["score_maps"]["text_score_map"]

        # long_size=128 with a 400x300 input keeps the canvas at 128 on the long
        # edge, and CRAFT's heatmap is half of that: 64x48.
        assert text_score_map.shape == (48, 64)

    def test_heatmaps_remain_colourised_visualisations(self, prediction):
        """``heatmaps`` keeps its upstream meaning: 3-channel uint8 BGR."""
        for key in ("text_score_heatmap", "link_score_heatmap"):
            heatmap = prediction["heatmaps"][key]
            assert heatmap.ndim == 3
            assert heatmap.shape[2] == 3
            assert heatmap.dtype == np.uint8


class TestCRAFTDetectorScoreMapExtraction:
    """CRAFTDetector.detect must consume scalar maps in input-image space."""

    def test_reads_raw_score_maps_and_resamples_to_image_size(self, small_text_image):
        """Raw maps at heatmap resolution are resampled to (H, W) of the image."""
        height, width = small_text_image.shape[:2]

        result = {
            "score_maps": {
                "text_score_map": np.full(
                    (height // 2, width // 2), 0.9, dtype=np.float32
                ),
                "link_score_map": np.full(
                    (height // 2, width // 2), 0.5, dtype=np.float32
                ),
            },
            # Colourised visualisations are present in real CRAFT output too.
            "heatmaps": {
                "text_score_heatmap": np.zeros(
                    (height // 2, width // 2, 3), dtype=np.uint8
                ),
                "link_score_heatmap": np.zeros(
                    (height // 2, width // 2, 3), dtype=np.uint8
                ),
            },
        }

        region_scores, affinity_scores = _detector_with_result(result).detect(
            small_text_image
        )

        assert region_scores.shape == (height, width)
        assert affinity_scores.shape == (height, width)
        assert region_scores.dtype == np.float32
        assert affinity_scores.dtype == np.float32
        assert region_scores.min() >= 0.0 and region_scores.max() <= 1.0
        assert affinity_scores.min() >= 0.0 and affinity_scores.max() <= 1.0
        assert region_scores.mean() == pytest.approx(0.9, abs=0.02)
        assert affinity_scores.mean() == pytest.approx(0.5, abs=0.02)

    def test_raw_score_maps_take_priority_over_heatmaps(self, small_text_image):
        """The raw map wins when a result carries both representations."""
        height, width = small_text_image.shape[:2]

        result = {
            "score_maps": {
                "text_score_map": np.full(
                    (height // 2, width // 2), 0.25, dtype=np.float32
                ),
                "link_score_map": np.zeros((height // 2, width // 2), dtype=np.float32),
            },
            "heatmaps": {
                "text_score_heatmap": np.full((height, width), 0.95, dtype=np.float32),
                "link_score_heatmap": np.zeros((height, width), dtype=np.float32),
            },
        }

        region_scores, _ = _detector_with_result(result).detect(small_text_image)

        assert region_scores.mean() == pytest.approx(0.25, abs=0.02)

    def test_colourised_heatmaps_are_never_treated_as_score_maps(self, small_text_image):
        """
        Regression guard: a 3-channel colourised heatmap must not be thresholded
        as if it were a score map (which would mark the whole image as text).
        """
        height, width = small_text_image.shape[:2]

        result = {
            "heatmaps": {
                "text_score_heatmap": np.full((height, width, 3), 255, dtype=np.uint8),
                "link_score_heatmap": np.full((height, width, 3), 255, dtype=np.uint8),
            },
            "boxes": np.zeros((0, 4, 2), dtype=np.float32),
        }

        region_scores, affinity_scores = _detector_with_result(result).detect(
            small_text_image
        )

        assert region_scores.shape == (height, width)
        assert affinity_scores.shape == (height, width)
        assert float(region_scores.max()) == 0.0

    def test_legacy_attribute_style_result_is_still_supported(self, small_text_image):
        """Attribute-style results with scalar heatmaps keep working."""
        height, width = small_text_image.shape[:2]

        result = Mock()
        result.heatmaps = {
            "text_score_heatmap": np.full((height, width), 0.8, dtype=np.float32),
            "link_score_heatmap": np.full((height, width), 0.2, dtype=np.float32),
        }

        region_scores, affinity_scores = _detector_with_result(result).detect(
            small_text_image
        )

        assert region_scores.shape == (height, width)
        assert region_scores.mean() == pytest.approx(0.8, abs=1e-5)
        assert affinity_scores.mean() == pytest.approx(0.2, abs=1e-5)

    def test_box_fallback_produces_masks_in_image_coordinates(self, small_text_image):
        """With no score maps at all, boxes are rasterised into the image mask."""
        height, width = small_text_image.shape[:2]

        box = np.array([[10, 10], [60, 10], [60, 50], [10, 50]], dtype=np.float32)
        result = {"boxes": np.array([box]), "heatmaps": {}}

        region_scores, affinity_scores = _detector_with_result(result).detect(
            small_text_image
        )

        assert region_scores.shape == (height, width)
        assert float(region_scores[30, 30]) == 1.0  # inside the box
        assert float(region_scores[200, 200]) == 0.0  # outside the box
        assert float(affinity_scores.max()) == 0.0

    def test_detect_is_deterministic(self, small_text_image):
        """Repeated detection over identical input yields identical maps."""
        height, width = small_text_image.shape[:2]

        result = {
            "score_maps": {
                "text_score_map": np.full(
                    (height // 2, width // 2), 0.7, dtype=np.float32
                ),
                "link_score_map": np.full(
                    (height // 2, width // 2), 0.3, dtype=np.float32
                ),
            }
        }

        detector = _detector_with_result(result)
        first = detector.detect(small_text_image.copy())
        second = detector.detect(small_text_image.copy())

        np.testing.assert_array_equal(first[0], second[0])
        np.testing.assert_array_equal(first[1], second[1])


class TestVendoredEndToEndWithoutNetwork:
    """Exercise the vendored loader and Craft.detect_text with local weights."""

    @pytest.fixture(scope="class")
    def craft_weight_paths(self, tmp_path_factory):
        """Write CRAFT/refiner state dicts to disk so no download is needed."""
        torch = pytest.importorskip("torch")
        pytest.importorskip("torchvision")

        from ml.text_detection.third_party.craft_text_detector.models.craftnet import (  # noqa: E501
            CraftNet,
        )
        from ml.text_detection.third_party.craft_text_detector.models.refinenet import (  # noqa: E501
            RefineNet,
        )

        weights_dir = tmp_path_factory.mktemp("craft_weights")
        craft_path = weights_dir / "craft_mlt_25k.pth"
        refiner_path = weights_dir / "craft_refiner_CTW1500.pth"

        torch.save(CraftNet().state_dict(), craft_path)
        torch.save(RefineNet().state_dict(), refiner_path)

        return craft_path, refiner_path

    def test_load_craftnet_model_reads_a_local_checkpoint(self, craft_weight_paths):
        """The vendored loader must load from an explicit weight_path."""
        pytest.importorskip("torch")
        import ml.text_detection.third_party.craft_text_detector as craft_text_detector
        from ml.text_detection.third_party.craft_text_detector.models.craftnet import (  # noqa: E501
            CraftNet,
        )

        craft_path, _ = craft_weight_paths

        model = craft_text_detector.load_craftnet_model(
            cuda=False, weight_path=craft_path
        )

        assert isinstance(model, CraftNet)
        assert model.training is False  # eval() is applied on load

    def test_load_refinenet_model_reads_a_local_checkpoint(self, craft_weight_paths):
        """The refiner path (enabled by default) must load too."""
        pytest.importorskip("torch")
        import ml.text_detection.third_party.craft_text_detector as craft_text_detector
        from ml.text_detection.third_party.craft_text_detector.models.refinenet import (  # noqa: E501
            RefineNet,
        )

        _, refiner_path = craft_weight_paths

        model = craft_text_detector.load_refinenet_model(
            cuda=False, weight_path=refiner_path
        )

        assert isinstance(model, RefineNet)
        assert model.training is False

    def test_craft_detect_text_runs_end_to_end(self, craft_weight_paths):
        """Craft.detect_text must run without network access using local weights."""
        pytest.importorskip("torch")
        pytest.importorskip("torchvision")
        import ml.text_detection.third_party.craft_text_detector as craft_text_detector

        craft_path, refiner_path = craft_weight_paths

        craft = craft_text_detector.Craft(
            output_dir=None,
            crop_type="box",
            cuda=False,
            long_size=128,
            weight_path_craft_net=craft_path,
            weight_path_refine_net=refiner_path,
        )

        image = (
            np.random.RandomState(7).rand(200, 300, 3).astype(np.float32) * 255
        ).astype(np.uint8)

        result = craft.detect_text(image)

        assert isinstance(result["boxes"], list)
        assert isinstance(result["polys"], list)
        assert result["score_maps"]["text_score_map"].dtype == np.float32

        craft.unload_craftnet_model()
        craft.unload_refinenet_model()
        assert craft.craft_net is None
        assert craft.refine_net is None


class TestRequirementsNoLongerPinArchivedPackage:
    """The unsatisfiable PyPI dependency must not come back."""

    def test_requirements_has_no_active_craft_text_detector_entry(self):
        requirements = (
            Path(__file__).resolve().parents[1] / "requirements.txt"
        ).read_text(encoding="utf-8")

        active_lines = [
            line.strip()
            for line in requirements.splitlines()
            if line.strip() and not line.strip().startswith("#")
        ]

        offenders = [
            line for line in active_lines if line.startswith("craft-text-detector")
        ]
        assert not offenders, (
            "craft-text-detector cannot resolve on Python 3.11 (it caps "
            "opencv-python below every cp311 wheel); CRAFT is vendored instead: "
            f"{offenders}"
        )

    def test_gdown_is_declared_for_weight_download(self):
        """The vendored copy still fetches weights via gdown."""
        requirements = (
            Path(__file__).resolve().parents[1] / "requirements.txt"
        ).read_text(encoding="utf-8")

        assert "gdown" in requirements
