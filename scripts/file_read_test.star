home = remote.info().home
print(home)
print(remote.read_file(join(home, "testing.txt")))
