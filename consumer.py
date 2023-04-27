import json
import pika
import requests


class Consumer:
    connection = None
    channel = None
    queue_name = None
    api_host = None
    api_port = None
    rmq_host = None
    rmq_port = None

    def __init__(self, que_name, api_host, rmq_host, rmq_port, function):
        self.function = function
        self.api_host = api_host  # Host
        self.rmq_host = rmq_host  # Host
        self.rmq_port = rmq_port
        self.connection = pika.BlockingConnection(pika.ConnectionParameters(host=rmq_host, port=rmq_port))
        self.channel = self.connection.channel()
        self.channel.queue_declare(queue=que_name)
        self.queue_name = que_name

    def parseRMQBody(self, json_data):
        return {
            "Src": f'http://{self.api_host}{json_data["Src"]}',
            "Dst": f'http://{self.api_host}{json_data["Dst"]}',
            "Stat": f'http://{self.api_host}{json_data["Stat"]}',
            "Params": json_data["Params"],
        }

    def consumeProcess(self, ch, method, props, body):
        print(f"Got message.\t"
              f"Correlation id: {props.correlation_id}\t"
              f"Reply to: {props.reply_to}\t"
              f"Method: {method.delivery_tag}\t")
        comm = self.parseRMQBody(json.loads(body))
        print(comm)
        with requests.get(comm["Src"]) as f:
            update_res = requests.post(comm["Stat"], data={})
            print(f'Update: {update_res.status_code}:{update_res.text}')
            # The magic line for calling external functions
            try:
                result = self.function(f.content, comm["Params"])
                r = requests.put(comm["Dst"], files={'file': result}, data={'Content-Type': 'multipart/form-data'})
                print(r.status_code)
                print(r.text)
            except:
                requests.post(comm["Stat"], data={"status": "error"})
                print(f'Caught Error, consumed message and updated status.')
        ch.basic_ack(delivery_tag=method.delivery_tag)

    def begin(self):
        self.channel.basic_qos(prefetch_count=1)
        self.channel.basic_consume(queue=self.queue_name, on_message_callback=self.consumeProcess)

        print(" [x] Awaiting RPC requests")
        self.channel.start_consuming()

    def close(self):
        print("Closing the channel and connection...")
        if self.channel:
            self.channel.close()
        if self.connection:
            self.connection.close()
