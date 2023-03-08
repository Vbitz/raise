home = remote.info()["home"]
local_content = client.read_file("build/ra")
remote.write_file(join(home, "testing_ra"), local_content)
