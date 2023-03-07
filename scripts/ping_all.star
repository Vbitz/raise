def main():
    for worker in client.get_workers():
        print(worker, client.remote(worker).ping())

main()
