import json
import sys
from pathlib import Path

from stable_baselines3 import PPO


ROOT = Path(__file__).parent

model = PPO.load(ROOT / "ppo_gungame_100k.zip")

while True:
    line = sys.stdin.readline()
    if not line:
        break

    obs = json.loads(line)

    action, _ = model.predict(obs, deterministic=True)

    response = {
        "angle": float(action[0]),
        "shoot": bool(action[1] > 0.5),
    }

    print(json.dumps(response), flush=True)
