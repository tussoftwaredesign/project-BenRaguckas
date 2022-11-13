package com.producer;
import com.rabbitmq.client.AMQP;
import com.rabbitmq.client.Channel;
import com.rabbitmq.client.Connection;
import com.rabbitmq.client.ConnectionFactory;

import java.io.IOException;
import java.util.UUID;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.ExecutionException;

public class RpcClient implements AutoCloseable {
    //  Some final variables
    private final static String QUEUE_NAME = "image_queue";
    private final static String HOST = "localhost";
    private final static int PORT = 5672;
    final RpcWindow window;


    private Connection connection;
    private Channel channel;

    public RpcClient(RpcWindow window){
        this.window = window;
        ConnectionFactory connectionFactory = new ConnectionFactory();
        connectionFactory.setHost(HOST);
        connectionFactory.setPort(PORT);

        try {
            connection = connectionFactory.newConnection();
            channel = connection.createChannel();
        } catch (Exception e) {
            System.err.println(e.getMessage());
        }
    }

    public byte[] sendMessage(byte[] message_bytes) throws IOException, ExecutionException, InterruptedException {
        final String CORRELATION_ID = UUID.randomUUID().toString();

        String replyQueueName = channel.queueDeclare().getQueue();
        AMQP.BasicProperties props = new AMQP.BasicProperties
                .Builder()
                .correlationId(CORRELATION_ID)
                .replyTo(replyQueueName)
                .build();

        channel.basicPublish("", QUEUE_NAME, props, message_bytes);

        final CompletableFuture<byte[]> response = new CompletableFuture<>();

        String ctag = channel.basicConsume(replyQueueName, true, (consumerTag, delivery) -> {
            if (delivery.getProperties().getCorrelationId().equals(CORRELATION_ID)) {
                response.complete(delivery.getBody());
            }
        }, consumerTag -> {
        });

        byte[] result = response.get();
        channel.basicCancel(ctag);
        return result;
    }

    //  Default example test rpc
    public String test_con(String message) throws IOException, InterruptedException, ExecutionException {
        final String corrId = UUID.randomUUID().toString();

        String replyQueueName = channel.queueDeclare().getQueue();
        AMQP.BasicProperties props = new AMQP.BasicProperties
                .Builder()
                .correlationId(corrId)
                .replyTo(replyQueueName)
                .build();

        channel.basicPublish("", "test_queue", props, message.getBytes("UTF-8"));

        final CompletableFuture<String> response = new CompletableFuture<>();

        String ctag = channel.basicConsume(replyQueueName, true, (consumerTag, delivery) -> {
            if (delivery.getProperties().getCorrelationId().equals(corrId)) {
                response.complete(new String(delivery.getBody(), "UTF-8"));
            }
        }, consumerTag -> {
        });

        String result = response.get();
        channel.basicCancel(ctag);
        return result;
    }

    public void close() throws IOException {
        connection.close();
    }

}
