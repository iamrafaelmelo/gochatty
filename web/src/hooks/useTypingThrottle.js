import { useRef } from 'react';

import { TYPING_THROTTLE_INTERVAL_MS } from '../types/chat';

export function useTypingThrottle(sendTypingEvent) {
  const lastTypingSentAtRef = useRef(0);

  const sendThrottledTyping = (content) => {
    const now = Date.now();
    if (now - lastTypingSentAtRef.current < TYPING_THROTTLE_INTERVAL_MS) {
      return;
    }

    lastTypingSentAtRef.current = now;
    sendTypingEvent(content);
  };

  return { sendThrottledTyping };
}
