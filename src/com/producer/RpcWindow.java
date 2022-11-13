package com.producer;

import javax.imageio.ImageIO;
import javax.swing.*;
import javax.xml.crypto.Data;
import java.awt.*;
import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.awt.image.BufferedImage;
import java.awt.image.DataBufferByte;
import java.io.ByteArrayInputStream;
import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.util.concurrent.ExecutionException;

public class RpcWindow extends JFrame implements ActionListener {
    private final ImagePanel image_pane = new ImagePanel();
    private final JButton select_button = new JButton("Select file");
    private final JButton test_button = new JButton("TEST");
    private final JFileChooser file_choose = new JFileChooser("./");

    public RpcWindow() {
        BorderLayout border_layout = new BorderLayout();
        setLayout(border_layout);
        add(image_pane, BorderLayout.CENTER);
        image_pane.setBackground(Color.LIGHT_GRAY);
        add(select_button, BorderLayout.SOUTH);
        select_button.addActionListener(this);
        add(test_button, BorderLayout.NORTH);
        test_button.addActionListener(this);

        setSize(new Dimension(1080, 720));
        setResizable(false);
        setVisible(true);
        setDefaultCloseOperation(EXIT_ON_CLOSE);
    }

    @Override
    public void actionPerformed(ActionEvent e) {
        try {
            if (e.getSource() == select_button) pickFile(file_choose.showOpenDialog(this));
            if (e.getSource() == test_button) test();
        } catch (Exception exc) {
            System.err.println(exc.getMessage());
        }
    }

    private void pickFile(int choice) throws IOException, ExecutionException, InterruptedException {
        if (choice == JFileChooser.APPROVE_OPTION){
//            BufferedImage image = ImageIO.read(file_choose.getSelectedFile());
//            DataBufferByte buffer = (DataBufferByte) image.getRaster().getDataBuffer();
//            byte[] data = buffer.getData();
            displayImage(ImageIO.read(file_choose.getSelectedFile()));
            byte[] data = Files.readAllBytes(file_choose.getSelectedFile().toPath());

            RpcClient rc = new RpcClient(this);
            byte[] response = rc.sendMessage(data);
            rc.close();

            ByteArrayInputStream stream = new ByteArrayInputStream(response);
            image_pane.setBackground(Color.DARK_GRAY);
            displayImage(ImageIO.read(stream));
        }
    }

    private void test() throws IOException, ExecutionException, InterruptedException {
        System.out.println(" [x] Requesting fibonacci 4");
        RpcClient rc = new RpcClient(this);
        for (int i = 0; i < 4; i++) {
            String i_str = Integer.toString(i);
            String response = rc.test_con(i_str);
            System.out.println(" [.] Got '" + response + "'");
        }
        rc.close();
    }

    private void displayImage(BufferedImage img) {
        image_pane.image = img;
        image_pane.update(image_pane.getGraphics());
    }
}
