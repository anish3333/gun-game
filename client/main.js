import { initUI }              from './ui.js';
import { initInput }           from './input.js';
import { initRenderer }        from './renderer.js';
import { connect, startNetworkLoops } from './network.js';

document.addEventListener('DOMContentLoaded', () => {
  initUI();
  initInput();
  initRenderer();
  connect();
  startNetworkLoops();
});