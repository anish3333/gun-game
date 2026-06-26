from stable_baselines3 import PPO
from env import GunGameEnv

env = GunGameEnv()

model = PPO(
    "MlpPolicy",
    env,
    verbose=1,
    learning_rate=3e-4,
    n_steps=2048,
    batch_size=64,
    gamma=0.99,
)

model.learn(total_timesteps=100_000)

model.save("ppo_gungame_100k")