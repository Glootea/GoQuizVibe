(() => {
  let timerInterval = null;

  function formatTime(seconds) {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins < 10 ? "0" : ""}${mins}:${secs < 10 ? "0" : ""}${secs}`;
  }

  function getTimerData() {
    const el = document.getElementById("timer-display");
    if (!el) return null;
    return {
      quizId: el.dataset.quizId,
      sessionId: el.dataset.sessionId,
      remainingSeconds: parseInt(el.dataset.remainingSeconds, 10) || 0,
    };
  }

  function updateTimer() {
    const data = getTimerData();
    if (!data) return;

    if (data.remainingSeconds > 0) {
      data.remainingSeconds--;
      const timerDisplay = document.getElementById("timer-display");
      if (timerDisplay)
        timerDisplay.dataset.remainingSeconds = data.remainingSeconds;
      const timerText = document.getElementById("timer-text");
      if (timerText) timerText.textContent = formatTime(data.remainingSeconds);
    }
  }

  function initTimer() {
    clearInterval(timerInterval);
    clearInterval(null);

    const data = getTimerData();
    if (!data) return;

    timerInterval = setInterval(updateTimer, 1000);
  }

  function setupHTMXHandlers() {
    if (document.body) {
      document.body.addEventListener("htmx:beforeSwap", () => {
        clearInterval(timerInterval);
        clearInterval(null);
      });

      document.body.addEventListener("htmx:afterSwap", () => {
        if (document.getElementById("timer-display")) {
          initTimer();
        }
      });
    }
  }

  function init() {
    initTimer();
    setupHTMXHandlers();
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
