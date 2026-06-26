from env import GunGameEnv

env = GunGameEnv()

obs, _ = env.reset()

for i in range(100):
    action = env.action_space.sample()
    obs, reward, done, _, _ = env.step(action)

    print(reward)

    if done:
        obs, _ = env.reset()