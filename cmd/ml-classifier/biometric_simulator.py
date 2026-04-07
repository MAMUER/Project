"""
Biometric Simulator -- Physiologically Correct Vital Signs Generator

Produces realistic, time-coherent biometric data patterns that simulate
real human vital signs. Uses smoothed noise (Perlin-like), exponential
recovery curves, time-of-day circadian effects, and device-specific
sampling characteristics.

Usage:
    from biometric_simulator import BiometricSimulator
    sim = BiometricSimulator(user_age=30, device_type="apple_watch")
    session = sim.generate_session(duration_minutes=1440, activity_profile="default")
"""

import math
import random
import json
import uuid
from datetime import datetime, timedelta, timezone
from dataclasses import dataclass, field, asdict
from typing import Literal, Optional
from enum import Enum


# ---------------------------------------------------------------------------
# Deterministic smoothed-noise helpers
# ---------------------------------------------------------------------------

def _seeded_random(seed: int) -> random.Random:
    """Return a seeded Random instance for reproducibility."""
    return random.Random(seed)


def _lerp(a: float, b: float, t: float) -> float:
    return a + (b - a) * t


def _smoothstep(t: float) -> float:
    t = max(0.0, min(1.0, t))
    return t * t * (3 - 2 * t)


class SmoothedNoise:
    """1-D Perlin-like smoothed noise for natural variability."""

    def __init__(self, seed: int = 42, grid_spacing: float = 1.0):
        self.rng = _seeded_random(seed)
        self.grid_spacing = grid_spacing
        self._gradients: dict[int, float] = {}

    def _get_gradient(self, ix: int) -> float:
        if ix not in self._gradients:
            self._gradients[ix] = self.rng.uniform(-1.0, 1.0)
        return self._gradients[ix]

    def __call__(self, x: float) -> float:
        gx = x / self.grid_spacing
        ix0 = math.floor(gx)
        ix1 = ix0 + 1
        t = gx - ix0
        t = _smoothstep(t)
        g0 = self._get_gradient(ix0)
        g1 = self._get_gradient(ix1)
        return _lerp(g0, g1, t)


# ---------------------------------------------------------------------------
# Device profiles
# ---------------------------------------------------------------------------

class DeviceType(Enum):
    APPLE_WATCH = "apple_watch"
    SAMSUNG_GALAXY_WATCH = "samsung_galaxy_watch"
    HUAWEI_WATCH_D2 = "huawei_watch_d2"
    AMAZFIT_T_REX_3 = "amazfit_trex_3"


@dataclass
class DeviceProfile:
    name: str
    hr_accuracy_bpm: float          # ± BPM
    hr_sample_interval_active: int  # seconds during activity
    hr_sample_interval_rest: int    # seconds at rest
    spo2_sample_interval: int       # seconds
    step_batch_interval: int        # seconds
    temp_sample_interval: int       # seconds
    has_blood_pressure: bool = False


DEVICE_PROFILES: dict[DeviceType, DeviceProfile] = {
    DeviceType.APPLE_WATCH: DeviceProfile(
        name="apple_watch", hr_accuracy_bpm=2,
        hr_sample_interval_active=5, hr_sample_interval_rest=60,
        spo2_sample_interval=60, step_batch_interval=30,
        temp_sample_interval=300,
    ),
    DeviceType.SAMSUNG_GALAXY_WATCH: DeviceProfile(
        name="samsung_galaxy_watch", hr_accuracy_bpm=3,
        hr_sample_interval_active=10, hr_sample_interval_rest=60,
        spo2_sample_interval=60, step_batch_interval=30,
        temp_sample_interval=300,
    ),
    DeviceType.HUAWEI_WATCH_D2: DeviceProfile(
        name="huawei_watch_d2", hr_accuracy_bpm=2,
        hr_sample_interval_active=10, hr_sample_interval_rest=60,
        spo2_sample_interval=60, step_batch_interval=30,
        temp_sample_interval=300, has_blood_pressure=True,
    ),
    DeviceType.AMAZFIT_T_REX_3: DeviceProfile(
        name="amazfit_trex_3", hr_accuracy_bpm=5,
        hr_sample_interval_active=30, hr_sample_interval_rest=60,
        spo2_sample_interval=60, step_batch_interval=30,
        temp_sample_interval=300,
    ),
}


# ---------------------------------------------------------------------------
# Activity profiles (a full-day schedule in minutes)
# ---------------------------------------------------------------------------

class ActivityMode(Enum):
    RESTING = "resting"
    WALKING = "walking"
    RUNNING = "running"
    SLEEPING = "sleeping"
    WORKOUT = "workout"


def _default_day_profile() -> list[tuple[int, int, ActivityMode]]:
    """Return a list of (start_min, end_min, mode) covering 0-1440."""
    return [
        (0, 420, ActivityMode.SLEEPING),     # 00:00 - 07:00  sleep
        (420, 480, ActivityMode.RESTING),     # 07:00 - 08:00  wake / breakfast
        (480, 540, ActivityMode.WALKING),     # 08:00 - 09:00  commute walk
        (540, 720, ActivityMode.RESTING),     # 09:00 - 12:00  desk work
        (720, 780, ActivityMode.WALKING),     # 12:00 - 13:00  lunch walk
        (780, 960, ActivityMode.RESTING),     # 13:00 - 16:00  desk work
        (960, 1020, ActivityMode.WORKOUT),    # 16:00 - 17:00  workout session
        (1020, 1080, ActivityMode.WALKING),   # 17:00 - 18:00  cool-down walk
        (1080, 1200, ActivityMode.RESTING),   # 18:00 - 20:00  evening rest
        (1200, 1260, ActivityMode.WALKING),   # 20:00 - 21:00  evening walk
        (1260, 1380, ActivityMode.RESTING),   # 21:00 - 23:00  wind down
        (1380, 1440, ActivityMode.SLEEPING),  # 23:00 - 24:00  sleep
    ]


# ---------------------------------------------------------------------------
# BiometricRecord
# ---------------------------------------------------------------------------

@dataclass
class BiometricRecord:
    metric_type: str
    value: float
    timestamp: str
    quality: str = "good"


# ---------------------------------------------------------------------------
# Main Simulator
# ---------------------------------------------------------------------------

class BiometricSimulator:
    """
    Generate physiologically realistic biometric streams.

    Parameters
    ----------
    user_age : int
        Used for HRmax = 220 - age calculations.
    device_type : DeviceType | str
        Determines sampling rates and measurement accuracy.
    resting_hr : float
        Baseline resting heart rate (default 65).
    seed : int
        Random seed for reproducibility.
    """

    def __init__(
        self,
        user_age: int = 30,
        device_type: DeviceType | str = DeviceType.APPLE_WATCH,
        resting_hr: float = 65.0,
        seed: int = 42,
    ):
        self.user_age = user_age
        self.hr_max = 220 - user_age
        self.resting_hr = resting_hr
        self.seed = seed

        # Resolve device
        if isinstance(device_type, str):
            for dt in DeviceType:
                if dt.value == device_type:
                    device_type = dt
                    break
            else:
                raise ValueError(f"Unknown device_type: {device_type}")
        self.device_type = device_type
        self.profile = DEVICE_PROFILES[device_type]

        # Noise generators (each metric gets its own noise stream)
        self._hr_noise = SmoothedNoise(seed=seed, grid_spacing=30.0)
        self._hr_micro_noise = SmoothedNoise(seed=seed + 1, grid_spacing=5.0)
        self._spo2_noise = SmoothedNoise(seed=seed + 2, grid_spacing=60.0)
        self._temp_noise = SmoothedNoise(seed=seed + 3, grid_spacing=120.0)
        self._bp_sys_noise = SmoothedNoise(seed=seed + 4, grid_spacing=60.0)
        self._bp_dia_noise = SmoothedNoise(seed=seed + 5, grid_spacing=60.0)

        self._rng = _seeded_random(seed)

        # State tracking across generate_session calls
        self._recovery_start_hr: float | None = None
        self._recovery_start_time: float | None = None
        self._last_activity: ActivityMode | None = None

    # ------------------------------------------------------------------
    # Public API
    # ------------------------------------------------------------------

    def generate_session(
        self,
        duration_minutes: int = 1440,
        activity_profile: list[tuple[int, int, ActivityMode]] | Literal["default"] = "default",
        start_time: datetime | None = None,
    ) -> dict:
        """
        Generate a complete biometric session.

        Parameters
        ----------
        duration_minutes : int
            Total simulated minutes.
        activity_profile : "default" | list of (start_min, end_min, mode)
            Schedule of activities across the session.
        start_time : datetime
            Simulated start timestamp (defaults to now).

        Returns
        -------
        dict matching the device-connector API format.
        """
        if activity_profile == "default":
            activity_profile = _default_day_profile()

        if start_time is None:
            start_time = datetime.now(timezone.utc)

        # Reset recovery state
        self._recovery_start_hr = None
        self._recovery_start_time = None
        self._last_activity = None

        records: list[BiometricRecord] = []

        # We simulate at 1-second resolution for internal state, but
        # only emit samples at device-specific intervals.
        t = 0  # seconds from start
        total_seconds = duration_minutes * 60

        # Accumulators for step batching
        step_accumulator = 0.0

        # Timestamps for last emission per metric
        last_hr_emit = -999
        last_spo2_emit = -999
        last_step_emit = -999
        last_temp_emit = -999

        while t < total_seconds:
            minute = t / 60.0
            current_mode = self._resolve_activity(minute, activity_profile)
            hour_of_day = (start_time.hour + minute / 60.0) % 24.0

            # --- Heart Rate ---
            hr_sample_interval = (
                self.profile.hr_sample_interval_active
                if current_mode in (ActivityMode.WALKING, ActivityMode.RUNNING, ActivityMode.WORKOUT)
                else self.profile.hr_sample_interval_rest
            )

            target_hr = self._compute_target_hr(minute, current_mode, hour_of_day)
            hr_value = self._apply_hr_dynamics(target_hr, t, current_mode)

            if t - last_hr_emit >= hr_sample_interval:
                records.append(BiometricRecord(
                    metric_type="heart_rate",
                    value=round(hr_value, 1),
                    timestamp=self._ts(start_time, t),
                    quality=self._hr_quality(hr_value),
                ))
                last_hr_emit = t

            # --- SpO2 ---
            spo2_target = self._compute_spo2_target(current_mode, hour_of_day)
            spo2_value = self._apply_spo2_dynamics(spo2_target, t)

            if t - last_spo2_emit >= self.profile.spo2_sample_interval:
                records.append(BiometricRecord(
                    metric_type="spo2",
                    value=round(spo2_value, 1),
                    timestamp=self._ts(start_time, t),
                    quality=self._spo2_quality(spo2_value),
                ))
                last_spo2_emit = t

            # --- Steps ---
            steps_per_sec = self._compute_steps_per_sec(current_mode, t)
            step_accumulator += steps_per_sec

            if t - last_step_emit >= self.profile.step_batch_interval:
                batch_steps = max(0, int(step_accumulator))
                if batch_steps > 0:
                    records.append(BiometricRecord(
                        metric_type="steps",
                        value=float(batch_steps),
                        timestamp=self._ts(start_time, t),
                        quality="good",
                    ))
                step_accumulator = 0.0
                last_step_emit = t

            # --- Temperature ---
            should_emit_temp = (t - last_temp_emit >= self.profile.temp_sample_interval)
            if should_emit_temp:
                temp_value = self._compute_temperature(current_mode, hour_of_day, t)
                records.append(BiometricRecord(
                    metric_type="temperature",
                    value=round(temp_value, 2),
                    timestamp=self._ts(start_time, t),
                    quality="good",
                ))
                last_temp_emit = t

            # --- Blood Pressure (Huawei Watch D2 only) ---
            if self.profile.has_blood_pressure and should_emit_temp:
                sys_bp, dia_bp = self._compute_blood_pressure(current_mode, hour_of_day, t)
                records.append(BiometricRecord(
                    metric_type="blood_pressure_systolic",
                    value=round(sys_bp, 1),
                    timestamp=self._ts(start_time, t),
                    quality="good",
                ))
                records.append(BiometricRecord(
                    metric_type="blood_pressure_diastolic",
                    value=round(dia_bp, 1),
                    timestamp=self._ts(start_time, t),
                    quality="good",
                ))

            self._last_activity = current_mode
            t += 1  # 1-second tick

        # Flush remaining step accumulator
        if step_accumulator > 0:
            records.append(BiometricRecord(
                metric_type="steps",
                value=float(max(0, int(step_accumulator))),
                timestamp=self._ts(start_time, total_seconds - 1),
                quality="good",
            ))

        sync_interval_ms = self.profile.hr_sample_interval_active * 1000

        return {
            "device_type": self.profile.name,
            "device_token": str(uuid.uuid4()),
            "sync_interval_ms": sync_interval_ms,
            "records": [
                {"metric_type": r.metric_type, "value": r.value, "timestamp": r.timestamp, "quality": r.quality}
                for r in records
            ],
        }

    def generate_continuous_stream(
        self,
        mode: ActivityMode,
        duration_minutes: int = 10,
        start_time: datetime | None = None,
    ) -> list[BiometricRecord]:
        """
        Generate a continuous stream in a single activity mode.
        Useful for testing individual scenarios.
        """
        if start_time is None:
            start_time = datetime.now(timezone.utc)

        profile = [(0, duration_minutes, mode)]
        return self.generate_session(
            duration_minutes=duration_minutes,
            activity_profile=profile,
            start_time=start_time,
        )["records"]

    # ------------------------------------------------------------------
    # Heart rate computation
    # ------------------------------------------------------------------

    def _resolve_activity(
        self,
        minute: float,
        profile: list[tuple[int, int, ActivityMode]],
    ) -> ActivityMode:
        for start, end, mode in profile:
            if start <= minute < end:
                return mode
        return ActivityMode.RESTING

    def _compute_target_hr(
        self,
        minute: float,
        mode: ActivityMode,
        hour_of_day: float,
    ) -> float:
        """Compute the ideal / target HR before dynamics and noise."""
        base = self.resting_hr

        if mode == ActivityMode.SLEEPING:
            # Deep sleep: 50-65 BPM, with circadian dip
            sleep_depth = self._sleep_depth_factor(minute)
            return 50.0 + 15.0 * (1.0 - sleep_depth)

        if mode == ActivityMode.RESTING:
            # Circadian-modulated resting
            circadian = self._circadian_factor(hour_of_day)
            return base + 10.0 * circadian

        if mode == ActivityMode.WALKING:
            # 90-110 BPM
            return 90.0 + 20.0 * self._rng.uniform(0.3, 0.7)

        if mode == ActivityMode.RUNNING:
            # 150-180 BPM, scaled by fitness
            intensity = 0.75 + 0.15 * self._rng.uniform(0, 1)
            hr_range = self.hr_max - base
            return base + hr_range * intensity

        if mode == ActivityMode.WORKOUT:
            return self._workout_hr_profile(minute, base)

        return base

    def _workout_hr_profile(self, minute: float, base_hr: float) -> float:
        """
        Realistic workout HR: warm-up ramp, peak plateau, cool-down, recovery.

        Assumes the WORKOUT segment is ~60 min. We normalise to 0-1 phase.
        """
        # Find workout start by scanning backwards
        workout_start = 0.0
        workout_end = 60.0
        # We use the last_activity reset to detect session start
        if self._recovery_start_time is None:
            self._recovery_start_time = minute
            self._recovery_start_hr = base_hr

        elapsed = minute - self._recovery_start_time
        phase = elapsed / 60.0  # 0..1 over the workout
        phase = max(0.0, min(1.0, phase))

        hr_range = self.hr_max - base_hr

        if phase < 0.15:
            # Warm-up: gradual ramp from resting to ~60% HR reserve
            t = phase / 0.15
            return base_hr + hr_range * 0.6 * _smoothstep(t)
        elif phase < 0.25:
            # Ramp to peak
            t = (phase - 0.15) / 0.10
            return base_hr + hr_range * (0.6 + 0.30 * _smoothstep(t))
        elif phase < 0.75:
            # Peak plateau with slight variability
            peak = base_hr + hr_range * 0.88
            wobble = 5.0 * math.sin(2 * math.pi * phase * 8)
            return peak + wobble
        elif phase < 0.90:
            # Cool-down ramp down
            t = (phase - 0.75) / 0.15
            cool_start = base_hr + hr_range * 0.88
            return cool_start - (cool_start - base_hr) * 0.5 * _smoothstep(t)
        else:
            # Final cool-down to near-resting
            t = (phase - 0.90) / 0.10
            return base_hr + hr_range * 0.35 * (1.0 - _smoothstep(t))

    def _apply_hr_dynamics(
        self,
        target_hr: float,
        t_seconds: int,
        mode: ActivityMode,
    ) -> float:
        """
        Apply physiological dynamics: smoothing, noise, device accuracy,
        and exponential recovery decay.
        """
        # Smoothed noise (Perlin-like)
        noise_sec = t_seconds / 60.0  # noise grid per minute
        macro_noise = 4.0 * self._hr_noise(noise_sec)
        micro_noise = 1.5 * self._hr_micro_noise(t_seconds / 10.0)

        # Respiratory sinusoidal component (~0.25 Hz = 15 breaths/min)
        resp_freq = 0.25
        resp_amplitude = 2.0 if mode == ActivityMode.RESTING else 4.0
        resp_component = resp_amplitude * math.sin(2 * math.pi * resp_freq * t_seconds)

        value = target_hr + macro_noise + micro_noise + resp_component

        # Device accuracy error
        device_error = self._rng.gauss(0, self.profile.hr_accuracy_bpm * 0.5)
        value += device_error

        # Exponential recovery decay after exercise
        if mode in (ActivityMode.RESTING, ActivityMode.SLEEPING) and self._last_activity in (
            ActivityMode.RUNNING, ActivityMode.WORKOUT
        ):
            # HR recovery: ~30 BPM drop in first minute, then slower
            if self._recovery_start_hr is None:
                self._recovery_start_hr = value
                self._recovery_start_time = t_seconds
            if self._recovery_start_time is not None:
                recovery_elapsed = t_seconds - self._recovery_start_time
                if recovery_elapsed > 0:
                    # Two-phase exponential: fast then slow
                    fast_decay = 20.0 * math.exp(-recovery_elapsed / 30.0)
                    slow_decay = 15.0 * math.exp(-recovery_elapsed / 300.0)
                    value = self.resting_hr + fast_decay + slow_decay + macro_noise + resp_component

        # Clamp to physiological limits
        value = max(40.0, min(self.hr_max + 5, value))

        return value

    def _hr_quality(self, hr: float) -> str:
        if hr < 45 or hr > self.hr_max + 3:
            return "poor"
        if hr < 55 or hr > self.hr_max - 5:
            return "moderate"
        return "good"

    # ------------------------------------------------------------------
    # SpO2 computation
    # ------------------------------------------------------------------

    def _compute_spo2_target(self, mode: ActivityMode, hour_of_day: float) -> float:
        if mode == ActivityMode.SLEEPING:
            return 97.0 + self._rng.uniform(-0.5, 0.5)
        if mode == ActivityMode.RESTING:
            return 98.0 + self._rng.uniform(-0.5, 0.5)
        if mode == ActivityMode.WALKING:
            return 96.5 + self._rng.uniform(-0.5, 0.5)
        if mode in (ActivityMode.RUNNING, ActivityMode.WORKOUT):
            # SpO2 can drop during intense exercise
            return 94.0 + self._rng.uniform(-1.0, 1.0)
        return 97.5

    def _apply_spo2_dynamics(self, target: float, t_seconds: int) -> float:
        """SpO2 changes slowly -- heavy smoothing / inertia."""
        noise = 0.3 * self._spo2_noise(t_seconds / 60.0)
        value = target + noise

        # Hard physiological clamp: never below 88%
        value = max(88.0, min(100.0, value))
        return value

    def _spo2_quality(self, spo2: float) -> str:
        if spo2 < 90:
            return "poor"
        if spo2 < 94:
            return "moderate"
        return "good"

    # ------------------------------------------------------------------
    # Steps
    # ------------------------------------------------------------------

    def _compute_steps_per_sec(self, mode: ActivityMode, t_seconds: int) -> float:
        if mode == ActivityMode.RESTING:
            return 0.0
        if mode == ActivityMode.SLEEPING:
            return 0.0

        if mode == ActivityMode.WALKING:
            cadence = 110.0 + 15.0 * self._hr_noise(t_seconds / 30.0)
            return max(0, cadence / 60.0)

        if mode == ActivityMode.RUNNING:
            cadence = 175.0 + 10.0 * self._hr_noise(t_seconds / 30.0)
            return max(0, cadence / 60.0)

        if mode == ActivityMode.WORKOUT:
            # Variable cadence depending on workout phase
            if self._recovery_start_time is not None:
                elapsed = t_seconds / 60.0 - self._recovery_start_time
                phase = max(0.0, min(1.0, elapsed / 60.0))
            else:
                phase = 0.5

            if phase < 0.15:
                cadence = 100.0 + 40.0 * (phase / 0.15)
            elif phase < 0.75:
                cadence = 165.0 + 15.0 * math.sin(2 * math.pi * phase * 6)
            else:
                cadence = 165.0 * (1.0 - (phase - 0.75) / 0.25)
            return max(0, cadence / 60.0)

        return 0.0

    # ------------------------------------------------------------------
    # Temperature
    # ------------------------------------------------------------------

    def _compute_temperature(
        self,
        mode: ActivityMode,
        hour_of_day: float,
        t_seconds: int,
    ) -> float:
        """Core body temperature with circadian rhythm and exercise elevation."""
        # Circadian: lowest ~4am, highest ~6pm
        circadian = 36.4 + 0.3 * math.sin(2 * math.pi * (hour_of_day - 4) / 24.0)

        # Exercise elevation
        exercise_delta = 0.0
        if mode in (ActivityMode.RUNNING, ActivityMode.WORKOUT):
            exercise_delta = 0.4 + 0.3 * self._rng.uniform(0, 1)
        elif mode == ActivityMode.WALKING:
            exercise_delta = 0.15

        noise = 0.05 * self._temp_noise(t_seconds / 60.0)

        return circadian + exercise_delta + noise

    # ------------------------------------------------------------------
    # Blood Pressure (Huawei Watch D2)
    # ------------------------------------------------------------------

    def _compute_blood_pressure(
        self,
        mode: ActivityMode,
        hour_of_day: float,
        t_seconds: int,
    ) -> tuple[float, float]:
        """Simplified BP model: elevated during exercise, normal at rest."""
        base_sys = 120.0
        base_dia = 80.0

        # Circadian variation
        circadian_sys = 5.0 * math.sin(2 * math.pi * (hour_of_day - 6) / 24.0)
        circadian_dia = 3.0 * math.sin(2 * math.pi * (hour_of_day - 6) / 24.0)

        # Exercise effect
        if mode == ActivityMode.RUNNING:
            base_sys += 40.0
            base_dia += 10.0
        elif mode == ActivityMode.WORKOUT:
            base_sys += 30.0
            base_dia += 8.0
        elif mode == ActivityMode.WALKING:
            base_sys += 15.0
            base_dia += 5.0

        noise_sys = 2.0 * self._bp_sys_noise(t_seconds / 60.0)
        noise_dia = 1.5 * self._bp_dia_noise(t_seconds / 60.0)

        sys_bp = base_sys + circadian_sys + noise_sys
        dia_bp = base_dia + circadian_dia + noise_dia

        return max(90, sys_bp), max(55, dia_bp)

    # ------------------------------------------------------------------
    # Circadian / sleep helpers
    # ------------------------------------------------------------------

    @staticmethod
    def _circadian_factor(hour: float) -> float:
        """Returns 0..1 circadian multiplier (low at night, high midday)."""
        # Sinusoid: min at 3am, max at 3pm
        return 0.5 + 0.5 * math.sin(2 * math.pi * (hour - 3) / 24.0)

    @staticmethod
    def _sleep_depth_factor(minute: float) -> float:
        """0 = light sleep / awake, 1 = deep sleep. Peaks in middle of night."""
        # Assume sleep starts ~minute 0, deepest around minute 240 (4am)
        phase = (minute % 480) / 480.0  # wrap every 8h
        return math.sin(math.pi * phase)

    # ------------------------------------------------------------------
    # Timestamp helpers
    # ------------------------------------------------------------------

    @staticmethod
    def _ts(start: datetime, offset_seconds: int) -> str:
        ts = start + timedelta(seconds=offset_seconds)
        return ts.isoformat(timespec="milliseconds")
