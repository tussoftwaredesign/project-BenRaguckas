import pika


def consumer(queue, host, port):
    connection = pika.BlockingConnection(pika.ConnectionParameters(host=host, port=port))
    channel = connection.channel()

    channel.queue_declare(queue=queue)

    def callback_print(ch, method, properties, body):
        print(" [x] Received %r" % body)

    channel.basic_consume(queue=queue, on_message_callback=callback_print, auto_ack=True)

    print('Awaiting for messages in queue:' + queue)
    channel.start_consuming()


if __name__ == '__main__':
    QUEUE_NAME = "test_queue"
    HOST = "localhost"
    PORT = 5672

    consumer(QUEUE_NAME, HOST, PORT)

