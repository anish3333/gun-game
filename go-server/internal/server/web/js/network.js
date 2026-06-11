import { state } from './state.js';
import * as ui from './ui.js';
import * as renderer from './renderer.js';

export async function connect() {
  ui.showConnecting();

  // 1. Check for an existing token in the browser
  let token = localStorage.getItem('recoil_token');

  // 2. If no token, fetch a new guest identity from the Go API
  if (!token) {
    try {
      const res = await fetch('http://localhost:3000/api/init-guest', { method: 'POST' });
      if (!res.ok) throw new Error('Failed to fetch guest token');
      const data = await res.json();
      token = data.token;
      localStorage.setItem('recoil_token', token);
      console.log(`Registered as new guest: ${data.display_name}`);
    } catch (err) {
      ui.showDisconnected('Authentication failed. Is the server running?');
      return;
    }
  }

  // 3. Connect to WebSocket with token in the URL
  state.ws = new WebSocket(`ws://localhost:3000/ws?token=${token}`);

  state.ws.onopen = () => {
    console.log('connected to server');
  };

  state.ws.onmessage = (e) => {
    let msg;
    try { msg = JSON.parse(e.data); } catch { return; }
    handleMessage(msg);
  };

  state.ws.onclose = (e) => {
    if (e.code === 1008 || (e.reason && e.reason.includes('Unauthorized'))) {
      localStorage.removeItem('recoil_token');
      ui.showDisconnected('Session expired. Reloading...');
      setTimeout(() => location.reload(), 2000);
    } else {
      ui.showDisconnected('disconnected · reload to reconnect');
    }
  };

  state.ws.onerror = () => {
    ui.showDisconnected('cannot reach server · is it running on localhost:3000?');
  };
}

export function send(msg) {
  if (state.ws && state.ws.readyState === 1) {
    state.ws.send(JSON.stringify(msg));
  }
}

export function startNetworkLoops() {
  // Input loop — 60hz
  setInterval(() => {
    if (!state.ws || state.ws.readyState !== 1 || !state.myId || state.gamePhase !== 'in-game') return;
    send({ type: 'input', angle: state.mouseAngle, shoot: state.shooting });
  }, 1000 / 60);

  // Ping loop — every 2s
  setInterval(() => {
    if (state.ws && state.ws.readyState === 1) {
      state.pingStart = performance.now();
      send({ type: 'ping' });
    }
  }, 2000);

  // Lobby refresh — every 3s while in lobby
  setInterval(() => {
    if (state.gamePhase === 'lobby') {
      send({ type: 'list_rooms' });
    }
  }, 3000);
}

function handleMessage(msg) {
  switch (msg.type) {

    case 'hello':
      state.gamePhase = 'lobby';
      state.weaponDefs = msg.weapons;
      state.availableWeapons = Object.keys(msg.weapons);
      ui.showLobby(msg.weapons, msg.rooms || []);
      break;

    case 'room_created':
      state.myId        = msg.playerId;
      state.myWeapon    = msg.weapon;
      state.currentRoom = { code: msg.code };
      state.gamePhase   = 'waiting';
      ui.showRoomWaiting(msg.code);
      break;

    case 'room_joined':
      state.myId        = msg.playerId;
      state.myWeapon    = msg.weapon;
      state.currentRoom = { code: msg.code };
      state.gamePhase   = 'waiting';
      ui.showRoomWaiting(msg.code);
      break;

    case 'room_list':
      ui.updateRoomList(msg.rooms || []);
      break;

    case 'match_start':
      state.gamePhase = 'in-game';
      ui.showMatchStart(msg.players);
      break;

    case 'snapshot':
      state.gameState = msg;
      if (state.gamePhase === 'in-game') ui.updateHUD();
      break;

    case 'hit':
      renderer.spawnHitNumber(msg.playerId, msg.damage);
      renderer.spawnSparks(msg.playerId, msg.damage);
      break;

    case 'player_died':
      renderer.spawnDeathEffect(msg.playerId);
      break;

    case 'player_respawned':
      // client handles visually via snapshot
      break;

    case 'match_over':
      state.gamePhase = 'finished';
      ui.showMatchOver(msg.winnerId);
      break;

    case 'opponent_disconnected':
      state.gamePhase = 'lobby';
      ui.showDisconnected('opponent disconnected · reload to find a new match');
      break;

    case 'pong':
      state.currentPing = Math.round(performance.now() - state.pingStart);
      document.getElementById('ping-display').textContent = `ping: ${state.currentPing}ms`;
      break;

    case 'error':
      ui.showError(msg.message);
      break;
  }
}