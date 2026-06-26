from stable_baselines3 import PPO
from env import GunGameEnv

env = GunGameEnv()

model = PPO.load("ppo_gungame_100k.zip")

obs, _ = env.reset()

while True:
    action, _ = model.predict(obs)

    obs, reward, done, _, _ = env.step(action)

    if done:
        obs, _ = env.reset()