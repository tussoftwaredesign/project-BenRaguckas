package com.producer;

import javax.swing.*;
import java.awt.*;
import java.awt.image.BufferedImage;

public class ImagePanel extends JPanel {
    BufferedImage image = null;
    public void paintComponent(Graphics g) {
        super.paintComponent(g);
        if (image!= null)
            g.drawImage(image, 0, 0, 640, 640, null);
    }
}
