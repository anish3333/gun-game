export const ARENA  = { x: 30, y: 30, w: 620, h: 420 };
export const COLORS = { pistol: '#4af0c8', shotgun: '#f0a84a', smg: '#a84af0', sniper: '#f04a4a' };

export const state = {
  ws:              null,
  myId:            null,
  myName:          null,
  myWeapon:        'pistol',
  shooting:        false,
  mouseAngle:      0,
  gameState:       { players: [], bullets: [] },
  pingStart:       0,
  currentPing:     0,
  gamePhase:       'connecting',
  currentRoom:     null,
  roomState:       null,
  matchResults:    null,
  availableWeapons: [],
  weaponDefs:      {},
  pendingJoinCode: null,
  isHost:          false,
  isSpectator:     false,
  playerLeftNotice: null,
  wireEncoding:    'json',
};

export function getPlayer(id) {
  return state.gameState.players.find(p => p.id === id);
}

export function getBaseURL() {
  return window.location.origin;
}

export function getWSURL() {
  const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  return `${wsProtocol}//${window.location.host}`;
}

export function getInviteCodeFromPath() {
  const parts = window.location.pathname.split('/').filter(Boolean);
  if (parts[0] === 'play' && parts[1]) {
    return parts[1].toUpperCase();
  }
  return null;
}
