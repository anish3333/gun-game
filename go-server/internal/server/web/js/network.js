import { state, getBaseURL, getWSURL, getInviteCodeFromPath } from './state.js';
import { decodeWsData, encodeWsMessage, isBinaryEncoding } from './codec.js';
import * as ui from './ui.js';
import * as renderer from './renderer.js';

async function ensureGuestToken() {
  let token = localStorage.getItem('recoil_token');
  if (token) return token;

  const res = await fetch(`${getBaseURL()}/api/init-guest`, { method: 'POST' });
  if (!res.ok) throw new Error('Failed to fetch guest token');
  const data = await res.json();
  token = data.token;
  localStorage.setItem('recoil_token', token);
  state.myName = data.display_name;
  return token;
}

export async function connect() {
  const inviteCode = getInviteCodeFromPath();
  state.pendingJoinCode = inviteCode;

  if (inviteCode) {
    ui.showInviteJoining(inviteCode);
  } else {
    ui.showConnecting();
  }

  let token;
  try {
    token = await ensureGuestToken();
  } catch {
    ui.showDisconnected('Authentication failed. Is the server running?');
    return;
  }

  state.ws = new WebSocket(`${getWSURL()}/ws?token=${token}`);
  state.ws.binaryType = 'arraybuffer';

  state.ws.onopen = () => {};

  state.ws.onmessage = async (e) => {
    let msg;
    try {
      msg = await decodeWsData(e.data);
    } catch {
      return;
    }
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
    ui.showDisconnected('cannot reach server');
  };
}

export function send(msg) {
  if (!state.ws || state.ws.readyState !== 1) return;

  const payload = encodeWsMessage(msg, state.wireEncoding);
  if (isBinaryEncoding(state.wireEncoding)) {
    state.ws.send(payload);
  } else {
    state.ws.send(payload);
  }
}

function setWireEncoding(encoding) {
  state.wireEncoding = encoding === 'msgpack' ? 'msgpack' : 'json';
  ui.updateWireEncoding(state.wireEncoding);
}

export function startNetworkLoops() {
  setInterval(() => {
    if (!state.ws || state.ws.readyState !== 1 || !state.myId || state.isSpectator || state.gamePhase !== 'in-game') return;
    send({ type: 'input', angle: state.mouseAngle, shoot: state.shooting });
  }, 1000 / 60);

  setInterval(() => {
    if (state.ws && state.ws.readyState === 1) {
      state.pingStart = performance.now();
      send({ type: 'ping' });
    }
  }, 2000);
}

async function tryJoinFromURL() {
  const code = state.pendingJoinCode;
  if (!code) return;

  try {
    const res = await fetch(`${getBaseURL()}/api/room/${code}`);
    const info = await res.json();
    if (!info.exists) {
      ui.showError('Room not found');
      state.pendingJoinCode = null;
      history.replaceState(null, '', '/');
      ui.showHome(state.weaponDefs);
      return;
    }
    ui.showInviteJoining(code);
    send({ type: 'join_room', code, weapon: state.myWeapon });
  } catch {
    ui.showError('Failed to check room');
    state.pendingJoinCode = null;
    ui.showHome(state.weaponDefs);
  }
}

function handleMessage(msg) {
  switch (msg.type) {

    case 'hello':
      setWireEncoding(msg.encoding || 'json');
      state.gamePhase = 'home';
      state.weaponDefs = msg.weapons;
      state.availableWeapons = Object.keys(msg.weapons);

      if (state.pendingJoinCode) {
        void tryJoinFromURL();
      } else {
        ui.showHome(msg.weapons);
      }
      break;

    case 'encoding_changed':
      setWireEncoding(msg.encoding || 'json');
      break;

    case 'room_created':
      state.myId = msg.playerId;
      state.myWeapon = msg.weapon;
      state.isHost = true;
      state.isSpectator = false;
      state.pendingJoinCode = null;
      state.currentRoom = { code: msg.code, inviteUrl: msg.inviteUrl };
      state.gamePhase = 'lobby';
      history.replaceState(null, '', `/play/${msg.code}`);
      ui.showRoomLobby();
      break;

    case 'room_joined':
      state.myId = msg.playerId;
      state.myWeapon = msg.weapon;
      state.isSpectator = false;
      state.pendingJoinCode = null;
      state.currentRoom = { code: msg.code };
      state.gamePhase = 'lobby';
      history.replaceState(null, '', `/play/${msg.code}`);
      ui.showRoomLobby();
      break;

    case 'room_spectated':
      state.myId = msg.playerId;
      state.isHost = false;
      state.isSpectator = true;
      state.pendingJoinCode = null;
      state.currentRoom = { code: msg.code, inviteUrl: msg.inviteUrl };
      state.gamePhase = 'lobby';
      history.replaceState(null, '', `/play/${msg.code}`);
      ui.showRoomLobby();
      break;

    case 'room_state':
      state.roomState = msg;
      state.isHost = (msg.hostId === state.myId);
      if (state.myId) {
        const me = msg.players.find(p => p.id === state.myId);
        if (me) state.myName = me.name;
      }
      if (msg.phase === 'finished') {
        if (state.matchResults) ui.showMatchResults(state.matchResults);
        else ui.updateRoomLobby(msg);
      } else if (state.gamePhase === 'in-game' && msg.phase === 'playing') {
        ui.updateReconnectOverlay(msg);
      } else {
        ui.updateRoomLobby(msg);
      }
      break;

    case 'room_closed':
      ui.showRoomClosed(msg.message || 'Room closed');
      break;

    case 'room_left':
      ui.goHome();
      break;

    case 'player_left':
      ui.showPlayerLeft(msg.playerId);
      break;

    case 'room_config_update':
      if (state.roomState) {
        state.roomState.config = msg.config;
        ui.updateRoomConfig(msg.config);
      }
      break;

    case 'match_start':
      state.gamePhase = 'in-game';
      state.pendingJoinCode = null;
      document.getElementById('match-over-screen')?.remove();
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
      break;

    case 'match_results':
      state.matchResults = msg;
      ui.showMatchResults(msg);
      break;

    case 'pong':
      state.currentPing = Math.round(performance.now() - state.pingStart);
      ui.updateWireEncoding(state.wireEncoding, state.currentPing);
      break;

    case 'error':
      if (state.pendingJoinCode) {
        state.pendingJoinCode = null;
        ui.showHome(state.weaponDefs);
      }
      ui.showError(msg.message);
      break;
  }
}
