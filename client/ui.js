import { state, getPlayer, COLORS } from './state.js';
import { send } from './network.js';

// ── screen refs ───────────────────────────────────────────────────────────
const overlay     = document.getElementById('overlay');
const overlayStatus = document.getElementById('overlay-status');
const lobbyMenu   = document.getElementById('lobby-menu');
const roomDisplay = document.getElementById('room-display');
const gameWrap    = document.getElementById('game-wrap');

export function initUI() {
  // Hide game area until match starts
  gameWrap.style.display = 'none';

  document.getElementById('create-room-btn').onclick = () => {
    clearError();
    send({ type: 'create_room', weapon: state.myWeapon });
  };

  document.getElementById('join-room-btn').onclick = () => {
    clearError();
    const code = document.getElementById('room-code-input').value.toUpperCase().trim();
    if (code.length !== 4) { showError('Room code must be 4 characters'); return; }
    send({ type: 'join_room', code, weapon: state.myWeapon });
  };

  document.getElementById('room-code-input').addEventListener('keydown', (e) => {
    if (e.key === 'Enter') document.getElementById('join-room-btn').click();
  });

  // Auto-uppercase room code input
  document.getElementById('room-code-input').addEventListener('input', (e) => {
    e.target.value = e.target.value.toUpperCase();
  });
}

// ── screen transitions ────────────────────────────────────────────────────

export function showConnecting() {
  overlay.style.display = 'flex';
  overlayStatus.textContent = 'connecting to server...';
  lobbyMenu.classList.remove('active');
  roomDisplay.classList.remove('active');
  gameWrap.style.display = 'none';
}

export function showLobby(weapons, rooms) {
  overlay.style.display = 'none';
  lobbyMenu.classList.add('active');
  roomDisplay.classList.remove('active');
  gameWrap.style.display = 'none';
  renderWeaponGrid(weapons);
  updateRoomList(rooms);
}

export function showRoomWaiting(code) {
  lobbyMenu.classList.remove('active');
  roomDisplay.classList.add('active');
  gameWrap.style.display = 'none';
  document.getElementById('room-code-display').textContent = code;
  document.getElementById('waiting-status').textContent = 'share the code · waiting for opponent...';
}

export function showMatchStart(players) {
  overlay.style.display = 'none';
  lobbyMenu.classList.remove('active');
  roomDisplay.classList.remove('active');
  gameWrap.style.display = 'block';

  players.forEach(p => {
    const color = COLORS[p.weaponType] || '#4af0c8';
    if (p.id === state.myId) {
      document.getElementById('you-name').textContent = 'YOU';
      document.getElementById('you-weapon').textContent = (p.label || p.weaponType) + ' · click to fire';
      document.querySelector('.player-hud.you').style.borderTopColor = color;
      document.querySelector('.hp-fill-you').style.background = color;
      document.querySelector('.hp-val.you-val').style.color = color;
      document.getElementById('score-you').style.color = color;
    } else {
      document.getElementById('opp-name').textContent = 'OPPONENT';
      document.getElementById('opp-weapon').textContent = p.label || p.weaponType;
      document.querySelector('.player-hud.opponent').style.borderTopColor = color;
      document.querySelector('.hp-fill-opp').style.background = color;
      document.querySelector('.hp-val.opp').style.color = color;
      document.getElementById('score-opp').style.color = color;
    }
  });
}

export function showDisconnected(reason) {
  gameWrap.style.display = 'none';
  lobbyMenu.classList.remove('active');
  roomDisplay.classList.remove('active');
  overlay.style.display = 'flex';
  overlayStatus.textContent = reason || 'disconnected';
}

// ── weapon grid ───────────────────────────────────────────────────────────

export function renderWeaponGrid(weapons) {
  const grid = document.getElementById('weapon-grid');
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
      document.querySelectorAll('.weapon-btn').forEach(b => b.classList.remove('selected'));
      btn.classList.add('selected');
    };
    grid.appendChild(btn);
  });
}

// ── room list ─────────────────────────────────────────────────────────────

export function updateRoomList(rooms) {
  const list = document.getElementById('room-list');
  if (!rooms || rooms.length === 0) {
    list.innerHTML = '<div class="no-rooms">no open rooms · create one above</div>';
    return;
  }
  list.innerHTML = '';
  rooms.forEach(room => {
    const item = document.createElement('div');
    item.className = 'room-item';
    item.innerHTML = `
      <span class="room-item-code">${room.code}</span>
      <span class="room-item-status">${room.players}/2 players</span>
      <span class="room-item-join">JOIN →</span>
    `;
    item.onclick = () => {
      clearError();
      send({ type: 'join_room', code: room.code, weapon: state.myWeapon });
    };
    list.appendChild(item);
  });
}

// ── HUD ───────────────────────────────────────────────────────────────────

export function updateHUD() {
  const me  = getPlayer(state.myId);
  const opp = state.gameState.players.find(p => p.id !== state.myId);

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

// ── error messages ────────────────────────────────────────────────────────

export function showError(text) {
  const c = document.getElementById('error-container');
  c.innerHTML = `<div class="error-msg">${text}</div>`;
  setTimeout(() => { c.innerHTML = ''; }, 4000);
}

function clearError() {
  document.getElementById('error-container').innerHTML = '';
}