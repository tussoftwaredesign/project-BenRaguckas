from rpc_tester import RpcTest

QUEUE_NAME = "request_queue"
HOST = "localhost"
PORT = 5672
test_rc = RpcTest("test_queue", HOST, PORT)
test_rc.begin()