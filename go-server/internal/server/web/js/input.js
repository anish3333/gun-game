import { state, getPlayer } from './state.js';

export function initInput() {
  const canvas = document.getElementById('arena');

  canvas.addEventListener('mousemove', (e) => {
    const rect = canvas.getBoundingClientRect();
    const mx = e.clientX - rect.left;
    const my = e.clientY - rect.top;
    const me = getPlayer(state.myId);
    if (me) {
      state.mouseAngle = Math.atan2(my - me.y, mx - me.x);
    }
  });

  canvas.addEventListener('mousedown',  () => { state.shooting = true;  });
  canvas.addEventListener('mouseup',    () => { state.shooting = false; });
  canvas.addEventListener('mouseleave', () => { state.shooting = false; });
}