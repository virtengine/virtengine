import type { LivenessChallengeResult, LivenessResult } from "../captureModels";
import { createId } from "../../utils/id";
import type { FaceSignal, LivenessChallenge, LivenessChallengeType, LivenessConfig, LivenessUpdate } from "./types";

interface ChallengeState {
  challenge: LivenessChallenge;
  startedAt: number;
  completedAt?: number;
  completed: boolean;
  attempts: number;
  notes?: string;
}

interface BlinkState {
  lastEyesOpen?: boolean;
  closedAt?: number;
  openedAt?: number;
}

interface HoldStillState {
  stableSince?: number;
}

export const DEFAULT_LIVENESS_CONFIG: LivenessConfig = {
  minFaceConfidence: 0.6,
  blinkEyeClosedThreshold: 0.25,
  blinkEyeOpenThreshold: 0.7,
  yawThreshold: 15,
  smileThreshold: 0.6,
  holdStillDurationMs: 1200
};

export function createDefaultChallenges(): LivenessChallenge[] {
  return [
    {
      id: createId("challenge"),
      type: "blink",
      instruction: "Blink once",
      timeoutMs: 6000
    },
    {
      id: createId("challenge"),
      type: "turn_left",
      instruction: "Turn your head left",
      timeoutMs: 6000
    },
    {
      id: createId("challenge"),
      type: "turn_right",
      instruction: "Turn your head right",
      timeoutMs: 6000
    },
    {
      id: createId("challenge"),
      type: "smile",
      instruction: "Smile for the camera",
      timeoutMs: 6000
    },
    {
      id: createId("challenge"),
      type: "hold_still",
      instruction: "Hold still",
      timeoutMs: 4000
    }
  ];
}

export class LivenessEngine {
  private readonly challenges: LivenessChallenge[];
  private readonly config: LivenessConfig;
  private readonly results: ChallengeState[] = [];
  private startedAt = 0;
  private currentIndex = 0;
  private blinkState: BlinkState = {};
  private holdStillState: HoldStillState = {};

  constructor(challenges: LivenessChallenge[], config: LivenessConfig = DEFAULT_LIVENESS_CONFIG) {
    this.challenges = challenges;
    this.config = config;
  }

  start(): void {
    this.startedAt = Date.now();
    this.currentIndex = 0;
    this.results.length = 0;
    this.blinkState = {};
    this.holdStillState = {};
  }

  update(signal: FaceSignal): LivenessUpdate {
    if (!this.startedAt) {
      this.start();
    }

    const current = this.ensureCurrentChallenge();

    if (current.completed) {
      return this.advanceIfNeeded(signal);
    }

    if (signal.faceConfidence < this.config.minFaceConfidence) {
      return this.buildUpdate(current, false, 0, "Face not stable");
    }

    if (signal.timestamp - current.startedAt > current.challenge.timeoutMs) {
      current.completed = false;
      current.notes = "timeout";
      current.completedAt = signal.timestamp;
      return this.advanceIfNeeded(signal);
    }

    switch (current.challenge.type) {
      case "blink":
        return this.handleBlink(current, signal);
      case "turn_left":
        return this.handleTurn(current, signal, "turn_left");
      case "turn_right":
        return this.handleTurn(current, signal, "turn_right");
      case "smile":
        return this.handleSmile(current, signal);
      case "hold_still":
        return this.handleHoldStill(current, signal);
      default:
        return this.buildUpdate(current, false, 0, "Unsupported challenge");
    }
  }

  getResult(): LivenessResult {
    const completedAt = this.results[this.results.length - 1]?.completedAt ?? Date.now();
    const challengeResults: LivenessChallengeResult[] = this.results.map((state) => ({
      challengeId: state.challenge.id,
      type: state.challenge.type,
      completed: state.completed,
      startedAt: state.startedAt,
      completedAt: state.completedAt,
      attempts: state.attempts,
      notes: state.notes
    }));

    const passed = challengeResults.length > 0 && challengeResults.every((result) => result.completed);
    const score = passed
      ? Math.round((challengeResults.filter((r) => r.completed).length / challengeResults.length) * 100)
      : Math.round((challengeResults.filter((r) => r.completed).length / Math.max(challengeResults.length, 1)) * 60);

    return {
      passed,
      score,
      startedAt: this.startedAt || Date.now(),
      completedAt,
      challenges: challengeResults,
      failureReason: passed ? undefined : "liveness_failed"
    };
  }

  private ensureCurrentChallenge(): ChallengeState {
    const challenge = this.challenges[this.currentIndex];
    const existing = this.results[this.currentIndex];

    if (existing) {
      return existing;
    }

    const state: ChallengeState = {
      challenge,
      startedAt: Date.now(),
      completed: false,
      attempts: 0
    };

    this.results[this.currentIndex] = state;
    return state;
  }

  private advanceIfNeeded(signal: FaceSignal): LivenessUpdate {
    const current = this.results[this.currentIndex];
    if (current && current.completed) {
      this.currentIndex = Math.min(this.currentIndex + 1, this.challenges.length - 1);
    } else if (current && !current.completed && current.notes === "timeout") {
      this.currentIndex = Math.min(this.currentIndex + 1, this.challenges.length - 1);
    }

    const next = this.ensureCurrentChallenge();
    return this.buildUpdate(next, next.completed, next.completed ? 1 : 0, next.notes);
  }

  private handleBlink(current: ChallengeState, signal: FaceSignal): LivenessUpdate {
    const left = signal.leftEyeOpenProbability ?? 1;
    const right = signal.rightEyeOpenProbability ?? 1;
    const eyesOpen = Math.min(left, right) >= this.config.blinkEyeOpenThreshold;
    const eyesClosed = Math.max(left, right) <= this.config.blinkEyeClosedThreshold;

    current.attempts += 1;

    if (eyesClosed && this.blinkState.lastEyesOpen !== false) {
      this.blinkState.closedAt = signal.timestamp;
      this.blinkState.lastEyesOpen = false;
    }

    if (eyesOpen && this.blinkState.lastEyesOpen === false) {
      this.blinkState.openedAt = signal.timestamp;
      const closedDuration = (this.blinkState.openedAt ?? 0) - (this.blinkState.closedAt ?? 0);
      if (closedDuration > 80 && closedDuration < 1000) {
        current.completed = true;
        current.completedAt = signal.timestamp;
      }
      this.blinkState.lastEyesOpen = true;
    }

    const progress = current.completed ? 1 : eyesClosed ? 0.5 : 0.2;
    return this.buildUpdate(current, current.completed, progress, current.notes);
  }

  private handleTurn(current: ChallengeState, signal: FaceSignal, type: LivenessChallengeType): LivenessUpdate {
    const threshold = this.config.yawThreshold;
    current.attempts += 1;

    const yaw = signal.yaw;
    const matches = type === "turn_left" ? yaw <= -threshold : yaw >= threshold;

    if (matches) {
      current.completed = true;
      current.completedAt = signal.timestamp;
    }

    const progress = matches ? 1 : Math.min(Math.abs(yaw) / threshold, 0.8);
    return this.buildUpdate(current, current.completed, progress, current.notes);
  }

  private handleSmile(current: ChallengeState, signal: FaceSignal): LivenessUpdate {
    current.attempts += 1;
    const smile = signal.smileProbability ?? 0;
    if (smile >= this.config.smileThreshold) {
      current.completed = true;
      current.completedAt = signal.timestamp;
    }

    const progress = Math.min(smile / this.config.smileThreshold, 1);
    return this.buildUpdate(current, current.completed, progress, current.notes);
  }

  private handleHoldStill(current: ChallengeState, signal: FaceSignal): LivenessUpdate {
    current.attempts += 1;
    const stable = Math.abs(signal.yaw) < 5 && Math.abs(signal.pitch) < 5 && Math.abs(signal.roll) < 5;

    if (stable) {
      if (!this.holdStillState.stableSince) {
        this.holdStillState.stableSince = signal.timestamp;
      }
      const elapsed = signal.timestamp - (this.holdStillState.stableSince ?? signal.timestamp);
      if (elapsed >= this.config.holdStillDurationMs) {
        current.completed = true;
        current.completedAt = signal.timestamp;
      }
    } else {
      this.holdStillState.stableSince = undefined;
    }

    const progress = this.holdStillState.stableSince
      ? Math.min((signal.timestamp - (this.holdStillState.stableSince ?? signal.timestamp)) / this.config.holdStillDurationMs, 1)
      : 0;

    return this.buildUpdate(current, current.completed, progress, current.notes);
  }

  private buildUpdate(current: ChallengeState, completed: boolean, progress: number, notes?: string): LivenessUpdate {
    return {
      challengeId: current.challenge.id,
      challengeType: current.challenge.type,
      instruction: current.challenge.instruction,
      completed,
      progress,
      ...(notes ? { note: notes } : {})
    } as LivenessUpdate;
  }
}
