// Procedural audio synthesizer using Web Audio API and Native Desktop Notifications

class AlertManager {
  private audioCtx: AudioContext | null = null;
  private soundEnabled: boolean = false;
  private desktopEnabled: boolean = false;
  private lastSoundTime: number = 0;
  private lastNotificationTime: number = 0;
  private readonly THROTTLE_MS = 1500;

  constructor() {
    this.soundEnabled = localStorage.getItem('timberlens_sound_enabled') === 'true';
    this.desktopEnabled = localStorage.getItem('timberlens_desktop_enabled') === 'true';
  }

  public isSoundEnabled(): boolean {
    return this.soundEnabled;
  }

  public isDesktopEnabled(): boolean {
    return this.desktopEnabled;
  }

  public setSoundEnabled(enabled: boolean) {
    this.soundEnabled = enabled;
    localStorage.setItem('timberlens_sound_enabled', String(enabled));
    if (enabled && !this.audioCtx) {
      this.initAudio();
    }
  }

  public async setDesktopEnabled(enabled: boolean): Promise<boolean> {
    if (enabled) {
      if (!('Notification' in window)) {
        return false;
      }
      if (Notification.permission !== 'granted') {
        const perm = await Notification.requestPermission();
        if (perm !== 'granted') {
          this.desktopEnabled = false;
          localStorage.setItem('timberlens_desktop_enabled', 'false');
          return false;
        }
      }
    }
    this.desktopEnabled = enabled;
    localStorage.setItem('timberlens_desktop_enabled', String(enabled));
    return true;
  }

  private initAudio() {
    try {
      const AudioCtxClass = window.AudioContext || (window as any).webkitAudioContext;
      if (AudioCtxClass) {
        this.audioCtx = new AudioCtxClass();
      }
    } catch {
      // AudioContext not supported
    }
  }

  /**
   * Discretely synthesizes a warm error chord (harmonic double chime).
   */
  public triggerErrorAlert(service: string, message: string) {
    const now = Date.now();

    // 1. Play Web Audio chime if enabled and throttled
    if (this.soundEnabled && now - this.lastSoundTime > this.THROTTLE_MS) {
      this.lastSoundTime = now;
      this.playErrorChime();
    }

    // 2. Trigger desktop notification if enabled, tab in background, and throttled
    if (
      this.desktopEnabled &&
      document.hidden &&
      'Notification' in window &&
      Notification.permission === 'granted' &&
      now - this.lastNotificationTime > this.THROTTLE_MS
    ) {
      this.lastNotificationTime = now;
      try {
        new Notification(`TimberLens: Error in ${service}`, {
          body: message.slice(0, 140),
          icon: '/favicon.ico',
        });
      } catch {
        // Notification failed or blocked
      }
    }
  }

  public playErrorChime() {
    try {
      if (!this.audioCtx) {
        this.initAudio();
      }
      if (!this.audioCtx) return;

      if (this.audioCtx.state === 'suspended') {
        this.audioCtx.resume();
      }

      const ctx = this.audioCtx;
      const now = ctx.currentTime;

      // Note 1: 587.33 Hz (D5)
      const osc1 = ctx.createOscillator();
      const gain1 = ctx.createGain();
      osc1.type = 'sine';
      osc1.frequency.setValueAtTime(587.33, now);
      gain1.gain.setValueAtTime(0.08, now);
      gain1.gain.exponentialRampToValueAtTime(0.001, now + 0.35);
      osc1.connect(gain1);
      gain1.connect(ctx.destination);
      osc1.start(now);
      osc1.stop(now + 0.35);

      // Note 2: 880.00 Hz (A5) slightly delayed
      const osc2 = ctx.createOscillator();
      const gain2 = ctx.createGain();
      osc2.type = 'triangle';
      osc2.frequency.setValueAtTime(880.00, now + 0.08);
      gain2.gain.setValueAtTime(0.06, now + 0.08);
      gain2.gain.exponentialRampToValueAtTime(0.001, now + 0.45);
      osc2.connect(gain2);
      gain2.connect(ctx.destination);
      osc2.start(now + 0.08);
      osc2.stop(now + 0.45);
    } catch {
      // Audio synthesis failed
    }
  }

  public playWarningBeep() {
    try {
      if (!this.audioCtx) {
        this.initAudio();
      }
      if (!this.audioCtx) return;

      if (this.audioCtx.state === 'suspended') {
        this.audioCtx.resume();
      }

      const ctx = this.audioCtx;
      const now = ctx.currentTime;

      const osc = ctx.createOscillator();
      const gain = ctx.createGain();
      osc.type = 'sine';
      osc.frequency.setValueAtTime(440, now); // A4
      gain.gain.setValueAtTime(0.05, now);
      gain.gain.exponentialRampToValueAtTime(0.001, now + 0.2);
      osc.connect(gain);
      gain.connect(ctx.destination);
      osc.start(now);
      osc.stop(now + 0.2);
    } catch {
      // Audio synthesis failed
    }
  }
}

export const alertManager = new AlertManager();
