// Chart Configuration
const maxDataPoints = 60; // Keep 60 seconds of history on screen

const createChart = (ctxId, label, color) => {
    return new Chart(document.getElementById(ctxId).getContext('2d'), {
        type: 'line',
        data: {
            labels: Array(maxDataPoints).fill(''),
            datasets: [{ label: label, data: Array(maxDataPoints).fill(0), borderColor: color, tension: 0.2, borderWidth: 2, pointRadius: 0 }]
        },
        options: {
            responsive: true, maintainAspectRatio: false,
            animation: false, // Turn off animation for better performance on rapid updates
            scales: {
                y: { beginAtZero: true, grid: { color: '#333' } },
                x: { grid: { display: false } }
            },
            plugins: { legend: { labels: { color: '#eee' } } }
        }
    });
};

const tickChart = createChart('tickChart', 'Physics Tick Latency (µs)', '#a84af0');
const memChart = createChart('memChart', 'Heap Allocation (MB)', '#4af0c8');

// WebSocket Connection
let ws;
function connect() {
    ws = new WebSocket('ws://localhost:3000/admin/metrics');

    ws.onopen = () => {
        const el = document.getElementById('conn-status');
        el.textContent = '● LIVE';
        el.className = 'status connected';
    };

    ws.onmessage = (e) => {
        const data = JSON.parse(e.data);
        
        // 1. Update Top HUD
        document.getElementById('val-strategy').textContent = data.strategy;
        document.getElementById('val-encoding').textContent = data.encoding || 'json';
        document.getElementById('val-clients').textContent = data.clientCount;
        document.getElementById('val-rooms').textContent = data.roomCount;
        document.getElementById('val-goroutines').textContent = data.goroutines;
        document.getElementById('val-tick').textContent = data.tickTimeUs;

        // 2. Update Charts (Shift array left, push new value right)
        tickChart.data.datasets[0].data.shift();
        tickChart.data.datasets[0].data.push(data.tickTimeUs);
        tickChart.update();

        memChart.data.datasets[0].data.shift();
        memChart.data.datasets[0].data.push(data.heapAllocMb);
        memChart.update();
    };

    ws.onclose = () => {
        const el = document.getElementById('conn-status');
        el.textContent = 'Disconnected. Retrying...';
        el.className = 'status';
        setTimeout(connect, 2000);
    };
}

// REST API call to hot-swap the engine
window.swapEngine = async (type) => {
    try {
        await fetch(`http://localhost:3000/admin/strategy?type=${type}`, { method: 'POST' });
        console.log(`Requested engine swap to: ${type}`);
    } catch (err) {
        console.error('Failed to swap engine', err);
    }
};

window.swapEncoding = async (type) => {
    try {
        await fetch(`http://localhost:3000/admin/encoding?type=${type}`, { method: 'POST' });
        console.log(`Requested encoding swap to: ${type}`);
    } catch (err) {
        console.error('Failed to swap encoding', err);
    }
};

// Start
connect();