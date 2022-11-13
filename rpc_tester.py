import pika


class RpcTest:
    connection = None
    channel = None
    queue = None
    response_method = None

    def __init__(self, que_name, host, port):
        self.connection = pika.BlockingConnection(pika.ConnectionParameters(host=host, port=port))
        self.channel = self.connection.channel()
        self.channel.queue_declare(queue=que_name)
        self.queue = que_name

    def test_call(self, ch, method, props, body):
        print("Fib test")
        n = int(body)
        response = self.fib(n)

        ch.basic_publish(exchange='',
                         routing_key=props.reply_to,
                         properties=pika.BasicProperties(correlation_id=props.correlation_id),
                         body=str(response))
        ch.basic_ack(delivery_tag=method.delivery_tag)

    def begin(self):
        self.channel.basic_qos(prefetch_count=1)
        self.channel.basic_consume(queue=self.queue, on_message_callback=self.test_call)

        print(" [x] Awaiting RPC requests")
        self.channel.start_consuming()

    def fib(self, n):
        if n == 0:
            return 0
        elif n == 1:
            return 1
        else:
            return self.fib(n - 1) + self.fib(n - 2)
