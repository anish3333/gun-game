import { state, getBaseURL, COLORS } from './state.js';
import { send } from './network.js';

const overlay       = document.getElementById('overlay');
const overlayStatus = document.getElementById('overlay-status');
const homeMenu      = document.getElementById('home-menu');
const joinMenu      = document.getElementById('join-menu');
const createMenu    = document.getElementById('create-room-menu');
const roomLobby     = document.getElementById('room-lobby');
const gameWrap      = document.getElementById('game-wrap');

const screens = [homeMenu, joinMenu, createMenu, roomLobby];

function hideAllScreens() {
  screens.forEach(s => s?.classList.remove('active'));
  gameWrap.style.display = 'none';
}

export function initUI() {
  gameWrap.style.display = 'none';

  document.getElementById('btn-play-online')?.addEventListener('click', () => {
    clearError();
    showJoinScreen(state.weaponDefs);
  });

  document.getElementById('btn-create-room')?.addEventListener('click', () => {
    clearError();
    showCreateRoom(state.weaponDefs);
  });

  document.getElementById('btn-matchmaking')?.addEventListener('click', () => {
    showError('Matchmaking coming soon');
  });

  document.getElementById('btn-back-home')?.addEventListener('click', showHomeFromSubmenu);
  document.getElementById('btn-back-home-create')?.addEventListener('click', showHomeFromSubmenu);

  document.getElementById('join-room-btn')?.addEventListener('click', () => {
    clearError();
    const code = document.getElementById('room-code-input').value.toUpperCase().trim();
    if (code.length !== 6) { showError('Room code must be 6 characters'); return; }
    send({ type: 'join_room', code, weapon: state.myWeapon });
  });

  document.getElementById('spectate-room-btn')?.addEventListener('click', () => {
    clearError();
    const code = document.getElementById('room-code-input').value.toUpperCase().trim();
    if (code.length !== 6) { showError('Room code must be 6 characters'); return; }
    document.getElementById('join-status').textContent = 'Opening spectator view...';
    send({ type: 'spectate_room', code });
  });

  document.getElementById('room-code-input')?.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') document.getElementById('join-room-btn').click();
  });

  document.getElementById('room-code-input')?.addEventListener('input', (e) => {
    e.target.value = e.target.value.toUpperCase();
  });

  document.getElementById('create-room-submit')?.addEventListener('click', () => {
    clearError();
    send({
      type: 'create_room',
      weapon: state.myWeapon,
      scoreLimit: parseInt(document.getElementById('cfg-score').value, 10),
      timeLimit: parseInt(document.getElementById('cfg-time').value, 10),
      weaponMode: document.getElementById('cfg-weapon-mode').value,
      map: document.getElementById('cfg-map').value,
      botEnabled: document.getElementById('cfg-bot-enabled').checked,
    });
  });

  document.getElementById('copy-invite-btn')?.addEventListener('click', () => {
    const url = state.roomState?.inviteUrl || `${getBaseURL()}/play/${state.currentRoom?.code}`;
    navigator.clipboard.writeText(url).then(() => showError('Invite link copied!'));
  });

  document.getElementById('start-match-btn')?.addEventListener('click', () => {
    send({ type: 'start_match' });
  });

  document.getElementById('leave-room-btn')?.addEventListener('click', () => {
    send({ type: 'leave_room' });
    // Optimistic — server confirms with room_left
    goHome();
  });

  document.getElementById('host-score-limit')?.addEventListener('change', (e) => {
    if (!state.isHost) return;
    send({ type: 'update_room_config', scoreLimit: parseInt(e.target.value, 10) });
  });
}

function showHomeFromSubmenu() {
  clearError();
  showHome(state.weaponDefs);
}

export function showConnecting() {
  overlay.style.display = 'flex';
  overlayStatus.textContent = 'connecting to server...';
  hideAllScreens();
}

export function showInviteJoining(code) {
  overlay.style.display = 'flex';
  overlayStatus.textContent = `joining room ${code}...`;
  hideAllScreens();
}

export function showHome(weapons) {
  overlay.style.display = 'none';
  hideAllScreens();
  homeMenu.classList.add('active');
  state.gamePhase = 'home';
  if (weapons) renderWeaponGrid('weapon-grid-home', weapons);
}

export function showJoinScreen(weapons) {
  overlay.style.display = 'none';
  hideAllScreens();
  joinMenu.classList.add('active');
  state.gamePhase = 'join';
  if (weapons) renderWeaponGrid('weapon-grid-join', weapons);
}

export function showCreateRoom(weapons) {
  overlay.style.display = 'none';
  hideAllScreens();
  createMenu.classList.add('active');
  state.gamePhase = 'create';
  if (weapons) renderWeaponGrid('weapon-grid-create', weapons);
}

export function showJoining(code) {
  overlay.style.display = 'none';
  hideAllScreens();
  joinMenu.classList.add('active');
  document.getElementById('room-code-input').value = code;
  document.getElementById('join-status').textContent = 'Joining room...';
}

export function showRoomLobby() {
  overlay.style.display = 'none';
  hideAllScreens();
  roomLobby.classList.add('active');
  gameWrap.style.display = 'none';
  state.gamePhase = 'lobby';
}

export function updateRoomLobby(roomState) {
  state.roomState = roomState;
  state.isHost = roomState.hostId === state.myId;

  document.getElementById('lobby-room-code').textContent = roomState.code;
  document.getElementById('lobby-invite-url').textContent = roomState.inviteUrl || `${getBaseURL()}/play/${roomState.code}`;

  const list = document.getElementById('lobby-players');
  list.innerHTML = '';

  const connected = roomState.players.filter(p => p.connected).length;
  const spectatorCount = roomState.spectatorCount || 0;

  roomState.players.forEach(p => {
    const row = document.createElement('div');
    row.className = 'lobby-player-row';

    let status = '✓';
    let suffix = '';
    if (!p.connected) {
      status = '⌛';
      suffix = p.reconnectSeconds > 0
        ? `reconnecting... (${p.reconnectSeconds}s)`
        : 'waiting...';
    } else if (p.isHost) {
      suffix = 'HOST';
    }

    row.innerHTML = `
      <span class="lobby-p-status">${status}</span>
      <span class="lobby-p-name">${p.name || p.id.slice(0, 8)}</span>
      <span class="lobby-p-tag">${suffix}</span>
    `;
    list.appendChild(row);
  });

  if (connected < 2) {
    const waiting = document.createElement('div');
    waiting.className = 'lobby-player-row dim';
    waiting.innerHTML = '<span class="lobby-p-status">⌛</span><span class="lobby-p-name">Waiting...</span>';
    list.appendChild(waiting);
  }

  updateRoomConfig(roomState.config);
  document.getElementById('lobby-spectators').textContent =
    spectatorCount === 1 ? '1 spectator' : `${spectatorCount} spectators`;

  if (connected >= 2) {
    state.playerLeftNotice = null;
  }

  const startBtn = document.getElementById('start-match-btn');
  const lobbyAction = document.getElementById('lobby-action-text');
  const hostPanel = document.getElementById('host-controls');
  const finishedPanel = document.getElementById('finished-controls');

  startBtn.style.display = 'none';
  lobbyAction.style.display = 'none';
  hostPanel.style.display = 'none';
  finishedPanel.style.display = 'none';
  document.getElementById('match-over-screen')?.remove();

  if (roomState.phase === 'lobby') {
    state.gamePhase = 'lobby';
    if (state.isHost) {
      hostPanel.style.display = 'block';
      startBtn.style.display = 'inline-block';
      startBtn.disabled = connected < 2;
      startBtn.textContent = 'START MATCH';
      if (state.playerLeftNotice) {
        lobbyAction.style.display = 'block';
        lobbyAction.textContent = state.playerLeftNotice;
      }
    } else {
      lobbyAction.style.display = 'block';
      lobbyAction.textContent = state.isSpectator ? 'Spectating' : 'Waiting for host...';
    }
  } else if (roomState.phase === 'finished') {
    if (state.matchResults) {
      showMatchResults(state.matchResults);
    } else {
      state.gamePhase = 'finished';
      gameWrap.style.display = 'none';
      roomLobby.classList.add('active');
    }
  } else if (roomState.phase === 'playing') {
    // mid-game reconnect handled by match_start
  }
}

export function updateRoomConfig(config) {
  document.getElementById('lobby-score-limit').textContent = config.scoreLimit;
  document.getElementById('lobby-time-limit').textContent = config.timeLimit + 'm';
  document.getElementById('lobby-weapon-mode').textContent = config.weaponMode;
  document.getElementById('lobby-map').textContent = config.map;

  const hostScore = document.getElementById('host-score-limit');
  if (hostScore && state.isHost && roomLobby.classList.contains('active')) {
    hostScore.value = config.scoreLimit;
  }
}

export function showMatchStart(players) {
  overlay.style.display = 'none';
  hideAllScreens();
  gameWrap.style.display = 'block';
  state.gamePhase = 'in-game';
  document.getElementById('spectator-badge').style.display = state.isSpectator ? 'block' : 'none';
  document.getElementById('aim-hint').textContent = state.isSpectator
    ? 'WATCHING LIVE MATCH'
    : 'MOVE MOUSE TO AIM · HOLD CLICK TO FIRE';

  players.forEach((p, index) => {
    const color = COLORS[p.weaponType] || '#4af0c8';
    const isPrimaryHud = (!state.isSpectator && p.id === state.myId) || (state.isSpectator && index === 0);
    if (isPrimaryHud) {
      document.getElementById('you-name').textContent = state.isSpectator ? 'PLAYER 1' : 'YOU';
      document.getElementById('you-weapon').textContent = p.label || p.weaponType || '—';
      document.querySelector('.player-hud.you').style.borderTopColor = color;
      document.querySelector('.hp-fill-you').style.background = color;
      document.querySelector('.hp-val.you-val').style.color = color;
      document.getElementById('score-you').style.color = color;
    } else if (!state.isSpectator || index === 1) {
      document.getElementById('opp-name').textContent = state.isSpectator ? 'PLAYER 2' : 'OPPONENT';
      document.getElementById('opp-weapon').textContent = p.label || p.weaponType;
      document.querySelector('.player-hud.opponent').style.borderTopColor = color;
      document.querySelector('.hp-fill-opp').style.background = color;
      document.querySelector('.hp-val.opp').style.color = color;
      document.getElementById('score-opp').style.color = color;
    }
  });
}

export function showMatchResults(msg) {
  state.gamePhase = 'finished';
  state.matchResults = msg;

  document.getElementById('match-over-screen')?.remove();
  document.getElementById('reconnect-overlay')?.remove();

  overlay.style.display = 'none';
  hideAllScreens();
  roomLobby.classList.add('active');
  gameWrap.style.display = 'none';

  const myStats = msg.stats[state.myId] || { kills: 0, deaths: 0 };
  const won = !state.isSpectator && msg.winnerId === state.myId;
  const winner = state.roomState?.players.find(p => p.id === msg.winnerId);

  const panel = document.getElementById('finished-controls');
  panel.style.display = 'block';
  panel.innerHTML = `
    <div class="finished-title">${state.isSpectator ? 'Match Over' : (won ? 'Victory!' : 'Defeat')}</div>
    <div class="finished-winner">Winner: ${winner?.name || msg.winnerId.slice(0, 8)}</div>
    <div class="finished-stats">${state.isSpectator ? 'Spectating' : `Kills: ${myStats.kills} · Deaths: ${myStats.deaths}`}</div>
    <div class="finished-actions">
      ${state.isSpectator ? '' : '<button class="action-btn" id="rematch-btn">REMATCH</button>'}
      <button class="action-btn" id="leave-finished-btn">LEAVE</button>
    </div>
    <div id="rematch-status" class="lobby-action-text"></div>
  `;

  document.getElementById('rematch-btn')?.addEventListener('click', () => {
    send({ type: 'rematch' });
    document.getElementById('rematch-btn').textContent = 'REMATCH ✓';
    document.getElementById('rematch-btn').disabled = true;
  });

  document.getElementById('leave-finished-btn').onclick = () => {
    send({ type: 'leave_room' });
    goHome();
  };

  if (state.isSpectator) {
    document.getElementById('rematch-status').textContent = 'Spectating';
  } else if (state.isHost) {
    document.getElementById('rematch-status').textContent = 'Start new match after both vote rematch';
  } else {
    document.getElementById('rematch-status').textContent = 'Waiting for host...';
  }

  if (state.roomState) {
    const votes = state.roomState.players.filter(p => p.rematchVote).length;
    const total = state.roomState.players.filter(p => p.connected).length;
    if (votes > 0) {
      document.getElementById('rematch-status').textContent =
        `Rematch votes: ${votes}/${total}`;
    }
  }
}

export function goHome() {
  state.currentRoom = null;
  state.roomState = null;
  state.matchResults = null;
  state.playerLeftNotice = null;
  state.isSpectator = false;
  document.getElementById('spectator-badge').style.display = 'none';
  document.getElementById('aim-hint').textContent = 'MOVE MOUSE TO AIM · HOLD CLICK TO FIRE';
  state.gamePhase = 'home';
  history.replaceState(null, '', '/');
  showHome(state.weaponDefs);
}

export function showRoomClosed(message) {
  goHome();
  showError(message);
}

export function showPlayerLeft(playerId) {
  const name = state.roomState?.players.find(p => p.id === playerId)?.name
    || playerId.slice(0, 8);
  state.playerLeftNotice = `${name} left the room`;
}

export function showDisconnected(reason) {
  gameWrap.style.display = 'none';
  screens.forEach(s => s?.classList.remove('active'));
  overlay.style.display = 'flex';
  overlayStatus.textContent = reason || 'disconnected';
}

export function renderWeaponGrid(gridId, weapons) {
  const grid = document.getElementById(gridId);
  if (!grid) return;
  grid.innerHTML = '';

  Object.entries(weapons).forEach(([type, def]) => {
    const btn = document.createElement('button');
    btn.className = 'weapon-btn' + (type === state.myWeapon ? ' selected' : '');
    btn.dataset.weapon = type;

    const color = COLORS[type] || '#4af0c8';
    btn.innerHTML = `
      <div class="wbtn-name" style="color:${color}">${def.label}</div>
      <div class="wbtn-desc">${def.desc}</div>
    `;
    btn.onclick = () => {
      state.myWeapon = type;
      document.querySelectorAll(`#${gridId} .weapon-btn`).forEach(b => b.classList.remove('selected'));
      btn.classList.add('selected');
    };
    grid.appendChild(btn);
  });
}

export function updateReconnectOverlay(roomState) {
  let el = document.getElementById('reconnect-overlay');
  const disconnected = roomState.players.filter(p => !p.connected);

  if (disconnected.length === 0) {
    el?.remove();
    return;
  }

  if (!el) {
    el = document.createElement('div');
    el.id = 'reconnect-overlay';
    el.style.cssText = 'position:absolute;top:60px;left:50%;transform:translateX(-50%);background:#080c12ee;border:1px solid #1a2535;padding:12px 20px;border-radius:6px;font-size:11px;z-index:50;text-align:center;';
    gameWrap.appendChild(el);
  }

  const p = disconnected[0];
  const secs = p.reconnectSeconds > 0 ? ` · expires in ${p.reconnectSeconds}s` : '';
  el.innerHTML = `<div>${p.name || 'Opponent'} reconnecting...${secs}</div>`;
}

export function updateWireEncoding(encoding, ping) {
  const el = document.getElementById('ping-display');
  if (!el) return;
  const pingText = ping != null ? `ping: ${ping}ms` : 'ping: —';
  el.textContent = `${pingText} · ${encoding}`;
}

export function updateHUD() {
  const me  = state.isSpectator ? state.gameState.players[0] : state.gameState.players.find(p => p.id === state.myId);
  const opp = state.isSpectator ? state.gameState.players[1] : state.gameState.players.find(p => p.id !== state.myId);

  if (me) {
    const hp = Math.max(0, me.hp);
    document.getElementById('hp-you').style.width  = hp + '%';
    document.getElementById('hp-you-val').textContent = hp;
    document.getElementById('score-you').textContent  = me.score;
  }
  if (opp) {
    const hp = Math.max(0, opp.hp);
    document.getElementById('hp-opp').style.width  = hp + '%';
    document.getElementById('hp-opp-val').textContent = hp;
    document.getElementById('score-opp').textContent  = opp.score;
  }
}

export function showError(text) {
  const c = document.getElementById('error-container');
  if (!c) return;
  c.innerHTML = `<div class="error-msg">${text}</div>`;
  setTimeout(() => { c.innerHTML = ''; }, 4000);
}

function clearError() {
  const c = document.getElementById('error-container');
  if (c) c.innerHTML = '';
}
