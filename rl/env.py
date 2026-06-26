import subprocess
import json
import numpy as np
import gymnasium as gym
from gymnasium import spaces


class GunGameEnv(gym.Env):

    def __init__(self):
        super().__init__()


        self.proc = subprocess.Popen(
            ["go", "run", "cmd/rl/main.go"],
            cwd="../go-server",      # VERY IMPORTANT
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            text=True,
            bufsize=1,
        )

        # angle + shoot probability
        self.action_space = spaces.Box(
            low=np.array([-np.pi, 0], dtype=np.float32),
            high=np.array([np.pi, 1], dtype=np.float32),
            dtype=np.float32
        )

        # 16 floats from Observation()
        self.observation_space = spaces.Box(
            low=-np.inf,
            high=np.inf,
            shape=(16,),
            dtype=np.float32
        )

    def reset(self, seed=None, options=None):
        request = {
            "type": "reset"
        }

        self.proc.stdin.write(json.dumps(request) + "\n")
        self.proc.stdin.flush()

        response = json.loads(self.proc.stdout.readline())

        obs = np.array(response["observation"], dtype=np.float32)

        return obs, {}

    def step(self, action):

        angle = float(action[0])
        shoot = bool(action[1] > 0.5)

        request = {
            "angle": angle,
            "shoot": shoot
        }

        self.proc.stdin.write(json.dumps(request) + "\n")
        self.proc.stdin.flush()

        response = json.loads(self.proc.stdout.readline())

        obs = np.array(response["observation"], dtype=np.float32)
        reward = response["reward"]
        done = response["done"]

        return obs, reward, done, False, {}

    def close(self):
        self.proc.kill()