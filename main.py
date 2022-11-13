from rpc_example import RpcExample


if __name__ == '__main__':
    QUEUE_NAME = "image_queue"
    HOST = "localhost"
    PORT = 5672

    main_rc = RpcExample(QUEUE_NAME, HOST, PORT)
    main_rc.begin()


