home = remote.info()["home"]
print(home, join(home, "testing.txt"))
print(remote.read_file(join(home, "testing.txt")))
