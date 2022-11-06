package com.producer;

import com.rabbitmq.client.Channel;
import com.rabbitmq.client.Connection;
import com.rabbitmq.client.ConnectionFactory;

import java.io.IOException;
import java.time.LocalDateTime;
import java.util.concurrent.TimeoutException;

public class Producer {
    private final static String QUEUE_NAME = "test_queue";
    private final static String HOST = "localhost";
    private final static int PORT = 5672;

    private final static String MESSAGE = "Producer message ";

    public static void main(String[] args) {
        ConnectionFactory connectionFactory = new ConnectionFactory();
        connectionFactory.setHost(HOST);
        connectionFactory.setPort(PORT);

        try {
            Connection connection = connectionFactory.newConnection();

            Channel channel = connection.createChannel();
            channel.queueDeclare(QUEUE_NAME, false, false, false, null);

            String message = MESSAGE + LocalDateTime.now();
            channel.basicPublish("", QUEUE_NAME, null, message.getBytes());

            System.out.println("Message queued successfully.");

            channel.close();
            connection.close();
        } catch (IOException | TimeoutException e) {
            throw new RuntimeException(e);
        }

    }
}
