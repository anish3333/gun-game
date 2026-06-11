export const ARENA  = { x: 30, y: 30, w: 620, h: 420 };
export const COLORS = { pistol: '#4af0c8', shotgun: '#f0a84a', smg: '#a84af0', sniper: '#f04a4a' };

export const state = {
  ws:              null,
  myId:            null,
  myWeapon:        'pistol',
  shooting:        false,
  mouseAngle:      0,
  gameState:       { players: [], bullets: [] },
  pingStart:       0,
  currentPing:     0,
  gamePhase:       'connecting', // connecting | lobby | waiting | in-game
  currentRoom:     null,
  availableWeapons: [],
  weaponDefs:      {},
};

export function getPlayer(id) {
  return state.gameState.players.find(p => p.id === id);
}