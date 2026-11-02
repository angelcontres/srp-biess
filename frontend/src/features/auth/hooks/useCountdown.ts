import { useCallback, useEffect, useRef, useState } from 'react';

export interface UseCountdownOptions {
  /** Initial time in seconds. */
  initialSeconds: number;
  /** Callback fired when the timer hits zero. */
  onFinish?: () => void;
  /** Whether the countdown starts automatically upon mount. Default is false. */
  autoStart?: boolean;
}

export interface UseCountdownReturn {
  secondsRemaining: number;
  isRunning: boolean;
  formattedTime: string;
  start: (seconds?: number) => void;
  stop: () => void;
  reset: () => void;
}

/**
 * Reactive countdown timer hook for security lockouts (e.g. 10-minute HTTP 423 lockouts).
 * Formats time as mm:ss with tabular numbers.
 */
export function useCountdown({
  initialSeconds,
  onFinish,
  autoStart = false,
}: UseCountdownOptions): UseCountdownReturn {
  const [secondsRemaining, setSecondsRemaining] = useState(initialSeconds);
  const [isRunning, setIsRunning] = useState(autoStart);
  const onFinishRef = useRef(onFinish);

  useEffect(() => {
    onFinishRef.current = onFinish;
  }, [onFinish]);

  const start = useCallback((seconds?: number) => {
    if (typeof seconds === 'number') {
      setSecondsRemaining(seconds);
    }
    setIsRunning(true);
  }, []);

  const stop = useCallback(() => {
    setIsRunning(false);
  }, []);

  const reset = useCallback(() => {
    setIsRunning(false);
    setSecondsRemaining(initialSeconds);
  }, [initialSeconds]);

  useEffect(() => {
    if (!isRunning) {
      return;
    }

    if (secondsRemaining <= 0) {
      setIsRunning(false);
      onFinishRef.current?.();
      return;
    }

    const interval = setInterval(() => {
      setSecondsRemaining((prev) => {
        if (prev <= 1) {
          clearInterval(interval);
          setIsRunning(false);
          onFinishRef.current?.();
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(interval);
  }, [isRunning, secondsRemaining]);

  const minutes = Math.floor(secondsRemaining / 60);
  const seconds = secondsRemaining % 60;
  const formattedTime = `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;

  return {
    secondsRemaining,
    isRunning,
    formattedTime,
    start,
    stop,
    reset,
  };
}
