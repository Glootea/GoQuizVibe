(function() {
    let timerInterval = null;
    let syncInterval = null;

    function formatTime(seconds) {
        const mins = Math.floor(seconds / 60);
        const secs = seconds % 60;
        return (mins < 10 ? '0' : '') + mins + ':' + (secs < 10 ? '0' : '') + secs;
    }

    function getTimerData() {
        const el = document.getElementById('timer-display');
        if (!el) return null;
        return {
            quizId: el.dataset.quizId,
            sessionId: el.dataset.sessionId,
            remainingSeconds: parseInt(el.dataset.remainingSeconds) || 0
        };
    }

    function updateTimer() {
        const data = getTimerData();
        if (!data) return;

        if (data.remainingSeconds > 0) {
            data.remainingSeconds--;
            const timerText = document.getElementById('timer-text');
            if (timerText) timerText.textContent = formatTime(data.remainingSeconds);
        }
    }

    function syncRemainingSeconds() {
        const data = getTimerData();
        if (!data) return;

        fetch('/quiz/' + data.quizId + '/sync?session=' + data.sessionId)
            .then(r => r.json())
            .then(d => {
                if (d.remaining_seconds !== undefined) {
                    const timerDisplay = document.getElementById('timer-display');
                    if (timerDisplay) timerDisplay.dataset.remainingSeconds = d.remaining_seconds;
                    const timerText = document.getElementById('timer-text');
                    if (timerText) timerText.textContent = formatTime(d.remaining_seconds);
                }
            }).catch(function() {});
    }

    function initTimer() {
        clearInterval(timerInterval);
        clearInterval(syncInterval);

        const data = getTimerData();
        if (!data) return;

        timerInterval = setInterval(updateTimer, 1000);
        syncInterval = setInterval(syncRemainingSeconds, 30000);
    }

    document.body.addEventListener('htmx:beforeSwap', function() {
        clearInterval(timerInterval);
        clearInterval(syncInterval);
    });

    document.body.addEventListener('htmx:afterSwap', function(evt) {
        if (document.getElementById('timer-display')) {
            initTimer();
        }
    });

    document.addEventListener('DOMContentLoaded', initTimer);
})();