import http from 'k6/http';
import ws from 'k6/ws';
import { sleep } from 'k6';

// Config via k6 __ENV (command line -e, shell export, or ./run-loadtest.sh + loadtest.env)
// See loadtest.env.sample
const BASE = __ENV.BASE_URL || 'http://localhost:3000';
const WS_BASE = BASE.replace(/^http/, 'ws');
const MATCH_DURATION_MS = Number(__ENV.MATCH_DURATION_MS || 35000);
const SCORE_LIMIT = Number(__ENV.SCORE_LIMIT || 3);
const STAGGER_MAX_S = Number(__ENV.STAGGER_MAX_S || 5);

export const options = {
  vus: Number(__ENV.VUS || 50),
  duration: __ENV.DURATION || '45s',
  thresholds: {
    ws_connecting: ['rate>0.95'],
    ws_msgs_received: ['count>0'],
  },
};

function initGuest() {
  const res = http.post(`${BASE}/api/init-guest`);
  if (res.status !== 200) {
    throw new Error(`init-guest failed: ${res.status}`);
  }
  return res.json();
}

function pickWeapon() {
  const weapons = ['pistol', 'smg', 'shotgun'];
  return weapons[Math.floor(Math.random() * weapons.length)];
}

function startCombatLoop(socket) {
  let angle = Math.random() * Math.PI * 2;
  let sweepDirection = Math.random() > 0.5 ? 1 : -1;

  socket.setInterval(() => {
    if (Math.random() < 0.05) {
      angle = Math.random() * Math.PI * 2;
      sweepDirection = Math.random() > 0.5 ? 1 : -1;
    } else {
      angle += 0.15 * sweepDirection;
      if (Math.random() < 0.1) sweepDirection *= -1;
    }

    socket.send(JSON.stringify({
      type: 'input',
      angle,
      shoot: Math.random() < 0.9,
    }));
  }, 33);
}

function handleGameMessages(socket, msg, ctx) {
  const { isHost, myId, matchStarted } = ctx;

  switch (msg.type) {
    case 'snapshot':
    case 'pong':
    case 'hit':
    case 'player_died':
    case 'player_respawned':
      return;

    case 'room_state': {
      const connected = (msg.players || []).filter((p) => p.connected).length;
      if (
        isHost &&
        !matchStarted.value &&
        msg.phase === 'lobby' &&
        connected >= 2 &&
        msg.hostId === myId
      ) {
        matchStarted.value = true;
        socket.send(JSON.stringify({ type: 'start_match' }));
      }
      break;
    }

    case 'match_start':
      startCombatLoop(socket);
      break;

    case 'match_results':
      socket.send(JSON.stringify({ type: 'leave_room' }));
      sleep(0.2);
      socket.close();
      break;

    case 'room_closed':
    case 'room_left':
      socket.close();
      break;

    case 'error':
      console.warn(`[${isHost ? 'host' : 'guest'}] ${msg.message}`);
      break;
  }
}

function connectPlayer(token, handlers) {
  ws.connect(`${WS_BASE}/ws?token=${token}`, null, (socket) => {
    socket.on('message', (raw) => {
      let msg;
      try {
        msg = JSON.parse(raw);
      } catch {
        return;
      }
      handlers.onMessage(socket, msg);
    });

    socket.setTimeout(() => socket.close(), MATCH_DURATION_MS);
  });
}

function spawnGuest(roomCode) {
  const guest = initGuest();

  connectPlayer(guest.token, {
    onMessage(socket, msg) {
      if (msg.type === 'hello') {
        socket.send(JSON.stringify({
          type: 'join_room',
          code: roomCode,
          weapon: pickWeapon(),
        }));
        return;
      }

      handleGameMessages(socket, msg, {
        isHost: false,
        myId: guest.player_id,
        matchStarted: { value: false },
      });
    },
  });
}

export default function () {
  sleep(Math.random() * STAGGER_MAX_S);

  const host = initGuest();
  const hostId = { value: host.player_id };
  const matchStarted = { value: false };
  let guestSpawned = false;

  connectPlayer(host.token, {
    onMessage(socket, msg) {
      if (msg.type === 'hello') {
        socket.send(JSON.stringify({
          type: 'create_room',
          weapon: pickWeapon(),
          scoreLimit: SCORE_LIMIT,
          timeLimit: 5,
          weaponMode: 'any',
          map: 'arena',
        }));
        return;
      }

      if (msg.type === 'room_created') {
        hostId.value = msg.playerId;
        if (!guestSpawned) {
          guestSpawned = true;
          // Defer so the host socket keeps processing messages
          socket.setTimeout(() => spawnGuest(msg.code), 50);
        }
        return;
      }

      handleGameMessages(socket, msg, {
        isHost: true,
        myId: hostId.value,
        matchStarted,
      });
    },
  });
}
