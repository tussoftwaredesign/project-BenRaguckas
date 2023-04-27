import pika
from rembg import remove
import urllib.request


class RpcExample:
    connection = None
    channel = None
    queue = None
    response_method = None

    def __init__(self, que_name, host, port, test=False):
        if test:
            self.response_method = self.test_call
        else:
            self.response_method = self.img_call
        self.connection = pika.BlockingConnection(pika.ConnectionParameters(host=host, port=port))
        self.channel = self.connection.channel()
        self.channel.queue_declare(queue=que_name)
        self.queue = que_name

    def img_call(self, ch, method, props, body):
        print(f"Got message.\t"
              f"Correlation id: {props.correlation_id}\t"
              f"Reply to: {props.reply_to}\t"
              f"Method: {method.delivery_tag}\t")
        processed_image = remove(body)


        print("\tImage processed, replying.")

        ch.basic_publish(exchange='',
                         routing_key=props.reply_to,
                         properties=pika.BasicProperties(correlation_id=props.correlation_id),
                         body=processed_image)
        ch.basic_ack(delivery_tag=method.delivery_tag)

    def begin(self):
        self.channel.basic_qos(prefetch_count=1)
        self.channel.basic_consume(queue=self.queue, on_message_callback=self.response_method)

        print(" [x] Awaiting RPC requests")
        self.channel.start_consuming()

