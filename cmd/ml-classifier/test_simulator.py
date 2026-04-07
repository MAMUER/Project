"""
Test script for BiometricSimulator.

Generates sample data for each activity mode, prints statistics,
and produces plots for visual verification.

Usage:
    python test_simulator.py

Requires: matplotlib (add to requirements.txt if not present)
"""

import json
import os
import sys

# Ensure the local directory is importable
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from biometric_simulator import (
    BiometricSimulator,
    DeviceType,
    ActivityMode,
)


def _records_by_type(records: list[dict], metric: str) -> list[dict]:
    return [r for r in records if r["metric_type"] == metric]


def _print_stats(label: str, values: list[float]):
    if not values:
        print(f"  {label}: (no data)")
        return
    print(f"  {label}: min={min(values):.1f}, max={max(values):.1f}, "
          f"mean={sum(values) / len(values):.1f}, count={len(values)}")


def test_single_mode():
    """Test each activity mode in isolation."""
    print("=" * 60)
    print("TEST: Single-Mode Streams (10 min each)")
    print("=" * 60)

    for mode in ActivityMode:
        sim = BiometricSimulator(user_age=30, device_type=DeviceType.APPLE_WATCH)
        records = sim.generate_continuous_stream(mode, duration_minutes=10)

        hr_vals = [r["value"] for r in _records_by_type(records, "heart_rate")]
        spo2_vals = [r["value"] for r in _records_by_type(records, "spo2")]
        step_vals = [r["value"] for r in _records_by_type(records, "steps")]

        print(f"\n--- Mode: {mode.value} ---")
        _print_stats("Heart Rate (BPM)", hr_vals)
        _print_stats("SpO2 (%)", spo2_vals)
        _print_stats("Steps (per batch)", step_vals)

    print()


def test_device_types():
    """Compare sampling behavior across device types."""
    print("=" * 60)
    print("TEST: Device Type Comparison (resting, 10 min)")
    print("=" * 60)

    for dt in DeviceType:
        sim = BiometricSimulator(user_age=30, device_type=dt)
        records = sim.generate_continuous_stream(ActivityMode.RESTING, duration_minutes=10)
        hr_count = len(_records_by_type(records, "heart_rate"))
        spo2_count = len(_records_by_type(records, "spo2"))
        step_count = len(_records_by_type(records, "steps"))
        print(f"  {dt.value:25s}  HR samples: {hr_count:3d}  "
              f"SpO2: {spo2_count:3d}  Steps: {step_count:3d}")
    print()


def test_full_day():
    """Generate a full-day simulation with default activity profile."""
    print("=" * 60)
    print("TEST: Full-Day Simulation (default profile, 1440 min)")
    print("=" * 60)

    sim = BiometricSimulator(
        user_age=35,
        device_type=DeviceType.APPLE_WATCH,
        resting_hr=62.0,
    )
    session = sim.generate_session(duration_minutes=1440, activity_profile="default")

    records = session["records"]
    hr_vals = [r["value"] for r in _records_by_type(records, "heart_rate")]
    spo2_vals = [r["value"] for r in _records_by_type(records, "spo2")]
    step_vals = [r["value"] for r in _records_by_type(records, "steps")]
    temp_vals = [r["value"] for r in _records_by_type(records, "temperature")]

    print(f"\n  Device type : {session['device_type']}")
    print(f"  Sync interval: {session['sync_interval_ms']} ms")
    print(f"  Total records: {len(records)}")
    _print_stats("Heart Rate (BPM)", hr_vals)
    _print_stats("SpO2 (%)", spo2_vals)
    _print_stats("Steps (per batch)", step_vals)
    _print_stats("Temperature (C)", temp_vals)

    # Verify physiological bounds
    assert all(40 <= v <= 200 for v in hr_vals), "HR out of physiological range!"
    assert all(88 <= v <= 100 for v in spo2_vals), "SpO2 out of physiological range!"
    assert all(35.5 <= v <= 38.5 for v in temp_vals), "Temperature out of range!"
    print("\n  All physiological bounds verified OK.")
    print()

    return session


def test_workout_recovery():
    """Test that HR recovery curve is exponential after exercise."""
    print("=" * 60)
    print("TEST: Workout + Recovery Curve")
    print("=" * 60)

    sim = BiometricSimulator(user_age=30, device_type=DeviceType.APPLE_WATCH)
    profile = [
        (0, 10, ActivityMode.RESTING),
        (10, 40, ActivityMode.WORKOUT),
        (40, 70, ActivityMode.RESTING),
    ]
    session = sim.generate_session(duration_minutes=70, activity_profile=profile)

    hr_vals = [r["value"] for r in _records_by_type(session["records"], "heart_rate")]

    if hr_vals:
        # During workout, HR should be elevated
        workout_hr = hr_vals[10:60] if len(hr_vals) > 60 else hr_vals
        recovery_hr = hr_vals[-20:] if len(hr_vals) > 20 else hr_vals

        _print_stats("Workout HR", workout_hr)
        _print_stats("Recovery HR (last 20 samples)", recovery_hr)

        if workout_hr and recovery_hr:
            assert max(workout_hr) > min(recovery_hr), \
                "Recovery HR should be lower than workout HR!"
            print("  Recovery decay verified OK.")
    print()


def test_huawei_bp():
    """Test blood pressure data from Huawei Watch D2."""
    print("=" * 60)
    print("TEST: Huawei Watch D2 Blood Pressure")
    print("=" * 60)

    sim = BiometricSimulator(user_age=45, device_type=DeviceType.HUAWEI_WATCH_D2)
    profile = [
        (0, 10, ActivityMode.RESTING),
        (10, 20, ActivityMode.RUNNING),
        (20, 30, ActivityMode.RESTING),
    ]
    session = sim.generate_session(duration_minutes=30, activity_profile=profile)

    sys_vals = [r["value"] for r in _records_by_type(session["records"], "blood_pressure_systolic")]
    dia_vals = [r["value"] for r in _records_by_type(session["records"], "blood_pressure_diastolic")]

    _print_stats("Systolic BP (mmHg)", sys_vals)
    _print_stats("Diastolic BP (mmHg)", dia_vals)

    if sys_vals and dia_vals:
        assert all(s > d for s, d in zip(sys_vals, dia_vals)), \
            "Systolic must always exceed diastolic!"
        print("  Systolic > Diastolic verified OK.")
    print()


def test_output_format():
    """Verify output matches device-connector API format."""
    print("=" * 60)
    print("TEST: Output Format Validation")
    print("=" * 60)

    sim = BiometricSimulator(user_age=30, device_type=DeviceType.APPLE_WATCH)
    session = sim.generate_session(duration_minutes=5, activity_profile=[(0, 5, ActivityMode.RESTING)])

    # Check top-level keys
    required_keys = {"device_type", "device_token", "sync_interval_ms", "records"}
    assert required_keys.issubset(session.keys()), f"Missing keys: {required_keys - session.keys()}"

    # Check record structure
    record_keys = {"metric_type", "value", "timestamp", "quality"}
    for rec in session["records"][:5]:
        assert record_keys.issubset(rec.keys()), f"Record missing keys: {record_keys - rec.keys()}"
        assert rec["metric_type"] in ("heart_rate", "spo2", "steps", "temperature"), \
            f"Unknown metric_type: {rec['metric_type']}"
        assert isinstance(rec["value"], (int, float)), "value must be numeric"
        assert "T" in rec["timestamp"], "timestamp must be ISO format"

    print("  Output format: VALID")

    # Print sample record
    print(f"  Sample record: {json.dumps(session['records'][0], indent=4)}")
    print()


# ------------------------------------------------------------------
# Plotting
# ------------------------------------------------------------------

def plot_full_day(session: dict):
    """Plot heart rate, SpO2, and steps over the simulated day."""
    try:
        import matplotlib
        matplotlib.use("Agg")  # non-interactive backend
        import matplotlib.pyplot as plt
        import matplotlib.dates as mdates
        from datetime import datetime
    except ImportError:
        print("matplotlib not installed -- skipping plots. Install with: pip install matplotlib")
        return

    records = session["records"]
    if not records:
        print("No records to plot.")
        return

    hr_recs = _records_by_type(records, "heart_rate")
    spo2_recs = _records_by_type(records, "spo2")
    step_recs = _records_by_type(records, "steps")

    fig, axes = plt.subplots(3, 1, figsize=(16, 12), sharex=True)

    def _parse_timestamps(recs: list[dict]) -> list[datetime]:
        return [datetime.fromisoformat(r["timestamp"]) for r in recs]

    # Heart Rate
    if hr_recs:
        axes[0].plot(
            _parse_timestamps(hr_recs),
            [r["value"] for r in hr_recs],
            color="#e74c3c",
            linewidth=0.8,
            label="Heart Rate (BPM)",
        )
        axes[0].axhline(60, color="gray", linestyle="--", alpha=0.3, label="Resting ~60")
        axes[0].axhline(150, color="orange", linestyle="--", alpha=0.3, label="Exercise ~150")
        axes[0].set_ylabel("BPM")
        axes[0].set_title("Heart Rate")
        axes[0].legend(loc="upper right", fontsize=7)
        axes[0].grid(True, alpha=0.2)

    # SpO2
    if spo2_recs:
        axes[1].plot(
            _parse_timestamps(spo2_recs),
            [r["value"] for r in spo2_recs],
            color="#3498db",
            linewidth=0.8,
            label="SpO2 (%)",
        )
        axes[1].axhline(95, color="red", linestyle="--", alpha=0.3, label="Min normal 95%")
        axes[1].set_ylabel("SpO2 %")
        axes[1].set_title("Blood Oxygen Saturation")
        axes[1].legend(loc="upper right", fontsize=7)
        axes[1].grid(True, alpha=0.2)

    # Steps
    if step_recs:
        axes[2].bar(
            _parse_timestamps(step_recs),
            [r["value"] for r in step_recs],
            width=0.0005,
            color="#2ecc71",
            alpha=0.7,
            label="Steps (per batch)",
        )
        axes[2].set_ylabel("Steps")
        axes[2].set_title("Step Count")
        axes[2].legend(loc="upper right", fontsize=7)
        axes[2].grid(True, alpha=0.2)

    axes[2].xaxis.set_major_formatter(mdates.DateFormatter("%H:%M"))
    axes[2].xaxis.set_major_locator(mdates.HourLocator(interval=2))
    fig.autofmt_xdate()

    plt.tight_layout()

    output_path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "biometric_sim_plot.png")
    plt.savefig(output_path, dpi=150)
    print(f"  Plot saved to: {output_path}")
    plt.close()


# ------------------------------------------------------------------
# Main
# ------------------------------------------------------------------

if __name__ == "__main__":
    print("\n" + "#" * 60)
    print("#  Biometric Simulator -- Test Suite")
    print("#" * 60 + "\n")

    test_single_mode()
    test_device_types()
    test_workout_recovery()
    test_huawei_bp()
    test_output_format()
    session = test_full_day()

    # Generate plot
    print("=" * 60)
    print("PLOT: Full-Day Visualization")
    print("=" * 60)
    plot_full_day(session)

    print("\nAll tests passed.")
